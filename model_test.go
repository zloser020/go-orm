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
				fieldMap: map[string]*Field{
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
			assertModelMetadata(t, tt.entity, tt.wantModel, res)
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
				fieldMap: map[string]*Field{
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
				fieldMap: map[string]*Field{
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
				fieldMap: map[string]*Field{
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
				fieldMap: map[string]*Field{
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
				fieldMap: map[string]*Field{
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
			assertModelMetadata(t, tt.entity, tt.wantModel, res)
			typ := reflect.TypeOf(tt.entity)
			cache, ok := r.models.Load(typ)
			assert.True(t, ok)
			if !ok {
				return
			}
			assert.Same(t, res, cache.(*Model))
		})
	}
}

func assertModelMetadata(t *testing.T, entity any, want, actual *Model) {
	t.Helper()
	if !assert.NotNil(t, actual) {
		return
	}
	assert.Equal(t, want.tableName, actual.tableName)
	assert.Len(t, actual.fieldMap, len(want.fieldMap))
	assert.Len(t, actual.columnMap, len(want.fieldMap))

	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	for goName, wantField := range want.fieldMap {
		actualField, ok := actual.fieldMap[goName]
		if !assert.True(t, ok, "fieldMap should contain %q", goName) {
			continue
		}
		structField, ok := typ.FieldByName(goName)
		if !assert.True(t, ok, "struct should contain field %q", goName) {
			continue
		}
		assert.Equal(t, goName, actualField.goName)
		assert.Equal(t, wantField.colName, actualField.colName)
		assert.Equal(t, structField.Type, actualField.typ)

		columnField, ok := actual.columnMap[wantField.colName]
		if assert.True(t, ok, "columnMap should contain %q", wantField.colName) {
			assert.Same(t, actualField, columnField)
		}
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
	assert.Equal(t, "first_name_a", m.fieldMap["FirstName"].colName)
	_, ok := m.columnMap["first_name"]
	assert.False(t, ok)
	assert.Same(t, m.fieldMap["FirstName"], m.columnMap["first_name_a"])
}
