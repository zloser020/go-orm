package orm

import (
	"orm/internal/errs"
	"reflect"
	"unicode"
)

type model struct {
	tableName string
	fields    map[string]*field
}

type field struct {
	colName string
}

// registry 代表的是元数据的注册中心
type registry struct {
	models map[reflect.Type]*model
}

func newRegistry() *registry {
	return &registry{
		models: make(map[reflect.Type]*model, 64),
	}
}

func (r *registry) Get(val any) (*model, error) {
	typ := reflect.TypeOf(val)
	m, ok := r.models[typ]
	if !ok {
		var err error
		m, err = r.parseModel(val)
		if err != nil {
			return nil, err
		}
		r.models[typ] = m
	}
	return m, nil
}

// 限制只支持一级指针
func (r *registry) parseModel(entity any) (*model, error) {
	typ := reflect.TypeOf(entity)

	// 只支持一级指针
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()
	numFields := typ.NumField()
	fieldMap := make(map[string]*field, numFields)
	for i := 0; i < numFields; i++ {
		fd := typ.Field(i)
		fieldMap[fd.Name] = &field{
			colName: underscoreName(fd.Name),
		}
	}
	return &model{
		tableName: underscoreName(typ.Name()),
		fields:    fieldMap,
	}, nil
}

// underscoreName 驼峰转下划线
// ID -> i_d
// TestModel -> test_name
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}
	}
	return string(buf)
}
