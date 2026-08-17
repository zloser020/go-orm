# MySQL 数据库学习笔记

[返回 ORM 框架学习笔记](orm框架.md)

## 1. 事务隔离级别

事务隔离级别用来规定并发事务之间的数据可见性。隔离性越强，并发异常越少，但通常也会付出更多的锁等待或并发性成本。

MySQL 支持四种标准隔离级别，从弱到强依次为：

```text
READ UNCOMMITTED
        ↓
READ COMMITTED
        ↓
REPEATABLE READ
        ↓
SERIALIZABLE
```

### 1.1 并发读异常

#### 脏读

一个事务读到了另一个事务尚未提交的修改。如果后者回滚，前者读到的就是从未真正生效的数据。

#### 不可重复读

同一事务使用相同条件多次读取同一行，由于其他事务在两次读取之间修改并提交了该行，导致两次读到的值不同。

#### 幻读

同一事务使用相同范围条件多次查询，由于其他事务插入或删除了符合条件的记录并提交，导致结果集的行数发生变化。

不可重复读关注的是“同一行的值变了”，幻读关注的是“符合范围条件的行变多或变少了”。

### 1.2 四种隔离级别

| 隔离级别 | 脏读 | 不可重复读 | 幻读 | 特点 |
| --- | --- | --- | --- | --- |
| 未提交读（`READ UNCOMMITTED`） | 可能 | 可能 | 可能 | 隔离性最弱，可读取其他事务未提交的修改 |
| 已提交读（`READ COMMITTED`） | 避免 | 可能 | 可能 | 每次一致性读通常可以看到该读操作开始前已提交的数据 |
| 可重复读（`REPEATABLE READ`） | 避免 | 避免 | SQL 标准下仍可能 | 事务内多次一致性读使用稳定的数据视图 |
| 串行化（`SERIALIZABLE`） | 避免 | 避免 | 避免 | 隔离性最强，事务效果如同串行执行，并发性最低 |

### 1.3 MySQL 默认隔离级别

MySQL InnoDB 的默认事务隔离级别是可重复读（`REPEATABLE READ`）。

InnoDB 主要通过 MVCC 提供一致性非锁定读；对于锁定读和范围更新，还会使用记录锁、间隙锁和 Next-Key Lock 等机制限制其他事务向相关范围插入记录。因此，MySQL InnoDB 在可重复读下对幻读的处理比 SQL 标准的最低要求更强。

但是，快照读和当前读的可见性与加锁方式不同，不应简单记成“可重复读在所有场景下都绝对不会出现幻读”。

### 1.4 设置隔离级别

MySQL 中可以设置会话后续事务的隔离级别：

```sql
SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;
```

Go 的 `database/sql` 可以通过 `sql.TxOptions` 指定隔离级别：

```go
tx, err := db.BeginTx(ctx, &sql.TxOptions{
	Isolation: sql.LevelRepeatableRead,
})
```

具体驱动和数据库是否支持某个隔离级别，需要根据其实现确认。ORM 的事务 API 最终也是将这类选项传递给底层数据库驱动。

### 1.5 如何选择

- 一般业务优先使用数据库默认级别，不要在没有具体问题时随意调整。
- 如果更希望每条查询看到已提交的最新数据，可考虑 `READ COMMITTED`。
- 如果事务内的多次读需要保持稳定视图，可使用 `REPEATABLE READ`。
- 只有在业务确实需要强串行化语义，且能接受锁等待和并发度下降时，才考虑 `SERIALIZABLE`。

隔离级别不会自动解决所有并发写问题。对于丢失更新、库存扣减等场景，还需要结合条件更新、乐观锁或 `SELECT ... FOR UPDATE` 等机制。

### 1.6 事务隔离级别面试要点

回答事务隔离级别问题时，可以按以下顺序展开：

1. MySQL 支持未提交读、已提交读、可重复读和串行化四种隔离级别。
2. 不同隔离级别主要考察脏读、不可重复读和幻读。
3. MySQL InnoDB 默认使用可重复读。
4. InnoDB 通过 MVCC 提供一致性读，并通过 Next-Key Lock 等锁机制防止锁定范围内插入新记录。
5. 如果继续追问底层实现，再说明 Read View、行记录的版本信息、undo log 和 redo log。

