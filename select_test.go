package orm

import (
	"context"
	"database/sql"
	"orm/internal/errs"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelector_Build(t *testing.T) {
	db := memoryDB(t)

	testCases := []struct {
		name      string
		builder   QueryBuilder
		wantErr   error
		wantQuery *Query
	}{
		{
			name:    "select",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "from",
			builder: NewSelector[TestModel](db).From("Test_Model"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM Test_Model;",
				Args: nil,
			},
		},
		{
			name:    "from`",
			builder: NewSelector[TestModel](db).From("`Test_Model`"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `Test_Model`;",
				Args: nil,
			},
		},
		{
			name:    "empty_from",
			builder: NewSelector[TestModel](db).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "where",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` = ?;",
				Args: []any{18},
			},
		},
		{
			name:    "not",
			builder: NewSelector[TestModel](db).Where(Not(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`age` = ?);",
				Args: []any{18},
			},
		},
		{
			name:    "and",
			builder: NewSelector[TestModel](db).Where(C("Id").Eq("0223").And(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`id` = ?) AND (`age` = ?);",
				Args: []any{"0223", 18},
			},
		},
		{
			name:    "or",
			builder: NewSelector[TestModel](db).Where(C("Id").Eq("0223").Or(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`id` = ?) OR (`age` = ?);",
				Args: []any{"0223", 18},
			},
		},
		{
			name:    "and",
			builder: NewSelector[TestModel](db).Where(C("Id").Eq("0223"), (C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`id` = ?) AND (`age` = ?);",
				Args: []any{"0223", 18},
			},
		},
		{
			name:    "invalid column",
			builder: NewSelector[TestModel](db).Where(C("level").Eq("0")),
			wantErr: errs.NewErrUnkonwnField("level"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestSelector_BuildRepeatedly(t *testing.T) {
	db := memoryDB(t)
	builder := NewSelector[TestModel](db).Where(C("Age").Eq(18))

	first, err := builder.Build()
	require.NoError(t, err)
	second, err := builder.Build()
	require.NoError(t, err)

	assert.Equal(t, first, second)
}

func TestSelector_Get(t *testing.T) {
	testCases := []struct {
		name string
		opts []DBOption
	}{
		{name: "reflect"},
		{name: "unsafe", opts: []DBOption{DBWithUnsafeValuer()}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := selectorTestDB(t, "get_"+tc.name, tc.opts...)
			insertTestModels(t, db)

			got, err := NewSelector[TestModel](db).
				Where(C("Id").Eq(int64(1))).
				Get(context.Background())
			require.NoError(t, err)
			assert.Equal(t, &TestModel{
				Id:        1,
				FirstName: "George",
				Age:       18,
				LastName:  &sql.NullString{String: "Wu", Valid: true},
			}, got)
		})
	}
}

func TestSelector_GetMulti(t *testing.T) {
	db := selectorTestDB(t, "get_multi")
	insertTestModels(t, db)

	got, err := NewSelector[TestModel](db).GetMulti(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 2)

	byID := make(map[int64]*TestModel, len(got))
	for _, entity := range got {
		byID[entity.Id] = entity
	}
	assert.Equal(t, "George", byID[1].FirstName)
	assert.Equal(t, "Tom", byID[2].FirstName)
}

func TestSelector_GetNoRows(t *testing.T) {
	db := selectorTestDB(t, "get_no_rows")

	got, err := NewSelector[TestModel](db).Get(context.Background())
	assert.ErrorIs(t, err, ErrNoRows)
	assert.Nil(t, got)
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func memoryDB(t *testing.T) *DB {
	db, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	return db
}

func selectorTestDB(t *testing.T, name string, opts ...DBOption) *DB {
	t.Helper()
	db, err := Open("sqlite3", "file:"+name+"?cache=shared&mode=memory", opts...)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.db.Close())
	})
	_, err = db.db.Exec(`
		CREATE TABLE test_model (
			id INTEGER,
			first_name TEXT,
			age INTEGER,
			last_name TEXT
		)`)
	require.NoError(t, err)
	return db
}

func insertTestModels(t *testing.T, db *DB) {
	t.Helper()
	_, err := db.db.Exec(`
		INSERT INTO test_model(id, first_name, age, last_name)
		VALUES (1, 'George', 18, 'Wu'), (2, 'Tom', 20, NULL)
	`)
	require.NoError(t, err)
}
