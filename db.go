package orm

type DBOption func(db *DB)
type DB struct {
	registry *registry
}

func NewDB(opts ...DBOption) (*DB, error) {
	res := &DB{
		registry: newRegistry(),
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func MustNewDB(opts ...DBOption) *DB {
	db, err := NewDB(opts...)
	if err != nil {
		panic(err)
	}
	return db
}