一个简短的面试回答可以是：

> MySQL 支持四种事务隔离级别，InnoDB 默认是可重复读。隔离级别主要解决脏读、不可重复读和幻读。InnoDB 的一致性读主要由 MVCC 实现，锁定读则结合记录锁、间隙锁和 Next-Key Lock 保护查询范围。

#### InnoDB 可重复读是否会出现幻读

面试中常见的简化结论是“InnoDB 在可重复读下不会出现幻读”，但完整回答需要区分两种读：

- **快照读**：普通 `SELECT` 通过 MVCC 读取一致性视图，事务内重复查询通常不会看到其他事务新提交的行。
- **当前读**：`SELECT ... FOR UPDATE`、`UPDATE` 和 `DELETE` 等需要读取最新已提交版本，InnoDB 在索引范围上使用 Next-Key Lock 阻止其他事务插入会影响结果集的记录。

因此，更准确的说法是：**InnoDB 在可重复读下，通过 MVCC 和 Next-Key Lock 分别处理快照读和锁定读中的幻读问题。**

### 1.7 MVCC、undo log 与 redo log

#### MVCC

MVCC 是 Multi-Version Concurrency Control，即多版本并发控制。它使同一行数据在逻辑上可以存在多个历史版本，读事务根据可见性规则选择自己能看到的版本，从而降低读写之间的锁冲突。

InnoDB 的行记录包含用于版本判断的隐藏信息，其中重要的有：

- 最近修改该行的事务 ID。
- 指向 undo log 中上一个历史版本的回滚指针。

```text
当前行版本
    │ 回滚指针
    ↓
undo log 中的上一版本
    │
    ↓
更早的行版本
```

#### Read View

Read View 是一致性读用来判断数据版本是否可见的快照。它会记录创建快照时系统中的活跃事务信息。读取某行时，InnoDB 将行版本的事务 ID 与 Read View 比较：

- 如果当前版本可见，就直接返回。
- 如果当前版本不可见，就沿 undo log 版本链查找可见的历史版本。

`READ COMMITTED` 和 `REPEATABLE READ` 的关键区别之一是 Read View 的创建时机：

- `READ COMMITTED` 通常在每次一致性读时创建新的 Read View，因此后续查询可以看到其他事务新提交的数据。
- `REPEATABLE READ` 通常在事务的第一次一致性读时创建 Read View，后续一致性读复用它，因此能够保持可重复读。

#### undo log

undo log 保存用于撤销数据修改的信息，主要有两个作用：

1. **事务回滚**：事务失败或主动回滚时，根据 undo log 将数据恢复到修改前的状态。
2. **MVCC 历史版本**：一致性读可以沿版本链找到对当前事务可见的历史数据。

#### redo log

redo log 记录 InnoDB 对数据页所做的物理修改，用于崩溃恢复和保证已提交事务的持久性。数据页可以先在内存中修改，再延后写回数据文件；只要对应 redo log 已按策略持久化，数据库异常重启后就可以重放日志恢复修改。

| 机制 | 主要职责 | 关键作用 |
| --- | --- | --- |
| MVCC | 利用多版本和可见性规则完成一致性读 | 降低读写冲突 |
| undo log | 保存撤销信息和历史版本 | 事务回滚、MVCC |
| redo log | 记录对数据页的修改 | 崩溃恢复、持久性 |

可以简化记忆为：

```text
MVCC：决定读哪个版本
undo log：提供旧版本并支持回滚
redo log：保证已提交修改在崩溃后可恢复
```

## 2. Go ORM 的事务设计

事务相关的 ORM 面试题通常分为两层：数据库层需要理解隔离级别、MVCC 和锁；ORM 层则需要说明事务如何在业务调用链中传递，以及如何安全地提交或回滚。

### 2.1 什么是事务传播

事务传播描述的是：一个业务方法被调用时，如果调用链中已经存在事务，当前方法应该加入原事务、创建新事务，还是以非事务方式运行。

常见语义包括：

