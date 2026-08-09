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
					"Id": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name",
					},
					"LastName": {
						colName: "last_name",
					},
					"Age": {
						colName: "age",
					},
				},
			},
		},
	}

	r := &registry{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := r.parseModel(tt.entity)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}
			assert.Equal(t, tt.wantModel, res)
		})
	}
}

func TestRegistry_get(t *testing.T) {
	tests := []struct {
		name      string
		entity    any
		wantErr   error
		wantModel *model
		cacheSize int
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
					"Id": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name",
					},
					"LastName": {
						colName: "last_name",
					},
					"Age": {
						colName: "age",
					},
				},
			},
			cacheSize: 1,
		},
	}

	r := newRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := r.Get(tt.entity)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}
			assert.Equal(t, tt.wantModel, res)
			assert.Equal(t, tt.cacheSize, len(r.models))
		})
	}
}
