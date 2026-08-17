# Go 反射与 unsafe 学习笔记

[返回 ORM 框架学习笔记](orm框架.md)

## 1. Go 反射

### 1.1 方法接收者

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

### 1.2 反射编程技巧

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

### 1.3 Map 遍历顺序

Go 的 map 是无序的，`MapRange` 和 `MapKeys` 都不保证返回顺序。因此测试 map 遍历结果时，不应直接比较切片顺序，而应比较完整的键值关系；如果业务需要固定顺序，则必须先对 key 排序。

### 1.4 反射面试要点

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

## 2. Go unsafe：unsafe.Pointer 与 uintptr

`unsafe.Pointer` 和 `uintptr` 都可以间接表示内存地址，但它们的语义完全不同：

- `unsafe.Pointer` 仍然是 Go 指针，指向某块内存，垃圾回收器会将它视为指针。
- `uintptr` 是一个无符号整数，宽度足以容纳当前平台的指针位模式，但 GC 只会把它当成数字。

| 对比项 | `unsafe.Pointer` | `uintptr` |
| --- | --- | --- |
| 类型本质 | 指针类型 | 无符号整数类型 |
| GC 是否当作指针 | 是 | 否 |
| 能否保持对象存活 | 可以参与对象的可达性判断 | 不能 |
| 能否直接做算术运算 | 不能 | 能 |
| 常见用途 | 在不同指针类型之间转换 | 短暂用于地址偏移计算或与底层 API 交互 |

### 2.1 unsafe.Pointer

`unsafe.Pointer` 类似 C 语言的 `void*`，可以在不同类型的指针之间作为转换桥梁：

```go
var n int64 = 10
p := &n

up := unsafe.Pointer(p)
ip := (*int64)(up)
fmt.Println(*ip)
```

常见转换路径是：

```text
*T ↔ unsafe.Pointer ↔ *U
```

`unsafe.Pointer` 不包含被指向数据的静态类型信息。程序员必须自己保证转换后的类型、对齐、大小和对象生命周期都正确，否则可能造成内存破坏或难以定位的错误。

### 2.2 uintptr

`uintptr` 是整数，不是指针。即使它的数值来自某个内存地址，GC 也不会因此认为它引用了原对象。

```go
p := &value
addr := uintptr(unsafe.Pointer(p))
```

此时 `addr` 只是一个数字。如果程序只保存 `addr` 而不再保留任何 Go 指针，原对象可能被判定为不可达。因此不应将 Go 对象的地址转换成 `uintptr` 后长期保存，再在以后的任意时刻转回指针。

### 2.3 unsafe.Pointer 与 GC

需要区分“GC 能够识别指针”和“GC 会移动对象”这两件事。

#### unsafe.Pointer 的 GC 语义

GC 会将可达的 `unsafe.Pointer` 当作指针扫描。因此，当它合法地指向一个 Go 对象时，可以使该对象继续保持可达，避免它被当作垃圾回收。

```text
GC Roots
   │
   └── unsafe.Pointer ──→ Go 对象
                            ↑
                         仍然可达
```

`uintptr` 没有这种语义：

```text
uintptr(0xAAAA) ──→ GC 只看到整数，不认为它引用了对象
```

因此，两者最重要的差异不是“数值是否像地址”，而是 **GC 是否把这个值当作指针追踪**。

#### 当前 Go 堆 GC 不会复制搬移存活对象

图中“Go GC 是标记—复制，对象从 `0xAAAA` 移动到 `0xAABB`，然后 GC 修改 `unsafe.Pointer`”的说法不符合当前 Go 堆 GC 的实现。

当前 Go runtime 将堆 GC 描述为：

- 并发标记—清扫（concurrent mark and sweep）。
- 非分代（non-generational）。
- 非压缩（non-compacting）。

所以，常规堆 GC 不会为了回收空间而将一个存活堆对象复制到新地址。不能使用图中的 `0xAAAA → 0xAABB` 过程来解释当前 Go 堆 GC。

```text
当前 Go 堆 GC：

标记可达对象 → 清扫不可达对象
                 ≠
       复制所有存活对象到新地址
```

#### 为什么仍然不能用 uintptr 长期保存指针

“当前堆 GC 不压缩”不等于“程序可以把 `uintptr` 当指针使用”。仍然不能这样做，主要原因是：

