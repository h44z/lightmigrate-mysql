# LightMigrate - MySQL/MariaDB migration driver

[![codecov](https://codecov.io/gh/h44z/lightmigrate-mysql/branch/master/graph/badge.svg?token=N7H27SQUUW)](https://codecov.io/gh/h44z/lightmigrate-mongodb)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://pkg.go.dev/badge/github.com/h44z/lightmigrate-mysql/mysql)](https://pkg.go.dev/github.com/h44z/lightmigrate-mysql/mysql)
![GitHub last commit](https://img.shields.io/github/last-commit/h44z/lightmigrate-mysql)
[![Go Report Card](https://goreportcard.com/badge/github.com/h44z/lightmigrate-mysql)](https://goreportcard.com/report/github.com/h44z/lightmigrate-mysql)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/h44z/lightmigrate-mysql)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/h44z/lightmigrate-mysql)
[![GitHub Release](https://img.shields.io/github/release/h44z/lightmigrate-mysql.svg)](https://github.com/h44z/lightmigrate-mysql/releases)

This module is part of the [LightMigrate](https://github.com/h44z/lightmigrate) library.
It provides a migration driver for MySQL or MariaDB.

## Features
 * Driver work with MySQL or MariaDB. 
 * If the database client was initialized with `multiStatements=true`, multiple statements are supported within the migration files.
 * [Examples](./examples)

## Configuration Options

Configuration options can be passed to the constructor using the `With<Config-Option>` functions.

| Config Value      | Defaults          | Description                                        |
|-------------------|-------------------|----------------------------------------------------|
| `MigrationsTable` | schema_migrations | Name of the migrations table.                      |
| `Locking`         | true              | If database locking should be used.                |
| `Logger`          | log.Default()     | The logger instance that should be used.           |
| `VerboseLogging`  | false             | If set to true, more log messages will be printed. |