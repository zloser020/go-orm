# ORM 框架学习笔记

## 专题导航

ORM 实现会涉及一些可独立学习的 Go 与数据库知识，已拆分为专题文档：

- [MySQL 数据库](mysql数据库.md)：事务隔离、MVCC、undo log 与 redo log。
- [Go 反射与 unsafe](go-reflection-unsafe.md)：运行时类型、字段与方法、`unsafe.Pointer`、GC、内存对齐和字段偏移。
- [Go 对象构造设计模式](go-design-patterns.md)：Builder、Functional Options、二者的适用场景及在 ORM 中的应用。

## 1. ORM 框架概览

ORM 是 Object-Relational Mapping 的缩写，即对象关系映射。它用于在 Go 对象和关系型数据库之间进行转换。

| Go | 关系型数据库 |
| --- | --- |
| 结构体 | 表 |
| 结构体实例 | 一行记录 |
| 结构体字段 | 列 |

ORM 位于业务代码和数据库驱动之间：

```text
业务代码
    ↓
ORM 框架
    ↓
database/sql 与数据库驱动
    ↓
数据库
```

### 1.1 ORM 框架的核心是什么

ORM 框架最核心的工作是：

1. **构造 SQL**：根据模型和查询条件生成 SQL，并绑定参数。
2. **处理结果集**：将数据库返回的行和列映射到 Go 结构体。

```text
Go 对象 → 构造 SQL → 执行查询 → 读取结果集 → Go 对象
```

一些语言的底层数据库库能力较弱，ORM 还需要管理连接和会话。Go 的 `database/sql` 已经提供连接池和事务等基础能力，因此 Go ORM 通常不需要重复实现连接池。

### 1.2 为什么需要 ORM

不使用 ORM 也可以操作数据库，但需要反复处理：

- 拼装 SQL 和参数。
- 将结构体字段转换为 SQL 参数。
- 将查询结果扫描到结构体字段。
- 维护字段名与列名的映射关系。

这些工作本身不难，但比较琐碎。ORM 将它们统一封装，使业务代码更加关注查询意图和业务逻辑。

### 1.3 ORM 的优点与代价

ORM 的主要优点是 API 对编程更加友好，可以减少样板代码并提高开发效率。

它也会带来一些代价：

- 增加一层抽象，排查问题时仍要理解最终 SQL。
- 反射、SQL 构造和对象映射会产生额外开销。
- 复杂查询可能不适合使用通用 ORM API。
- 不恰当的关联查询可能导致 N+1 问题。

因此，使用 ORM 并不意味着可以不学习 SQL。

### 1.4 使用 ORM 性能会更好吗

通常不会。直接编写 SQL 的调用路径更短，也更容易针对复杂查询进行优化。

```text
直接 SQL：业务代码 → database/sql → 数据库

ORM：    业务代码 → SQL 构造与结果映射 → database/sql → 数据库
```

不过，实际性能通常更取决于最终 SQL、索引、查询次数和数据库执行情况，而不是 ORM 本身的少量开销。

常见的优化方向包括：

- 建立合适的索引。
- 只查询需要的列。
- 使用批量操作减少数据库往返。
- 避免 N+1 查询。
- 通过日志观察最终生成的 SQL。

### 1.5 ORM 如何使用缓存

缓存可以通过 AOP、插件或钩子拦截查询，在 ORM 的查询流程中加入缓存逻辑：

```text
执行查询
    ↓
读取缓存
    ├── 命中 → 返回结果
    └── 未命中 → 查询数据库 → 写入缓存 → 返回结果
```

写操作需要在事务提交后删除或更新相关缓存。缓存设计的难点是缓存键、过期时间和数据一致性，而不是简单地读取和写入缓存。

### 1.6 ORM 元数据面试要点

#### ORM 如何在结构体和表之间进行映射

ORM 框架依靠元数据建立 Go 结构体与数据库表之间的映射关系。在 Go 中，通常通过反射读取结构体的类型、字段和标签，再将这些信息解析为 ORM 模型元数据。

```text
Go 结构体
    ↓ 反射与标签解析
ORM 元数据
    ↓
数据库表、列、索引与关联关系
```

#### ORM 元数据有什么用

ORM 元数据主要用于两个方向：

1. 构造 SQL 时，将 Go 类型、字段和字段值映射为表名、列名和 SQL 参数。
2. 处理结果集时，将数据库的表和列映射回 Go 结构体及其字段。

```text
Go 模型 ──元数据──→ SQL 和参数
Go 模型 ←─元数据── 数据库结果集
```

