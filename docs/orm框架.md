# ORM 框架学习笔记

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

## 2. SELECT 起步

### 2.1 GORM 如何构造 SQL

GORM 中主要涉及 `Builder`、`Expression`、`Clause` 和 `Interface` 四个抽象。核心思路是将 SQL 的不同部分分别构造，最后再拼接成完整语句。

```text
SELECT 部分 + FROM 部分 + WHERE 部分 + 其他子句 → 完整 SQL
```

当前项目也采用类似思路：`Selector` 负责组织 SELECT 查询，`Predicate` 表示查询条件，`BuildExpression` 递归构造条件表达式。

### 2.2 Builder 模式

Builder（建造者）模式将复杂对象的构造过程拆成多个步骤，使调用方不必一次提供全部参数，而是逐步设置需要的部分，最后统一生成目标对象。

它适合具有以下特点的对象：

- 构造步骤较多。
- 包含大量可选参数。
- 不同参数之间存在组合关系。
- 希望使用链式 API 提高可读性。

SQL 正好符合这些特点。一条 SELECT 语句可能包含 `FROM`、`WHERE`、`GROUP BY`、`ORDER BY` 和 `LIMIT` 等部分，而且多数部分都是可选的。

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

需要注意，链式调用不是 Builder 模式的必要条件，只是一种常见写法。Builder 模式的关键是逐步收集构造信息，最后通过 `Build` 生成完整对象。

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

## 3. 反射：获取并调用方法

### 3.1 方法接收者

使用反射遍历方法时，需要注意方法接收者：

- 输入是结构体值 `T`，只能获取值接收者 `func (t T) Method()` 的方法。
- 输入是结构体指针 `*T`，可以获取值接收者和指针接收者的方法。
- 通过 `reflect.Type.Method` 得到的方法函数，其第一个输入参数永远是接收者本身，后面才是方法声明的参数。
- `NumMethod` 和 `Method` 只能获取导出方法。

```text
输入 User  → 获取 User 的值接收者方法
输入 *User → 获取 User 和 *User 的方法
```

例如：

```go
func (u User) GetAge() int
func (u *User) ChangeName(name string)
```

传入 `User` 时只能找到 `GetAge`；传入 `*User` 时可以同时找到 `GetAge` 和 `ChangeName`。`ChangeName` 的反射输入参数依次为接收者 `*User` 和方法参数 `string`。

### 3.2 反射编程技巧

- 读写具体数据使用 `reflect.Value`，读取类型信息使用 `reflect.Type`。
- `T` 和 `*T` 在反射中是两种类型，操作前要确认是否为指针；通常使用 `Elem()` 获取指针指向的值或类型。
- 指针类型主要用于判断指向类型和获取指针方法集；分析结构体字段时，一般操作其 `Elem()`。
- 反射 API 经常在类型不匹配、值不可修改或操作不支持时触发 panic，因此需要充分测试，并在调用前使用 `Kind`、`CanSet` 等方法检查。
- 数组和切片分别对应 `reflect.Array` 和 `reflect.Slice`，不能当成同一种类型判断。
- 字段和方法需要使用不同的反射 API，例如 `Field` 与 `Method`。

#### Value、Type 与原始对象的关系

```text
&user
  └── reflect.ValueOf(&user)       指针 Value
          ├── Type()               → *User 的 Type
          └── Elem()               → User 的 Value，可读写字段
                  └── Type()       → User 的 Type

reflect.TypeOf(&user)              *User 的 Type
  └── Elem()                       → User 的 Type
```

核心关系如下：

- `reflect.ValueOf` 得到值的反射表示，用于读取或修改数据。
- `reflect.TypeOf` 得到类型的反射表示，只描述类型信息。
- `Value.Type()` 可以从 Value 获取对应的 Type。
- 对指针调用 `Elem()`，可以从 `*User` 进入其指向的 `User`。
- 若要修改结构体字段，应传入 `&user`，再通过 `ValueOf(&user).Elem()` 获得可设置的结构体 Value。

### 3.3 Map 遍历顺序

Go 的 map 是无序的，`MapRange` 和 `MapKeys` 都不保证返回顺序。因此测试 map 遍历结果时，不应直接比较切片顺序，而应比较完整的键值关系；如果业务需要固定顺序，则必须先对 key 排序。

### 3.4 反射面试要点

#### 什么是反射

反射是程序在运行期间描述类型和值，并间接读取或操作对象的能力。Go 主要通过 `reflect.Type` 获取类型信息，通过 `reflect.Value` 操作具体值。

#### 反射有哪些使用场景

反射常用于无法在编译期确定具体类型的通用框架，例如：

- ORM 的模型与数据库字段映射。
- JSON 等序列化和反序列化。
- 依赖注入和配置解析。
- Web 框架中的参数绑定。

#### 能否通过反射修改方法

不能。Go 的反射 API 可以查找和调用方法，但不能修改方法实现，Go runtime 也没有提供相应接口。

#### 什么样的字段可以被反射修改

可以使用 `CanSet()` 判断值能否修改。通常需要传入对象指针，再通过 `Elem()` 得到可寻址的结构体值；字段还必须是可设置的导出字段。

```go
val := reflect.ValueOf(&user).Elem()
Field := val.FieldByName("Name")
if Field.CanSet() {
	Field.SetString("Tom")
}
```

直接传入结构体值通常只能读取，不能修改：

```text
reflect.ValueOf(user)         → 通常不可设置
reflect.ValueOf(&user).Elem() → 可寻址，再通过 CanSet 判断
```

## 4. 学习要点

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
