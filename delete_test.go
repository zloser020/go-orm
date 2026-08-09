package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleter_Build(t *testing.T) {
	testCases := []struct {
		name    string
		builder QueryBuilder
		wantErr error
		wantRes *Query
	}{
		{
			name:    "delete",
			builder: &Deleter[TestModel]{},
			wantErr: nil,
			wantRes: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			name:    "where",
			builder: (&Deleter[TestModel]{}).Where(C("Id").Eq("0223")),
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
