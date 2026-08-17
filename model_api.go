package orm

import internalmodel "orm/internal/model"

// 根包只保留公开 API 兼容层；模型元数据实现位于 internal/model。
type Registry = internalmodel.Registry
type Model = internalmodel.Model
type Field = internalmodel.Field
type ModelOption = internalmodel.ModelOption

func NewRegistry() Registry {
	return internalmodel.NewRegistry()
}

func ModelWithTableName(tableName string) ModelOption {
	return internalmodel.WithTableName(tableName)
}

func ModelWithFieldName(field, colName string) ModelOption {
	return internalmodel.WithFieldName(field, colName)
}
