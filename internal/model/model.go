package model

import (
	"orm/internal/errs"
	"orm/internal/valuer"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

const (
	tagKeyColumn = "column"
)

type TableName interface {
	TableName() string
}

type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...ModelOption) (*Model, error)
}

type Model struct {
	TableName string
	// 字段名到字段定义的映射
	FieldMap map[string]*Field
	// 列名到字段定义的映射
	ColumnMap map[string]*Field
}

type ModelOption func(*Model) error

type Field struct {
	GoName  string
	ColName string
	Typ     reflect.Type
	Offset  uintptr
}

// registry 代表的是元数据的注册中心
type registry struct {
	models sync.Map
}

var _ Registry = (*registry)(nil)

func newRegistry() *registry {
	return &registry{}
}

func NewRegistry() Registry {
	return newRegistry()
}

func (r *registry) Get(val any) (*Model, error) {
	if val == nil {
		return nil, errs.ErrPointerOnly
	}
	typ := reflect.TypeOf(val)
	m, ok := r.models.Load(typ)
	if ok {
		return m.(*Model), nil
	}
	var err error
	m, err = r.Register(val)
	if err != nil {
		return nil, err
	}
	return m.(*Model), nil
}

// 限制只支持一级指针
func (r *registry) Register(entity any, opts ...ModelOption) (*Model, error) {
	if entity == nil {
		return nil, errs.ErrPointerOnly
	}
	typ := reflect.TypeOf(entity)

	// 只支持一级指针
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	elemType := typ.Elem()
	numFields := elemType.NumField()
	fieldMap := make(map[string]*Field, numFields)
	columnMap := make(map[string]*Field, numFields)
	for i := 0; i < numFields; i++ {
		fd := elemType.Field(i)
		if !fd.IsExported() {
			continue
		}
		pair, err := r.parseTag(fd.Tag)
		if err != nil {
			return nil, err
		}
		colName := pair[tagKeyColumn]
		if colName == "" {
			colName = underscoreName(fd.Name)
		}

		fdMeta := &Field{
			ColName: colName,
			Typ:     fd.Type,
			GoName:  fd.Name,
			Offset:  uintptr(fd.Offset),
		}

		fieldMap[fd.Name] = fdMeta
		columnMap[colName] = fdMeta
	}
	var tableName string
	if tbl, ok := entity.(TableName); ok {
		tableName = tbl.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(elemType.Name())
	}
	res := &Model{
		TableName: tableName,
		FieldMap:  fieldMap,
		ColumnMap: columnMap,
	}
	for _, opt := range opts {
		err := opt(res)
		if err != nil {
			return nil, err
		}
	}
	r.models.Store(typ, res)
	return res, nil
}

func (m *Model) FieldByColumn(column string) (valuer.Field, bool) {
	fd, ok := m.ColumnMap[column]
	if !ok {
		return valuer.Field{}, false
	}
	return valuer.Field{
		GoName: fd.GoName,
		Typ:    fd.Typ,
		Offset: fd.Offset,
	}, true
}

func WithTableName(tableName string) ModelOption {
	return func(m *Model) error {
		m.TableName = tableName
		return nil
	}
}

func WithFieldName(field string, colName string) ModelOption {
	return func(m *Model) error {
		fd, ok := m.FieldMap[field]
		if !ok {
			return errs.NewErrUnkonwnField(field)
		}
		delete(m.ColumnMap, fd.ColName)
		fd.ColName = colName
		m.ColumnMap[colName] = fd
		return nil
	}
}

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag, ok := tag.Lookup("orm")
	if !ok {
		return map[string]string{}, nil
	}
	pairs := strings.Split(ormTag, ",")
	result := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		segs := strings.Split(pair, ":")
		if len(segs) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		key := segs[0]
		val := segs[1]
		result[key] = val
	}
	return result, nil
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