| 传播语义 | 已存在事务 | 不存在事务 |
| --- | --- | --- |
| `REQUIRED` | 加入当前事务 | 开启新事务 |
| `SUPPORTS` | 加入当前事务 | 非事务运行 |
| `MANDATORY` | 加入当前事务 | 返回错误 |
| `REQUIRES_NEW` | 暂时不用外层事务，开启新事务 | 开启新事务 |
| `NOT_SUPPORTED` | 暂时不用当前事务 | 非事务运行 |
| `NEVER` | 返回错误 | 非事务运行 |
| `NESTED` | 建立嵌套事务或保存点 | 开启新事务 |

图片中提到的“有事务就复用，没有事务就新建”，对应的就是 `REQUIRED`。如果没有现有事务，也并非只能新建事务；具体行为取决于 ORM 提供的传播策略和业务要求。

需要注意，Go 标准库 `database/sql` 只提供事务原语，并没有直接定义上述传播级别。传播规则通常由 ORM 或业务框架自行实现；`NESTED` 等能力还依赖数据库和驱动是否支持保存点。

### 2.2 Go 中如何传递事务

Go 没有线程本地存储可以可靠地表示一条业务调用链，因为 goroutine 可能切换执行线程。常见做法是将事务与 `context.Context` 关联，并沿调用链传递：

```text
业务入口
   │
   ├── context 中已有事务 ──→ 按传播策略复用或拒绝
   │
   └── context 中没有事务 ──→ 开启事务、非事务运行或报错
```

一种简化实现如下：

```go
type txKey struct{}

func contextWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func txFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}
```

使用时应遵守以下原则：

- key 使用包内私有类型，避免与其他包的 context key 冲突。
- context 只沿调用链传递，不要保存在结构体中长期复用。
- 事务对象只应在其生命周期内使用，不要泄漏给异步 goroutine。
- context 中存事务虽然便于实现传播，但会形成隐藏依赖；简单系统也可以显式传递 `*sql.Tx`，或者让仓储依赖统一的 `Executor` 接口。

例如，可以让 `*sql.DB` 和 `*sql.Tx` 共同满足同一个最小接口：

```go
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}
```

这样 SQL 构造器不需要关心当前执行者究竟是数据库连接池还是事务。

### 2.3 重复提交或回滚会怎样

一个事务只能成功结束一次。调用 `Commit` 或 `Rollback` 后，再次调用通常会得到 `sql.ErrTxDone`。因此 ORM 可以维护事务状态，提前返回更清晰的错误，但不能仅依赖一个布尔标记代替底层数据库事务的状态判断。

还要注意以下情况：

- `Commit` 返回错误时，不能假定事务一定提交成功，也不能安全地直接重试整个 `Commit`。
- `Rollback` 在事务已经提交或回滚后返回 `sql.ErrTxDone`，通常不是新的业务故障。
- 事务对象不应该被无约束地并发提交、回滚或继续执行 SQL。

### 2.4 如何实现事务闭包

闭包式事务 API 的目标是统一管理 `Begin`、`Commit` 和 `Rollback`，确保业务返回错误或发生 panic 时不会遗漏回滚。

```go
func InTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}

		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				err = errors.Join(err, rbErr)
			}
			return
		}

		err = tx.Commit()
	}()

	err = fn(tx)
	return err
}
```

这段代码体现了事务闭包的三个核心规则：

1. 业务函数返回 `nil` 才提交事务。
2. 业务函数返回错误时回滚，并保留原始业务错误。
3. 业务函数发生 panic 时先回滚，再重新抛出 panic，不能悄悄吞掉异常。

生产实现还需要考虑事务选项、日志和监控、回滚错误、context 取消、嵌套调用以及传播级别等问题。

### 2.5 面试回答模板

> 事务传播决定业务方法遇到已有事务或没有事务时应该如何执行。例如 `REQUIRED` 会复用已有事务，否则开启新事务；也可以根据业务选择报错或非事务运行。Go 没有适合业务调用链的 thread-local，ORM 通常通过 `context.Context` 传递事务，或者显式传递统一的数据库执行接口。事务闭包需要保证正常返回时提交，业务报错时回滚，发生 panic 时回滚后继续抛出。重复提交或回滚时，`database/sql` 通常返回 `sql.ErrTxDone`，ORM 可以额外维护状态以提供更清晰的错误信息。
