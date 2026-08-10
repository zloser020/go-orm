package orm

import (
	"orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

const (
	tagKeyColumn = "column"
)

type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...ModelOption) (*Model, error)
}

type Model struct {
	tableName string
	fields    map[string]*Field
}

type ModelOption func(*Model) error

type Field struct {
	colName string
}

// registry 代表的是元数据的注册中心
type registry struct {
	// 读多写少 使用读写锁
	// lock sync.RWMutex
	// models map[reflect.Type]*Model

	// 使用sync.map
	models sync.Map
}

func newRegistry() *registry {
	return &registry{}
}

func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	m, ok := r.models.Load(typ)
	if ok {
		return m.(*Model), nil
	}
	var err error
	m, err = r.Registry(val)
	if err != nil {
		return nil, err
	}
	// 可能存在覆盖，影响有限
	r.models.Store(typ, m)
	return m.(*Model), nil
}

// 使用读写锁
//func (r *registry) Get(val any) (*Model, error) {
//	typ := reflect.TypeOf(val)
//
//	r.lock.RLock()
//	m, ok := r.models[typ]
//	r.lock.RUnlock()
//	if ok {
//		return m, nil
//	}
//
//	r.lock.Lock()
//	defer r.lock.Unlock()
//	m, ok = r.models[typ]
//	if ok {
//		return m, nil
//	}
//	m, err := r.Registry(val)
//	if err != nil {
//		return nil, err
//	}
//	r.models[typ] = m
//	return m, nil
//}

// 限制只支持一级指针
func (r *registry) Registry(entity any, opts ...ModelOption) (*Model, error) {
	Typ := reflect.TypeOf(entity)

	// 只支持一级指针
	if Typ.Kind() != reflect.Ptr || Typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	elemType := Typ.Elem()
	numFields := elemType.NumField()
	fieldMap := make(map[string]*Field, numFields)
	for i := 0; i < numFields; i++ {
		fd := elemType.Field(i)
		pair, err := r.parseTag(fd.Tag)
		if err != nil {
			return nil, err
		}
		colName := pair[tagKeyColumn]
		if colName == "" {
			colName = underscoreName(fd.Name)
		}
		fieldMap[fd.Name] = &Field{
			colName: colName,
		}
	}
	var tableName string
	if tbl, ok := entity.(TableName); ok {
		tableName = tbl.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(elemType.Name())
	}
	res := &Model{
		tableName: tableName,
		fields:    fieldMap,
	}
	for _, opt := range opts {
		err := opt(res)
		if err != nil {
			return nil, err
		}
	}
	r.models.Store(Typ, res)
	return res, nil
}

func ModelWithTableName(tableName string) ModelOption {
	return func(m *Model) error {
		m.tableName = tableName
		//if tableName == "" {
		//	return err
		//}
		return nil
	}
}

func ModelWithFieldName(field string, colName string) ModelOption {
	return func(m *Model) error {
		fd, ok := m.fields[field]
		if !ok {
			return errs.NewErrUnkonwnField(field)
		}
		fd.colName = colName
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
