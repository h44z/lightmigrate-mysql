package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"sync/atomic"

	"github.com/h44z/lightmigrate"
)

const advisoryLockIDSalt uint = 1486364155

type driver struct {
	client            *sql.DB
	cfg               *config
	reentrantLockFlag int32 // must be accessed by atomic.XXX functions!

	logger  lightmigrate.Logger
	verbose bool
}

// DriverOption is a function that can be used within the driver constructor to
// modify the driver object.
type DriverOption func(svc *driver)

// NewDriver instantiates a new MongoDB driver. A MongoDB client and the database name are required arguments.
// If you have migration file that contain multiple statements, ensure that the sql.DB was opened with
// the multiStatements=true parameter!
func NewDriver(client *sql.DB, database string, opts ...DriverOption) (lightmigrate.MigrationDriver, error) {
	if database == "" {
		return nil, ErrNoDatabaseName
	}

	if client == nil {
		return nil, ErrNoDatabaseClient
	}

	cfg := &config{
		DatabaseName:    database,
		MigrationsTable: DefaultMigrationsTable,
		Locking:         true,
	}

	d := &driver{
		client: client,
		cfg:    cfg,
	}

	for _, opt := range opts {
		opt(d)
	}

	err := d.prepareMigrationTable()
	if err != nil {
		return nil, err
	}

	return d, nil
}

// WithLogger sets the logging instance used by the driver.
func WithLogger(logger lightmigrate.Logger) DriverOption {
	return func(d *driver) {
		d.logger = logger
	}
}

// WithVerboseLogging sets the verbose flag of the driver.
func WithVerboseLogging(verbose bool) DriverOption {
	return func(d *driver) {
		d.verbose = verbose
	}
}

// WithMigrationTable allows to specify the name of the table that contains the migration state.
func WithMigrationTable(migrationTable string) DriverOption {
	return func(d *driver) {
		d.cfg.MigrationsTable = migrationTable
	}
}

// WithLocking can be used to configure the locking behaviour of the MongoDB migration driver.
func WithLocking(lockingEnabled bool) DriverOption {
	return func(d *driver) {
		d.cfg.Locking = lockingEnabled
	}
}

func (d *driver) Close() error {
	return nil // nothing to clean up
}

func (d *driver) Lock() error {
	if !d.cfg.Locking {
		return nil
	}

	// check if already locked, if not, lock
	if !atomic.CompareAndSwapInt32(&d.reentrantLockFlag, 0, 1) {
		return nil // no swap happened, already locked
	}

	lockKey := d.getLockingKey()
	query := "SELECT GET_LOCK(?, 5)" // 5 second timeout
	var success bool
	if err := d.client.QueryRowContext(context.Background(), query, lockKey).Scan(&success); err != nil {
		atomic.StoreInt32(&d.reentrantLockFlag, 0) // restore unlock flag
		return &lightmigrate.DriverError{OrigErr: err, Msg: "try lock failed", Query: []byte(query)}
	}

	if !success {
		atomic.StoreInt32(&d.reentrantLockFlag, 0) // restore unlock flag
		return ErrDatabaseLocked
	}

	return nil
}

func (d *driver) Unlock() error {
	if !d.cfg.Locking {
		return nil
	}

	// check if already unlocked, if not, unlock
	if !atomic.CompareAndSwapInt32(&d.reentrantLockFlag, 1, 0) {
		return nil // no swap happened, already unlocked
	}

	lockKey := d.getLockingKey()
	query := "SELECT RELEASE_LOCK(?)" // 5 second timeout
	if _, err := d.client.ExecContext(context.Background(), query, lockKey); err != nil {
		atomic.StoreInt32(&d.reentrantLockFlag, 1) // restore lock flag
		return &lightmigrate.DriverError{OrigErr: err, Msg: "release lock failed", Query: []byte(query)}
	}

	// NOTE: RELEASE_LOCK could return NULL or (or 0 if the code is changed),
	// in which case isLocked should be true until the timeout expires -- synchronizing
	// these states is likely not worth trying to do; reconsider the necessity of isLocked.

	return nil
}