1. `uintptr` 不保持对象存活，对象可能被回收，原地址可能被重新利用。
2. Go 的 `unsafe` 规则不保证任意的 `uintptr → unsafe.Pointer` 转换合法。
3. 官方 API 语义明确保留了“如果对象移动，GC 不会更新 `uintptr`”这一边界，代码不应依赖当前 GC 恰好不压缩的实现细节。

例如，下面的写法仍然是错误的：

```go
p := new(int)
addr := uintptr(unsafe.Pointer(p))
p = nil

runtime.GC()
q := (*int)(unsafe.Pointer(addr)) // 非法：addr 没有保持原对象存活
```

#### runtime.KeepAlive 的作用

当底层操作使用了编译器难以看出生命周期的指针或系统调用时，可以在最后一次底层访问之后调用 `runtime.KeepAlive`：

```go
p := &value
usePointer(unsafe.Pointer(p))
runtime.KeepAlive(p)
```

`runtime.KeepAlive(p)` 表示 `p` 引用的对象至少在该调用之前保持可达。它不会将非法的 `uintptr` 转换变成合法操作，也不是通用的“禁止 GC”方法。

### 2.4 指针偏移计算

`uintptr` 的常见误区是：因为它能容纳地址数值，就把它当作可以长期保存的指针。实际上，`uintptr` 更适合表示“相对数量”，例如字段偏移、对齐值和内存大小。

```go
type FieldMeta struct {
	offset uintptr // 字段相对于结构体起始位置的字节数
}
```

这里的 `offset` 不是一个对象地址，而是一个数值型的相对距离：

```text
结构体起始地址 + 字段偏移 = 字段地址

unsafe.Pointer        uintptr       unsafe.Pointer
   指针语义          数值语义          指针语义
```

字段偏移本身不引用任何 Go 对象，也不需要保持对象存活，所以将它保存为 `uintptr` 是合理的。GC 回收不可达对象不会改变某个已编译结构体类型的字段布局，因此 GC 不会修改 `reflect.StructField.Offset` 这类相对量。

但“字段偏移不受 GC 影响”不等于“字段偏移在任何情况下永远不变”。偏移依赖于：

- 结构体字段的声明顺序和类型。
- 编译目标的指针宽度和对齐规则。
- 嵌套类型和字段本身的内存布局。
- Go 工具链和底层实现所保证的布局规则。

因此，偏移应该由当前程序使用 `unsafe.Offsetof` 或 `reflect.StructField.Offset` 获取，不应使用手写魔法数字，也不应将其序列化后假设可以在不同程序、类型版本或架构间通用。

例如：

```go
type User struct {
	ID   int64
	Name string
}

nameOffset := unsafe.Offsetof(User{}.Name)
```

`nameOffset` 表示 `Name` 相对于 `User` 起始位置的字节偏移，而不是某一个 `User` 实例的地址。

#### 仅在地址运算中短暂经过 uintptr

传统写法会将 `unsafe.Pointer` 短暂转换成 `uintptr` 进行地址计算，然后立即转回指针：

```go
fieldPtr := unsafe.Pointer(
	uintptr(entityPtr) + fieldOffset,
)
```

这些转换应当位于同一个表达式中。不应将中间的 `uintptr` 保存到字段或全局变量中：

```go
// 危险：将地址当成整数跨越了时间保存
addr := uintptr(entityPtr) + fieldOffset
// 经过其他调用或 GC 后再转回指针
fieldPtr := unsafe.Pointer(addr)
```

现代 Go 代码中可以优先使用 `unsafe.Add`，避免显式经过 `uintptr`：

```go
fieldPtr := unsafe.Add(entityPtr, fieldOffset)
```

`unsafe.Add` 的偏移单位是字节。计算结果必须仍然落在合法的内存对象范围内，不能借此访问任意地址。

可以使用下面的原则快速判断：

```text
保存字段偏移、大小、对齐值 → uintptr 合理
持有某个 Go 对象的地址       → 使用指针，不要用 uintptr
计算基地址加偏移             → 优先 unsafe.Add
```

### 2.5 ORM 中的应用

ORM 可以在解析模型时使用 `reflect.StructField.Offset` 缓存字段相对于结构体起始地址的偏移，以后通过对象地址与偏移快速找到字段：

```go
type FieldMeta struct {
	offset uintptr
}

func fieldAddress(entity unsafe.Pointer, field FieldMeta) unsafe.Pointer {
	return unsafe.Add(entity, field.offset)
}
```

这种做法可能比每次通过反射查找字段更快，但会放弃 Go 类型安全保证。实现时至少需要保证：

