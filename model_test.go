package orm

import (
	"orm/internal/errs"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseModel(t *testing.T) {

	tests := []struct {
		name      string
		entity    any
		wantErr   error
		wantModel *Model
	}{
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "test Model pointer",
			entity:  &TestModel{},
			wantErr: nil,
			wantModel: &Model{
				tableName: "test_model",
				fields: map[string]*Field{
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
			res, err := r.Registry(tt.entity)
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
		wantModel *Model
		cacheSize int
	}{
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "test Model pointer",
			entity:  &TestModel{},
			wantErr: nil,
			wantModel: &Model{
				tableName: "test_model",
				fields: map[string]*Field{
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
		{
			name: "tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column:first_name_t"`
				}
				return &TagTable{}
			}(),
			wantErr: nil,
			wantModel: &Model{
				tableName: "tag_table",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name_t",
					},
				},
			},
		},
		{
			name: "empty",
			entity: func() any {
				type TagTable struct {
					FirstName string
				}
				return &TagTable{}
			}(),
			wantErr: nil,
			wantModel: &Model{
				tableName: "tag_table",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		{
			name: "column only",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column"`
				}
				return &TagTable{}
			}(),
			wantErr: errs.NewErrInvalidTagContent("column"),
		},
		{
			name: "column",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column:"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				tableName: "tag_table",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		{
			name:   "table name",
			entity: &UserName{},
			wantModel: &Model{
				tableName: "user_name_table",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
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
			typ := reflect.TypeOf(tt.entity)
			cache, ok := r.models.Load(typ)
			assert.True(t, ok)
			if !ok {
				return
			}
			assert.Equal(t, tt.wantModel, cache.(*Model))
		})
	}
}

type UserName struct {
	FirstName string `orm:"column:first_name"`
}

func (u *UserName) TableName() string {
	return "user_name_table"
}

func TestModelWithTableName(t *testing.T) {
	r := newRegistry()
	m, err := r.Registry(&TestModel{}, ModelWithTableName("user_name_table"))
	assert.NoError(t, err)
	assert.Equal(t, "user_name_table", m.tableName)
}

func TestModelWithFieldName(t *testing.T) {
	r := newRegistry()
	m, err := r.Registry(&UserName{}, ModelWithFieldName("FirstName", "first_name_a"))
	assert.NoError(t, err)
	assert.Equal(t, "first_name_a", m.fields["FirstName"].colName)
}
