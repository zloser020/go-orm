# Go 对象构造设计模式

[返回 ORM 框架学习笔记](orm框架.md)

## 1. Builder 模式

Builder（建造者）模式将复杂对象的构造过程拆分成多个步骤。调用方逐步设置构造信息，Builder 在内部保存中间状态，最后通过 `Build` 生成目标对象。

```text
创建 Builder → 设置参数 → 组合可选步骤 → Build → 最终对象
```

### 1.1 适用场景

Builder 适合以下情况：

- 对象构造步骤较多。
- 包含多个可选部分。
- 参数之间存在组合和顺序关系。
- 需要在生成最终结果前校验整体状态。
- 希望用链式 API 表达构造意图。

SQL 是典型场景。一条语句可能包含 `SELECT`、`FROM`、`WHERE`、`GROUP BY`、`ORDER BY` 和 `LIMIT` 等部分，而且大多数子句都是可选的。

### 1.2 基本结构

```go
type QueryBuilder struct {
	table string
	where []string
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

func (b *QueryBuilder) From(table string) *QueryBuilder {
	b.table = table
	return b
}

func (b *QueryBuilder) Where(expr string) *QueryBuilder {
	b.where = append(b.where, expr)
	return b
}

func (b *QueryBuilder) Build() (string, error) {
	if b.table == "" {
		return "", errors.New("table is required")
	}
	// 根据已收集状态组装最终 SQL。
	return "SELECT ...", nil
}
```

调用方可以逐步表达意图：

```go
query, err := NewQueryBuilder().
	From("users").
	Where("age > 18").
	Build()
```

### 1.3 核心特点

- Builder 是有状态的，会在 `Build` 之前持续保存中间信息。
- 设置方法返回 Builder 自身可以实现链式调用，但链式调用不是 Builder 模式的必要条件。
- `Build` 是统一的收口，适合执行完整性校验并生成最终对象。
- Builder 默认通常不是并发安全的，不应由多个 goroutine 无同步地共享和修改。
- 需要明确 `Build` 后 Builder 能否复用；如果内部累积字符串或参数，重复调用可能导致结果重复。

### 1.4 ORM 中的 SQL Builder

在 ORM 中，`Selector`、`Inserter`、`Updater` 和 `Deleter` 可以分别作为不同 SQL 的 Builder。`From`、`Where`、`OrderBy` 等方法收集查询意图，`Build` 负责按 SQL 语法顺序构造语句和绑定参数。

```text
查询意图 → Builder 中间状态 → SQL + Args
```