- `entity` 确实指向解析元数据时对应的结构体类型。
- 偏移来自同一结构体类型的 `StructField.Offset`。
- 字段指针按字段的真实类型转换和读写。
- 不读写对象边界之外的内存。
- 原对象在底层指针使用期间保持存活。必要时可在最后一次底层访问后调用 `runtime.KeepAlive(entity)` 明确生命周期。

### 2.6 常见错误

#### 把 uintptr 当成真正的指针

```go
type Accessor struct {
	address uintptr
}
```

如果 `address` 是 Go 堆对象的唯一“引用”，它不能保证对象存活。需要持有指针语义时，应保存合法的 Go 指针或 `unsafe.Pointer`，而不是 `uintptr`。

#### 根据偏移访问错误的类型

相同数值的偏移不代表不同结构体布局完全一致。结构体的字段顺序、类型和对齐都会影响内存布局。

#### 指针越界

指针算术必须限制在原对象的合法范围内。即使计算后的地址在当前运行中“看起来可用”，也不代表符合 Go 的 `unsafe` 规则。

#### 依赖未保证的内存布局

`unsafe` 代码常与 Go 版本、架构、对齐规则和底层实现细节耦合。优化前应先基准测试，并将 `unsafe` 代码集中封装、充分测试，避免在业务代码中扩散。

### 2.7 面试总结

可以这样回答：

> `unsafe.Pointer` 是 Go 指针，GC 会将可达的 `unsafe.Pointer` 当作指针扫描，它主要用于不同指针类型之间的转换。`uintptr` 是可以容纳地址数值的无符号整数，GC 不会将它当作对象引用。当前 Go 堆 GC 是非压缩的并发标记—清扫，不会像标记—复制 GC 那样常规搬移存活堆对象；但这不改变 `uintptr` 缺乏指针语义的事实。需要进行地址偏移时可以短暂经过 `uintptr`，但不应将 Go 对象的地址以 `uintptr` 形式长期保存；新代码可优先使用 `unsafe.Add`。

一句话记忆：**`unsafe.Pointer` 有指针语义，`uintptr` 只有数字语义。**

### 2.8 unsafe 面试要点

#### uintptr 和 unsafe.Pointer 有什么区别

`uintptr` 是能够容纳指针位模式的无符号整数。它可以保存某个地址的数值，但没有指针语义，GC 不会通过它追踪对象，它也不能保持对象存活。

`unsafe.Pointer` 是指针类型，可以作为不同 Go 指针类型之间的转换桥梁。GC 会将可达的 `unsafe.Pointer` 当作指针扫描。

```text
uintptr       = 地址的数字表示，没有引用语义
unsafe.Pointer = Go 指针，具有指针和 GC 语义
```

不应简单说“Go runtime 永远会把 `unsafe.Pointer` 修正到对象的新地址”。当前 Go 堆 GC 是非压缩的并发标记—清扫，并不会常规搬移存活堆对象。但 Go 仍不保证可以将 `uintptr` 当作长期指针使用。

#### Go 对象是如何对齐的

“按照字长对齐”是过度简化的说法。更准确地说，每种类型都有自己的对齐要求：

- 字段起始地址需要满足该字段类型的对齐要求。
- 编译器可能在字段之间插入填充字节。
- 结构体的对齐值通常受其字段中最大对齐要求影响。
- 结构体总大小可能包含尾部填充，以便结构体数组中的每个元素都正确对齐。

```go
type Layout struct {
	A byte  // 1 字节
	B int64 // 前面可能存在填充
	C byte
}

fmt.Println(unsafe.Alignof(Layout{}))
fmt.Println(unsafe.Sizeof(Layout{}))
fmt.Println(unsafe.Offsetof(Layout{}.B))
```

可以使用以下 API 检查当前编译目标下的布局：

| API | 作用 |
| --- | --- |
| `unsafe.Sizeof(x)` | 获取 `x` 的大小，包含结构体填充 |
| `unsafe.Alignof(x)` | 获取 `x` 所需的对齐值 |
| `unsafe.Offsetof(s.Field)` | 获取字段相对于结构体起始位置的偏移 |

字段顺序会影响填充和结构体总大小。但不应仅为节省少量内存就随意重排对外数据结构；还需要考虑 API 语义、缓存局部性、原子操作对齐以及序列化兼容性。

#### 如何计算对象和字段地址

如果已经有一个具体结构体变量，可以直接获取它的指针，不需要经过反射：

```go
user := User{}
base := unsafe.Pointer(&user)
```

