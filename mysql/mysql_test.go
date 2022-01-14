package mysql

import (
	"log"
	"testing"
)

func TestNewDriver(t *testing.T) {

}

func TestNewDriver_NoDb(t *testing.T) {
	_, err := NewDriver(nil, "")
	if err == nil {
		t.Fatalf("expected error, got: %v", err)
	}
}

func TestNewDriver_NoClient(t *testing.T) {
	_, err := NewDriver(nil, "db")
	if err == nil {
		t.Fatalf("expected error, got: %v", err)
	}
}

func TestWithLocking(t *testing.T) {
	d := &driver{cfg: &config{}}

	WithLocking(true)(d)
	if d.cfg.Locking != true {
		t.Fatalf("failed to set lock config")
	}
}

func TestWithLogger(t *testing.T) {
	d := &driver{}

	WithLogger(log.Default())(d)
	if d.logger != log.Default() {
		t.Fatalf("failed to set logger")
	}
}

func TestWithMigrationTable(t *testing.T) {
	d := &driver{cfg: &config{}}

	WithMigrationTable("name")(d)
	if d.cfg.MigrationsTable != "name" {
		t.Fatalf("failed to set migration table name")
	}
}

func TestWithVerboseLogging(t *testing.T) {
	d := &driver{}

	WithVerboseLogging(true)(d)
	if d.verbose != true {
		t.Fatalf("failed to set verbose flag")
	}
}

func Test_driver_Close(t *testing.T) {
	d := &driver{}
	err := d.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_driver_GetVersion(t *testing.T) {

}

func Test_driver_Lock(t *testing.T) {

}

func Test_driver_Reset(t *testing.T) {

}

func Test_driver_RunMigration(t *testing.T) {

}

func Test_driver_SetVersion(t *testing.T) {

}

func Test_driver_Unlock(t *testing.T) {

}

func Test_driver_getLockingKey(t *testing.T) {
	d := &driver{cfg: &config{DatabaseName: "testdb"}}
	key := d.getLockingKey()
	if key != "2584668960" {
		t.Fatalf("unexpected key 2584668960, got: %s", key)
	}

	d = &driver{cfg: &config{DatabaseName: "testdb2"}}
	key = d.getLockingKey()
	if key != "2083671126" {
		t.Fatalf("unexpected key 2083671126, got: %s", key)
	}
}

func Test_driver_prepareMigrationTable(t *testing.T) {

}