#### ORM 元数据一般包含什么

ORM 元数据通常包含：

- 表信息。
- 列信息。
- 索引信息。
- 表之间的关联关系，如一对一、一对多和多对多。

#### 表信息包含什么

表信息是表级别的配置，最基本的是表名。如果 ORM 支持分库分表，还可能包含库名、分库键、分表键和路由策略等信息。

#### 列信息包含什么

列信息通常包含：

- 列名以及对应的 Go 字段名。
- 数据库类型以及对应的 Go 类型。
- 是否为主键、自增列、可空列或唯一列。
- 默认值、索引和关联关系等扩展信息。

#### 索引信息包含什么

索引信息主要描述：

- 索引名。
- 索引包含的列及其顺序。
- 是否为唯一索引。
- 是否为联合索引，以及联合索引中的列顺序。

一句话概括：**ORM 的本质是利用元数据，完成 Go 对象与关系型数据之间的双向映射。**

### 1.7 ORM 如何获取模型信息

ORM 主要使用反射解析 Go 类型，获取类型名、字段名、字段类型和结构体标签等信息。在约定的默认映射之外，还可以通过 Tag 或编程接口允许用户定制模型。

```text
Go 类型
  ├─ 反射：读取类型、字段和方法
  ├─ Tag：读取字段的额外映射配置
  └─ 接口：允许模型以方法提供定制配置
          ↓
      ORM 模型元数据
```

例如，当模型实现下面的接口时，ORM 可以优先使用用户指定的表名：

```go
type TableName interface {
	TableName() string
}
```

通常可以按以下优先级解析配置：

```text
编程接口或显式配置 > Tag 配置 > ORM 默认约定
```

#### Go 结构体 Tag 有什么用

Tag 用来描述字段的额外信息。Tag 本身不会主动执行任何逻辑，必须由 JSON 库、ORM 等程序通过反射读取并解释。

```go
type User struct {
	ID   int64  `json:"id" orm:"column=id;primary_key"`
	Name string `json:"name" orm:"column=user_name;index"`
}
```

不同程序只处理自己关心的 Tag：

- `encoding/json` 使用 `json` Tag 指定 JSON 字段名。
- ORM 使用 `orm` Tag 描述列名、主键、索引、自增和关联关系等信息。

可以通过 `reflect.StructField.Tag` 读取 Tag：

```go
typ := reflect.TypeOf(User{})
field, _ := typ.FieldByName("Name")
ormTag := field.Tag.Get("orm")
```

#### 如何概括 GORM、Beego ORM 等框架的实现

面试时可以先回答三个核心模块：

1. **元数据解析**：使用反射、Tag 和编程接口建立 Go 模型与数据库表的映射。
2. **SQL 构造**：根据元数据和链式 API 收集的查询条件，生成 SQL 和绑定参数。
3. **结果集处理**：读取数据库返回的列，再根据元数据将值写入 Go 结构体。

```text
反射、Tag 与接口
        ↓
    模型元数据
      ↙      ↘
  SQL 构造   结果集映射
```

如果面试官继续追问，可再展开说明 SQL 各子句的构造顺序、占位符与参数收集，以及结果集如何通过反射写入结构体字段。

## 2. SELECT 起步

### 2.1 GORM 如何构造 SQL

GORM 中主要涉及 `Builder`、`Expression`、`Clause` 和 `Interface` 四个抽象。核心思路是将 SQL 的不同部分分别构造，最后再拼接成完整语句。

```text
SELECT 部分 + FROM 部分 + WHERE 部分 + 其他子句 → 完整 SQL
```

当前项目也采用类似思路：`Selector` 负责组织 SELECT 查询，`Predicate` 表示查询条件，`BuildExpression` 递归构造条件表达式。

### 2.2 当前项目中的 SQL Builder

Builder 模式的通用原理、适用场景及与 Functional Options 的对比，见 [Go 对象构造设计模式](go-design-patterns.md)。本节只关注它在当前 ORM SQL 构造中的落地。

```text
设置查询模型
    ↓
添加 FROM、WHERE 等查询条件
    ↓
Builder 保存构造过程中的状态
    ↓
调用 Build 统一生成 SQL 和参数
```

在当前项目中，各部分的职责是：

| 组件 | 职责 |
| --- | --- |
| `Selector[T]` | Builder，保存表名和查询条件等状态 |
| `From`、`Where` | 分步骤设置查询内容，并返回 Builder 自身 |
| `Predicate` | 描述 `WHERE` 中的条件表达式 |
| `Build` | 将已保存的状态组装成最终的 `Query` |
| `Query` | 最终产物，包含 SQL 和参数 |

