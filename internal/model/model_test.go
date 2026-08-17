package model

import (
	"database/sql"
	"orm/internal/errs"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

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
				TableName: "test_model",
				FieldMap: map[string]*Field{
					"Id": {
						ColName: "id",
					},
					"FirstName": {
						ColName: "first_name",
					},
					"LastName": {
						ColName: "last_name",
					},
					"Age": {
						ColName: "age",
					},
				},
			},
		},
	}

	r := NewRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := r.Register(tt.entity)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}
			assertModelMetadata(t, tt.entity, tt.wantModel, res)
		})
	}
}

func TestRegistry_RegisterNil(t *testing.T) {
	r := NewRegistry()
	model, err := r.Register(nil)
	assert.ErrorIs(t, err, errs.ErrPointerOnly)
	assert.Nil(t, model)
}

func TestRegistry_IgnoreUnexportedFields(t *testing.T) {
	type privateModel struct {
		Public  string
		private string
	}

	r := NewRegistry()
	model, err := r.Register(&privateModel{})
	assert.NoError(t, err)
	assert.Contains(t, model.FieldMap, "Public")
	assert.NotContains(t, model.FieldMap, "private")
	assert.NotContains(t, model.ColumnMap, "private")
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
				TableName: "test_model",
				FieldMap: map[string]*Field{
					"Id": {
						ColName: "id",
					},
					"FirstName": {
						ColName: "first_name",
					},
					"LastName": {
						ColName: "last_name",
					},
					"Age": {
						ColName: "age",
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
				TableName: "tag_table",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name_t",
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
				TableName: "tag_table",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
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
				TableName: "tag_table",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
					},
				},
			},
		},
		{
			name:   "table name",
			entity: &UserName{},
			wantModel: &Model{
				TableName: "user_name_table",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
					},
				},
			},
		},
	}

	r := NewRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := r.Get(tt.entity)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}
			assertModelMetadata(t, tt.entity, tt.wantModel, res)
			cached, cacheErr := r.Get(tt.entity)
			assert.NoError(t, cacheErr)
			assert.Same(t, res, cached)
		})
	}
}

func assertModelMetadata(t *testing.T, entity any, want, actual *Model) {
	t.Helper()
	if !assert.NotNil(t, actual) {
		return
	}
	assert.Equal(t, want.TableName, actual.TableName)
	assert.Len(t, actual.FieldMap, len(want.FieldMap))
	assert.Len(t, actual.ColumnMap, len(want.FieldMap))

	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	for goName, wantField := range want.FieldMap {
		actualField, ok := actual.FieldMap[goName]
		if !assert.True(t, ok, "fieldMap should contain %q", goName) {
			continue
		}
		structField, ok := typ.FieldByName(goName)
		if !assert.True(t, ok, "struct should contain field %q", goName) {
			continue
		}
		assert.Equal(t, goName, actualField.GoName)
		assert.Equal(t, wantField.ColName, actualField.ColName)
		assert.Equal(t, structField.Type, actualField.Typ)
		assert.Equal(t, structField.Offset, actualField.Offset)

		columnField, ok := actual.ColumnMap[wantField.ColName]
		if assert.True(t, ok, "columnMap should contain %q", wantField.ColName) {
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
	r := NewRegistry()
	m, err := r.Register(&TestModel{}, WithTableName("user_name_table"))
	assert.NoError(t, err)
	assert.Equal(t, "user_name_table", m.TableName)
}

func TestModelWithFieldName(t *testing.T) {
	r := NewRegistry()
	m, err := r.Register(&UserName{}, WithFieldName("FirstName", "first_name_a"))
	assert.NoError(t, err)
	assert.Equal(t, "first_name_a", m.FieldMap["FirstName"].ColName)
	_, ok := m.ColumnMap["first_name"]
	assert.False(t, ok)
	assert.Same(t, m.FieldMap["FirstName"], m.ColumnMap["first_name_a"])
}