func (d *driver) GetVersion() (version uint64, dirty bool, err error) {
	query := "SELECT version, dirty FROM `" + d.cfg.MigrationsTable + "` LIMIT 1"
	err = d.client.QueryRowContext(context.Background(), query).Scan(&version, &dirty)
	switch {
	case err == sql.ErrNoRows:
		return lightmigrate.NoMigrationVersion, false, nil

	case err != nil:
		return 0, false, &lightmigrate.DriverError{OrigErr: err, Msg: "failed to select version", Query: []byte(query)}
	default:
		return version, dirty, nil
	}
}

func (d *driver) SetVersion(version uint64, dirty bool) error {
	tx, err := d.client.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return &lightmigrate.DriverError{OrigErr: err, Msg: "transaction start failed"}
	}

	// Delete all entries in the migrations table.
	query := "DELETE FROM `" + d.cfg.MigrationsTable + "`"
	if _, err := tx.ExecContext(context.Background(), query); err != nil {
		if errRollback := tx.Rollback(); errRollback != nil {
			origMsg := fmt.Sprintf("failed rollback for previous error: %v", err)
			return &lightmigrate.DriverError{OrigErr: err, Msg: origMsg, Query: []byte(query)}
		}
		return &lightmigrate.DriverError{OrigErr: err, Msg: "failed to clean migration table", Query: []byte(query)}
	}

	query = "INSERT INTO `" + d.cfg.MigrationsTable + "` (version, dirty) VALUES (?, ?)"
	if _, err := tx.ExecContext(context.Background(), query, version, dirty); err != nil {
		if errRollback := tx.Rollback(); errRollback != nil {
			origMsg := fmt.Sprintf("failed rollback for previous error: %v", err)
			return &lightmigrate.DriverError{OrigErr: err, Msg: origMsg, Query: []byte(query)}
		}
		return &lightmigrate.DriverError{OrigErr: err, Msg: "failed to update migration table", Query: []byte(query)}
	}

	if err := tx.Commit(); err != nil {
		return &lightmigrate.DriverError{OrigErr: err, Msg: "transaction commit failed"}
	}

	return nil
}

func (d *driver) RunMigration(migration io.Reader) error {
	migr, err := ioutil.ReadAll(migration)
	if err != nil {
		return err
	}

	query := string(migr[:]) // each line is a query
	if _, err := d.client.ExecContext(context.Background(), query); err != nil {
		return &lightmigrate.DriverError{OrigErr: err, Msg: "migration failed", Query: migr}
	}

	return nil
}

func (d *driver) Reset() error {
	// Delete all entries in the migrations table.
	query := "DROP TABLE IF EXISTS `" + d.cfg.MigrationsTable + "`"
	if _, err := d.client.ExecContext(context.Background(), query); err != nil {
		return &lightmigrate.DriverError{OrigErr: err, Msg: "failed drop migration table", Query: []byte(query)}
	}
	return nil
}

// Generate a unique locking key for the given database.
// The key will be derived from the database name only.
func (d *driver) getLockingKey() string {
	sum := crc32.ChecksumIEEE([]byte(d.cfg.DatabaseName))
	sum = sum * uint32(advisoryLockIDSalt)

	return fmt.Sprint(sum)
}

// prepareMigrationTable will create the migration table if it does not exist.
func (d *driver) prepareMigrationTable() (err error) {
	if err = d.Lock(); err != nil {
		return err
	}
	defer func() {
		if e := d.Unlock(); e != nil {
			if err == nil {
				err = e
			} else {
				err = fmt.Errorf("failed to unlock (%v) after: %v", e, err)
			}
		}
	}()

	query := "CREATE TABLE IF NOT EXISTS `" + d.cfg.MigrationsTable + "` (version bigint not null primary key, dirty boolean not null)"
	if _, err := d.client.ExecContext(context.Background(), query); err != nil {
		return &lightmigrate.DriverError{OrigErr: err, Msg: "failed create migration table", Query: []byte(query)}
	}
	return nil
}