```go
query, err := (&Selector[TestModel]{}).
	From("users").
	Where(C("Age").Eq(18)).
	Build()
```

生成的结果类似于：

```go
&Query{
	SQL:  "SELECT * FROM users WHERE `Age` = ?;",
	Args: []any{18},
}
```

这种设计将“如何表达查询”和“如何拼接 SQL”分离：调用方负责描述查询意图，Builder 负责处理拼接顺序、占位符和参数收集。以后增加 `OrderBy`、`Limit` 等能力时，只需为 Builder 增加新的构造步骤。

### 2.3 Where 多条件合并

一次向 `Where` 传入多个条件时，循环会将它们依次用 `AND` 连接：

```go
p := s.where[0]
for i := 1; i < len(s.where); i++ {
	p = p.And(s.where[i])
}
```

例如：

```go
Where(
	C("Age").Eq(18),
	C("id").Eq("0223"),
)
```

当 `len(s.where) >= 2` 时循环才会执行，最终生成：

```sql
WHERE (`Age` = ?) AND (`id` = ?)
```

如果使用 `Where(p1.And(p2))`，传给 `Where` 的只有一个已经组合好的条件，因此不会进入循环。

### 2.4 泛型的作用

ORM 使用泛型可以约束传入的模型类型和返回值类型，减少类型断言，提高类型安全性。例如 `Selector[TestModel]` 可以明确查询对应的模型是 `TestModel`。

### 2.5 SELECT 语句顺序

面试中可能需要手写 SQL，应记住 SELECT 常见子句的书写顺序：

```sql
SELECT ...
FROM ...
WHERE ...
GROUP BY ...
HAVING ...
ORDER BY ...
LIMIT ...;
```

## 3. ORM 结果集与性能面试要点

### 3.1 ORM 如何处理数据库返回的数据

ORM 处理查询结果的核心是：**根据模型元数据，将数据库列映射到 Go 字段，为 `rows.Scan` 准备正确类型的接收地址。**

完整流程可以拆成以下步骤：

```text
执行 SQL
   ↓
获取结果集列名
   ↓
根据 ColumnMap 找到字段元数据
   ↓
为每列准备类型匹配的扫描目标
   ↓
rows.Scan 完成数据库值到 Go 值的转换
   ↓
将 Go 值写入模型字段
```

其中元数据至少需要提供：

- 数据库列名。
- Go 字段名或索引路径。
- Go 字段类型。
- 如果使用 `unsafe`，还需要字段偏移。

`database/sql` 负责将驱动返回的值转换到 `Scan` 目标类型。ORM 的主要责任是找到正确字段，并传入正确的字段指针或临时值指针。

### 3.2 使用反射处理结果集

反射方案可以先为每一列创建对应类型的临时值：

```go
scanTargets := make([]any, len(columns))
values := make([]reflect.Value, len(columns))

for i, column := range columns {
	field := model.ColumnMap[column]
	value := reflect.New(field.Typ)
	scanTargets[i] = value.Interface()
	values[i] = value.Elem()
}

if err := rows.Scan(scanTargets...); err != nil {
	return err
}
```

扫描成功后，再根据字段元数据将临时值写入目标结构体：

```go
entityValue := reflect.ValueOf(entity).Elem()
for i, column := range columns {
	field := model.ColumnMap[column]
	entityValue.FieldByName(field.GoName).Set(values[i])
}
```

实际实现中必须检查：

- `entity` 是否为非空结构体指针。
- 列名是否存在于元数据中。
- 字段是否存在且 `CanSet()` 为 `true`。
- 临时值类型与目标字段类型是否一致。
- `rows.Scan`、`rows.Err()` 和 `rows.Close()` 是否正确处理。

反射方案语义清晰、安全性更好，适合先实现为默认方案。

### 3.3 使用 unsafe 处理结果集

`unsafe` 方案不再通过字段名动态定位字段，而是使用已缓存的字段偏移直接计算目标地址：

```go
entityAddress := reflect.ValueOf(entity).UnsafePointer()

for i, column := range columns {
	field := model.ColumnMap[column]
	fieldAddress := unsafe.Add(entityAddress, field.Offset)
	fieldPointer := reflect.NewAt(field.Typ, fieldAddress)
	scanTargets[i] = fieldPointer.Interface()
}

err := rows.Scan(scanTargets...)
runtime.KeepAlive(entity)
return err
```

这里不是“在目标地址额外创建了一个新对象”。`reflect.NewAt` 是使用已知类型解释现有内存地址，并生成指向该地址的反射值。`rows.Scan` 最终直接将数据写入模型字段。

