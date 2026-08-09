package orm

import (
	"database/sql"
	"orm/internal/errs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelector_Build(t *testing.T) {
	db, err := NewDB()
	require.NoError(t, err)

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

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
