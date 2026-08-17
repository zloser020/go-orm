package orm

import (
	"database/sql"
	"orm/internal/valuer"
)

type DBOption func(db *DB)

// DB 是一个sql.db装饰器
type DB struct {
	registry      Registry
	db            *sql.DB
	valuerCreator valuer.Creator
}

func Open(driver string, dataSourceName string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		registry:      NewRegistry(),
		db:            db,
		valuerCreator: valuer.NewReflectValue,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func DBWithReflectValuer() DBOption {
	return func(db *DB) {
		db.valuerCreator = valuer.NewReflectValue
	}
}

func DBWithUnsafeValuer() DBOption {
	return func(db *DB) {
		db.valuerCreator = valuer.NewUnsafeValue
	}
}

func MustOpen(driver string, dataSourceName string, opts ...DBOption) *DB {
	db, err := Open(driver, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return db
}
