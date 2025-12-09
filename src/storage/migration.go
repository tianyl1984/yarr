package storage

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
)

var migrations = []func(*sql.Tx) error{
	m01_initial,
	m02_feedconfig,
	m03_drop_http_states,
}

var maxVersion = int64(len(migrations))

func migrate(db *sql.DB) error {
	version, err := getVersion(db)
	if err != nil {
		return err
	}
	if version >= maxVersion {
		return nil
	}

	log.Printf("db version is %d. migrating to %d", version, maxVersion)

	for v := version + 1; v <= maxVersion; v++ {
		log.Printf("[migration:%d] starting", v)
		err := migrateVersion(v, db)
		if err != nil {
			return err
		}
		log.Printf("[migration:%d] done", v)
	}
	return nil
}

func getVersion(db *sql.DB) (int64, error) {
	exist, err := checkTableExists(db, "settings")
	if err != nil {
		return 0, err
	}
	var version int64
	if !exist {
		version = 0
	} else {
		if err := db.QueryRow("select val from settings where `key` = ?", "db_version").Scan(&version); err != nil {
			return 0, err
		}
	}
	return version, nil
}

func checkTableExists(db *sql.DB, tableName string) (bool, error) {
	sql := "select count(1) from information_schema.tables where table_schema = database() and table_name = ?"
	var count int
	err := db.QueryRow(sql, tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func migrateVersion(v int64, db *sql.DB) error {
	var err error
	var tx *sql.Tx
	migratefunc := migrations[v-1]
	if tx, err = db.Begin(); err != nil {
		log.Printf("[migration:%d] failed to start transaction", v)
		return err
	}
	if err = migratefunc(tx); err != nil {
		log.Printf("[migration:%d] failed to migrate", v)
		tx.Rollback()
		return err
	}
	if err = tx.Commit(); err != nil {
		log.Printf("[migration:%d] failed to commit changes", v)
		return err
	}
	return nil
}

//go:embed sql/m01_initial.sql
var m01_initial_sql string

//go:embed sql/m02_feedconfig.sql
var m02_feedconfig_sql string

//go:embed sql/m03_drop_http_states.sql
var m03_drop_http_states_sql string

func m01_initial(tx *sql.Tx) error {
	fmt.Println(m01_initial_sql)
	_, err := tx.Exec(m01_initial_sql)
	return err
}

func m02_feedconfig(tx *sql.Tx) error {
	fmt.Println(m02_feedconfig_sql)
	_, err := tx.Exec(m02_feedconfig_sql)
	return err
}

func m03_drop_http_states(tx *sql.Tx) error {
	fmt.Println(m03_drop_http_states_sql)
	_, err := tx.Exec(m03_drop_http_states_sql)
	return err
}
