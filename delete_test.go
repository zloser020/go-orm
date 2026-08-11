package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleter_Build(t *testing.T) {
	db := memoryDB(t)

	testCases := []struct {
		name    string
		builder QueryBuilder
		wantErr error
		wantRes *Query
	}{
		{
			name:    "delete",
			builder: NewDeleter[TestModel](db),
			wantErr: nil,
			wantRes: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			name:    "where",
			builder: NewDeleter[TestModel](db).Where(C("Id").Eq("0223")),
			wantErr: nil,
			wantRes: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `id` = ?;",
				Args: []interface{}{"0223"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if tc.wantErr != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