只有在 ORM 这类通用框架中，当编译期不知道具体模型类型时，才常使用反射验证参数并取得底层指针。例如可先要求输入是非空的结构体指针：

```go
value := reflect.ValueOf(entity)
if value.Kind() != reflect.Ptr || value.IsNil() ||
	value.Elem().Kind() != reflect.Struct {
	return errors.New("entity must be a non-nil struct pointer")
}

base := value.UnsafePointer()
```

使用反射的 `Pointer`、`UnsafePointer` 或 `UnsafeAddr` 时，必须严格满足对应方法对 `Kind`、可寻址性和生命周期的要求，否则可能 panic 或产生非法指针。

字段地址可以由结构体起始指针和字段偏移得到：

```go
offset := unsafe.Offsetof(user.Name)
fieldPtr := unsafe.Add(base, offset)
```

在 ORM 中，偏移可以在解析模型元数据时通过 `reflect.StructField.Offset` 缓存，但必须确保实际对象与该元数据属于同一结构体类型。

#### unsafe 为什么可能比反射快

反射是通用的运行时抽象。在一些热路径中，它可能包含：

- 运行时的类型和 `Kind` 检查。
- `reflect.Value` 的封装与方法调用。
- 根据字段名或索引路径查找字段。
- 值可设置性、间接引用和类型转换检查。

如果 ORM 已在模型注册阶段验证类型并缓存字段偏移，读写热路径就可以简化为：

```text
对象起始指针 + 已缓存偏移 → 字段指针 → 按已知类型读写
```

这可以减少重复的反射查找和检查开销，但不能简单推导为“所有 `unsafe` 代码都比反射快”。真实结果取决于访问模式、编译器优化、缓存局部性和封装方式，应使用 benchmark 测量。

```go
func BenchmarkReflectAccess(b *testing.B) {
	// 反射实现
}

func BenchmarkUnsafeAccess(b *testing.B) {
	// unsafe 实现
}
```

使用 `unsafe` 是以放弃部分类型安全、GC 安全检查和可维护性为代价的。在 ORM 等框架中，应先使用反射实现正确版本，只在基准测试证明某条热路径确实值得优化时，才将 `unsafe` 限制在小而独立的内部封装中。

#### 面试简短回答

> `uintptr` 是可以容纳地址数值的整数，没有指针和 GC 语义；`unsafe.Pointer` 是 Go 指针，可以在不同指针类型之间转换。Go 类型按各自的对齐要求布局，可用 `Sizeof`、`Alignof` 和 `Offsetof` 检查。字段地址可由对象起始指针加字段偏移得到，新代码优先使用 `unsafe.Add`。`unsafe` 可以减少反射热路径中的动态查找和检查，但是否更快必须通过 benchmark 验证，并且要承担更高的安全和维护成本。

## 3. 反射与 unsafe 如何选择

反射和 `unsafe` 都能支持编译期未知类型的通用框架，但它们处于不同抽象层次：

| 对比项 | 反射 | unsafe |
| --- | --- | --- |
| 核心能力 | 在运行时读取类型、字段、方法和值 | 绕过类型系统直接解释和访问内存 |
| 安全性 | 保留运行时类型与可设置性检查 | 由程序员自行保证类型、对齐、边界和生命周期 |
| 可维护性 | 较高，语义清晰 | 较低，容易依赖实现和内存布局 |
| 常见开销 | 动态查找、包装和运行时检查 | 可减少热路径检查，但必须用 benchmark 验证 |
| ORM 用途 | 解析模型、Tag、字段类型和方法 | 根据缓存偏移快速访问已验证对象的字段 |

推荐的实现顺序是：

```text
先用反射解析并验证模型
        ↓
缓存字段类型、索引路径和偏移
        ↓
优先使用反射完成正确实现
        ↓
benchmark 发现明确热路径
        ↓
仅在内部小范围使用 unsafe 优化
```

二者可以组合，但职责应当分离：

- 注册阶段使用反射建立可信的模型元数据。
- 普通路径继续使用反射，获得更好的安全性和可维护性。
- 确有性能需求时，底层访问器使用已经验证的类型和偏移执行 `unsafe` 访问。
- 对外 API 不暴露 `unsafe.Pointer`，将风险限制在可测试的内部包中。
- 为反射版本和 unsafe 版本编写相同的行为测试，并使用 `go vet`、`checkptr` 和 benchmark 验证。

一句话概括：**反射解决“运行时如何理解类型”，unsafe 解决“确认布局后如何直接访问内存”；先保证正确，再考虑绕过抽象层。**