详细的项目实现见 [ORM 框架学习笔记的 SELECT 章节](orm框架.md#2-select-起步)。

## 2. Functional Options 模式

Functional Options（函数选项模式）是 Go 中常用的对象构造方式。它将必传参数放在构造函数的普通参数中，将可选配置封装成函数，再通过可变参数传入构造函数。

```text
必传参数 → 创建默认对象 → 依次应用 Option → 返回最终对象
```

### 2.1 基本结构

Functional Options 通常由四部分组成：

1. 需要构造的目标类型。
2. 接收目标对象指针的 Option 函数类型。
3. `WithXxx` Option 生成函数。
4. 接收 `opts ...Option` 的构造函数。

```go
type MyStruct struct {
	id      string
	name    string
	address string
}

type MyStructOption func(*MyStruct)

func WithMyStructAddress(address string) MyStructOption {
	return func(ms *MyStruct) {
		ms.address = address
	}
}

func NewMyStruct(id, name string, opts ...MyStructOption) *MyStruct {
	res := &MyStruct{
		id:   id,
		name: name,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}
```

使用时只需传入需要的可选配置：

```go
ms := NewMyStruct(
	"001",
	"George",
	WithMyStructAddress("Shenzhen"),
)
```

### 2.2 为什么使用 Functional Options

如果构造函数包含大量可选参数，直接使用普通参数会带来一些问题：

```go
NewServer("localhost", 8080, 30, true, false, "")
```

调用方很难从参数值理解每个配置的意义，而且新增可选参数可能迫使大量已有调用一起修改。

Functional Options 的调用意图更清晰：

```go
NewServer(
	"localhost",
	WithPort(8080),
	WithTimeout(30*time.Second),
	WithTLS(true),
)
```

它的主要优点是：

- 必传参数和可选参数职责清晰。
- Option 名称能直接表达配置意图。
- 调用方只传入需要的配置，其余使用默认值。
- 新增 Option 通常不会破坏旧的构造函数调用。
- Option 可以组合、复用，也适合在不同包中扩展 API。

### 2.3 Option 的应用顺序

Option 会按传入顺序依次执行。多个 Option 修改同一属性时，通常以最后一个为准：

```go
NewMyStruct(
	"001",
	"George",
	WithMyStructAddress("Shenzhen"),
	WithMyStructAddress("Shanghai"),
)
```

最终 `address` 是 `Shanghai`。这种“后传入的配置覆盖前面配置”的语义应当在 API 中保持一致。

### 2.4 需要校验时如何设计

如果 Option 可能遇到非法参数或彼此冲突的配置，可以让 Option 返回 `error`：

```go
type Option func(*MyStruct) error

func WithAddress(address string) Option {
	return func(ms *MyStruct) error {
		if address == "" {
			return errors.New("address cannot be empty")
		}
		ms.address = address
		return nil
	}
}

func NewMyStruct(id, name string, opts ...Option) (*MyStruct, error) {
	res := &MyStruct{id: id, name: name}
	for _, opt := range opts {
		if err := opt(res); err != nil {
			return nil, err
		}
	}
	return res, nil
}
```

如果所有 Option 都只是简单赋值，不存在失败可能，则没有必要为了形式统一而强制返回 `error`。

### 2.5 ORM 项目中的 ModelOption

当 ORM 允许用户在反射解析结果之上修改表名或列名时，Functional Options 非常合适：

```go
type ModelOption func(*Model) error

m, err := registry.Register(
	&User{},
	ModelWithTableName("user_detail"),
	ModelWithFieldName("Name", "user_name"),
)
```

应用 `ModelOption` 时要维护模型元数据的整体一致性。例如，修改字段的 `ColName` 时，不能只修改 `FieldMap` 中的 `Field`，还要将 `ColumnMap` 中的旧列名删除，再建立新列名到该字段的映射。

```go
func ModelWithFieldName(field, colName string) ModelOption {
	return func(m *Model) error {
		fd, ok := m.fieldMap[field]
		if !ok {
			return errs.NewErrUnkonwnField(field)
		}
		delete(m.columnMap, fd.colName)
		fd.colName = colName
		m.columnMap[colName] = fd
		return nil
	}
}
```

## 3. Builder 与 Functional Options 对比

| 对比项 | Builder | Functional Options |
| --- | --- | --- |
| 主要目标 | 分步收集复杂对象的构造信息 | 为构造函数提供可读、可扩展的可选配置 |
| 常见调用方式 | `NewBuilder().A(...).B(...).Build()` | `NewXxx(required, WithA(...), WithB(...))` |
| 状态保存 | Builder 在 `Build` 前持续保存中间状态 | Option 通常在构造期间立即应用 |
| 生成时机 | 通过 `Build` 显式生成最终结果 | 构造函数返回时对象已完成配置 |
| 适合场景 | SQL、复杂请求和多阶段对象构造 | 客户端、服务器、数据库句柄和模型配置 |

两种模式不冲突。例如可以使用 Functional Options 创建 ORM 的 `DB`，再使用 Builder 逐步构造 SQL。

### 3.1 如何选择

```text
只是为一次构造提供可选配置？
    └── 使用 Functional Options

需要分多步收集状态，最后统一校验和生成结果？
    └── 使用 Builder

Builder 自身也有大量可选的初始化配置？
    └── Functional Options + Builder
```

在 ORM 中：

- `Open(driver, dsn, opts...)` 或模型注册配置适合 Functional Options。
- `Selector.Where(...).OrderBy(...).Limit(...).Build()` 适合 Builder。
- Option 负责“初始化时如何配置”，Builder 负责“如何逐步构造一次操作”。

### 3.2 Functional Options 使用注意事项

- 不要把明显的必传参数全部设计成 Option，否则会将构造错误推迟到运行时。
- 应先建立完整的默认值，再让 Option 覆盖默认配置。
- 要明确多个 Option 修改同一属性时的覆盖顺序。
- 如果允许传入 `nil` Option，应在调用前判断 `opt != nil`；如果将 `nil` 视为编程错误，则可以直接触发 panic。
- Option 修改多个关联字段时，必须保证对象的不变式和内部映射一致。
- Option 应尽量只用于对象初始化；对象已被多个 goroutine 共享后，不应在没有同步机制的情况下再应用会修改它的 Option。

一句话概括：**Functional Options 用具名配置函数表达可选参数，在保持构造 API 清晰的同时，为后续扩展留出空间。**