`unsafe` 方案必须保证：

- 对象的实际类型与元数据一致。
- 偏移来自同一个结构体类型。
- 字段类型、对齐和内存边界正确。
- 对象在扫描期间保持存活。
- 代码通过 `go vet`、`checkptr` 和完整测试。

详细的指针、GC 与反射原理见 [Go 反射与 unsafe 学习笔记](go-reflection-unsafe.md)。

### 3.4 使用 unsafe 有什么优点

`unsafe` 的潜在优势是减少结果映射热路径中的一部分动态工作，例如：

- 根据字段名调用 `FieldByName`。
- 创建临时反射值，再二次赋值给结构体。
- 重复执行部分类型和可设置性检查。

但“使用 `unsafe` 一定更快、CPU 和内存消耗一定更少”并不严谨：

- `reflect.NewAt`、接口装箱和 `rows.Scan` 仍然有开销。
- 数据库驱动转换值的开销可能高于字段定位开销。
- 查询总时间可能主要消耗在网络和数据库执行上。
- 不同 Go 版本、驱动、数据类型和数据量会导致不同结果。

因此应使用 benchmark 同时比较执行时间和内存分配：

```bash
go test -bench=. -benchmem ./...
```

只有测量证明结果集映射是真实瓶颈时，才值得引入 `unsafe` 的额外复杂度。

### 3.5 ORM 的性能瓶颈在哪里

从 ORM 框架自身看，主要开销可以分为两类：

1. **SQL 构造**：遍历表达式、校验字段、拼接 SQL 并收集参数。
2. **结果集映射**：列名匹配、驱动类型转换、反射值创建和结构体赋值。

但对于一次完整的数据库请求，常见的更大瓶颈还包括：

- 数据库查询计划和 SQL 执行。
- 索引缺失或回表、排序等高成本操作。
- 锁等待和事务冲突。
- 应用与数据库之间的网络往返。
- 返回行数过多、查询了不需要的列。
- N+1 查询和过多的数据库往返。

优化时应先通过慢查询、执行计划、trace 和 profile 确认瓶颈，而不是默认认为 ORM 的反射就是最大问题。

#### SQL 构造如何优化

- 缓存模型元数据，避免重复反射解析。
- 使用 `strings.Builder` 或 `bytes.Buffer` 减少字符串拼接分配。
- 预分配参数切片容量。
- 避免重复构造和遍历相同的表达式。
- 对稳定查询考虑预编译语句，但需要评估驱动和数据库端的实际收益。

buffer pool 可能减少 Buffer 对象分配，但不是无条件的优化：池化对象需要正确重置，可能长期持有过大的底层内存，也可能增加管理成本。应在 profile 显示 SQL 构造分配确实显著时再考虑。

#### 结果集处理如何优化

- 缓存列名到字段元数据的映射。
- 缓存字段索引路径或字节偏移，避免每行调用 `FieldByName`。
- 仅查询实际需要的列，避免无条件使用 `SELECT *`。
- 使用批量查询和分页，控制单次返回数据量。
- 减少不必要的临时值、接口装箱和切片分配。
- 必要时再用 `unsafe` 优化已被证明的热路径。

### 3.6 面试回答模板

> ORM 处理结果集时，先获取列名，再根据模型元数据找到对应的 Go 字段和类型，为 `rows.Scan` 准备扫描目标。反射方案通常先扫描到临时值，再写入结构体；unsafe 方案可以根据对象起始地址和字段偏移获得字段指针，让 `Scan` 直接写入字段。unsafe 可能减少字段查找和临时值开销，但必须通过 benchmark 证明收益，并承担类型、对齐、边界和对象生命周期风险。ORM 性能瓶颈要分层看：框架内部包括 SQL 构造和结果集映射，完整请求则往往更受数据库执行、索引、锁等待、网络往返和返回数据量影响，应先测量再优化。

面试时可以记住 ORM 的三个核心：

```text
构造 SQL → 设计并缓存元数据 → 处理结果集
```

## 4. 学习路线

后续实现 ORM 框架时，可以重点学习：

```text
模型元数据与反射
        ↓
SQL 构造与参数绑定
        ↓
查询结果映射
        ↓
CRUD 与事务
        ↓
方言、钩子和缓存
```

ORM 本质上是对 SQL 构造和结果集处理的工程化封装。学习重点不是记忆链式 API，而是理解 Go 对象如何转换成 SQL，以及数据库结果如何恢复成 Go 对象。
