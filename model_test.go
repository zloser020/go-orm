package orm

import (
	"orm/internal/errs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseModel(t *testing.T) {

	tests := []struct {
		name      string
		entity    any
		wantErr   error
		wantModel *model
	}{
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "test model pointer",
			entity:  &TestModel{},
			wantErr: nil,
			wantModel: &model{
				tableName: "test_model",
				fields: map[string]*field{
					"id": {
						colName: "id",
					},
					"FirstName": {
						colName: "FirstName",
					},
					"LastName": {
						colName: "LastName",
					},
					"Age": {
						colName: "Age",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parseModel(tt.entity)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				return
			}
			assert.Equal(t, tt.wantModel, res)
		})
	}
}
