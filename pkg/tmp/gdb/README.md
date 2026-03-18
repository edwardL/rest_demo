# utils/gdb（模型友好版）

`utils/gdb` 是本仓库的数据库核心包，提供 3 套能力：

- `Db`：运行 SQL 并拿结果（链式 ORM 风格）
- `Sql`：仅构建 SQL（不执行）
- `DbT[T]`：泛型版 `Db`（强类型返回）

## 1) 初始化

```go
package main

import (
	hhdb "nwgit.gzhhit.com/BD/hhitdb.git"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/gdb"
)

func main() {
	var dbConf = map[string]string{
		"db_host": "127.0.0.1",
		"db_port": "3306",
		"db_user": "root",
		"db_pass": "***",
		"db_name": "test",
		"db_charset": "utf8mb4",
		"db_type": "mysql",
		"need_mode_onebyone": "1",
	}
	var err = hhdb.InitDB(dbConf["db_type"], &dbConf)
	if err != nil {
		panic(err)
	}
	gdb.Init(
		gdb.WithLogLevel(gdb.InfoLogLevel),
		gdb.WithWriteErrSql(true),
	)
}
```

## 2) 最小使用路径

### `Db` 查询

```go
var users []User
var res = gdb.New(&User{}).
	Where("age > ?", 18).
	Order("id DESC").
	Page(1, 20).
	Select(&users)
if res.Error != nil {
	panic(res.Error)
}
```

### `Sql` 只构建

```go
var r, err = gdb.NewSql("users").Where("id = ?", 1).Select()
if err != nil {
	panic(err)
}
_ = r.CompSql()
```

### `DbT[T]` 泛型

```go
var list, err = gdb.Model[User]().Where("age > ?", 18).Select()
if err != nil {
	panic(err)
}
_ = list
```

## 3) 全部导出 API（完整索引）

说明：本节按“入口函数/类型/方法族”组织，便于大模型检索调用。

### A. 顶层常量与变量

- `ExecTypeInsert` `ExecTypeUpdate` `ExecTypeDelete` `ExecTypeReplace`
- `FieldRemove`
- `ErrRecordNotFound`

### B. 顶层函数

- 初始化与日志：`Init` `SetDefLog` `SetLevel` `SetTraceFunc` `SetDebugLogWriter` `SetInfoLogWriter` `SetWarnLogWriter`
  `SetErrorLogWriter`
- 构建入口：`New` `NewCtx` `NewSql` `Model`
- 轻量查询执行：`Query` `QueryContext` `QueryOne` `QueryOneContext` `Exec` `ExecContext`
- 事务执行版：`TQuery` `TQueryContext` `TQueryOne` `TQueryOneContext` `TExec` `TExecContext`
- 轻量写入：`Create` `CreateContext` `CreateInBatches` `CreateInBatchesContext` `Update` `UpdateContext` `Save`
  `SaveContext` `SaveInBatches` `SaveInBatchesContext` `Replace` `ReplaceContext` `ReplaceInBatches`
  `ReplaceInBatchesContext`
- 事务轻量写入：`TCreate` `TCreateContext` `TCreateInBatches` `TCreateInBatchesContext` `TUpdate` `TUpdateContext`
  `TSave` `TSaveContext` `TSaveInBatches` `TSaveInBatchesContext` `TReplace` `TReplaceContext` `TReplaceInBatches`
  `TReplaceInBatchesContext`
- 事务入口：`Begin`
- 工具：`MapToMapAny` `MapsToMapsAny` `StructDbField` `Dump` `DumpCtx` `InArr` `As` `EscapeLike` `GetTableName`
  `ValidateOrderParam`
- 错误：`NewDbErr` `ConvDbErr` `IsRecordNotFound`
- 原始片段：`Raw` `RawAny`

### C. 配置 Option

- `WithLog`
- `WithLogLevel`
- `WithWriteHhDbLog`
- `WithWriteLog`
- `WithTraceIdFunc`
- `WithAppendDbKeywords`
- `WithAppendDbFieldChar`
- `WithAppendZeroValIgnoreField`
- `WithWriteErrSql`
- `WithWriteCompSql`
- `WithEmptyError`
- `WithDbConvInitPtr`
- `WithTimeLocation`
- `WithDriveType`
- `WithDriveMap`

### D. 主要类型

- `type Db`
- `type Sql`
- `type DbT[T]`
- `type DbTx`
- `type OrmResult` `type TOrmResult[T]` `type ExecResult`
- `type DbError`
- `type Conf` `type Option`
- `type CteQuery`
- `type Wrapper` `type SqlWrapperFace` `type NewWrapperFunc`
- `type BaseMode` `type BaseModeTime`
- `type Page[T]`
- `type OrderedMap[K,V]` `type KeyValue[K,T]`
- `type JSONType[T]` `type ENUMType[T]`
- `type HhSearchParam`

### E. `Db` 方法族（高频）

- **条件构建**：`Where` `WhereOr` `WhereCond` `WhereGroup` `WhereGroupOr` `WhereGroupCond` `WhereBlock` `WhereBlockOr`
  `WhereReset` `Eq` `NotEq` `In` `Like` `LikeRaw` `Gt` `Ge` `Lt` `Le` `Between` `IsNotNull`
- **字段与排序**：`Field` `Fields` `FieldsByData` `FieldsByDataAlias` `OmitFields` `Group` `Groups` `Order` `Orders`
  `OrderByFilter`
- **分页与表**：`Table` `As` `Limit` `Limits` `Offset` `Page` `PageReset`
- **Join/Union**：`Join` `Joins` `LeftJoin` `RightJoin` `InnerJoin` `Union` `UnionAll`
- **查询执行**：`Map` `Maps` `Slices` `OrderMap` `OrderMaps` `Pluck` `Count` `GetCount` `Exists` `Select` `One` `Scan`
  `SelectAndCount` `ScanAndCount` `TypeMap` `TypeMaps` `Stream`
- **写入执行**：`Create` `CreateInBatches` `Update` `Updates` `UpdateIgnore` `Save` `SaveInBatches` `Replace`
  `ReplaceInBatches` `Delete` `RawExec` `RawExecType`
- **上下文/事务/日志**：`Ctx` `Tx` `Transaction` `LogLevel` `LogCallDepth` `WriteLog` `WriteHhDbLog` `WriteErrSql`
  `WriteCompSql`
- **高级控制**：`SetDuplicate` `IgnoreDuplicate` `ReplaceFields` `ConvFieldsType` `EmptyError` `WhereSearch`
  `WhereSearchParams` `WhereSearchView` `CteQuery`

### F. `Sql` 方法族（高频）

- 条件与结构：`Where` `WhereOr` `WhereGroup` `WhereGroupOr` `WhereReset` `Join` `Joins` `LeftJoin` `RightJoin`
  `InnerJoin` `Union` `UnionAll`
- 字段分页：`Fields` `FieldsByData` `FieldsByDataAlias` `OmitFields` `Group` `Order` `OrderByFilter` `Limit` `Offset`
  `Page` `PageReset`
- 表与上下文：`Table` `As` `Ctx`
- 生成 SQL：`Select` `Count` `Exists` `Create` `CreateInBatches` `Update` `Updates` `UpdateIgnore` `Save` `SaveInBatches`
  `Replace` `ReplaceInBatches` `Delete` `RawExec`
- 原始与控制：`Raw` `SetDuplicate` `IgnoreDuplicate` `ReplaceFields` `WhereSearch` `WhereSearchView` `CteQuery`
- Getter：`GetTable` `GetArgs` `GetWhereCtrl` `GetJoinCtrl` `GetErr` 等（完整见 `go doc .../gdb.Sql`）

### G. `DbT[T]` 方法族（高频）

- 与 `Db` 基本同构，返回值强类型
- 查询：`One() (*T,error)` `Select() ([]*T,error)` `SelectAndCount()` `SelectPage()`
- 写入：`Create/Update/Save/Replace/...` 返回 `TOrmResult[T]`
- 其余链式方法（`Where/Join/Page/Order/...`）均返回 `*DbT[T]`

### H. `OrderedMap` 导出方法

- `Set` `Get` `Delete` `Keys` `Values`
- `All` `AllPairs`
- `MarshalJSON` `UnmarshalJSON`
- `SetValues` `SetKey`

## 4) 常见误区

- `NewSql` 只生成 SQL，不执行数据库操作
- `WriteCompSql`/`WriteErrSql` 可能带来敏感信息打印风险，生产环境谨慎开启
- `Db` 与 `DbT[T]` 可混用，但同一段业务建议统一风格

## 5) 测试命令

```bash
go test ./utils/gdb -v
go test ./utils/gdb -run '^TestDb_Select$' -v
go test ./utils/gdb -run '^TestGetSearchValue$' -v
go test ./utils/gdb/performance_test -bench '^BenchmarkSql_Select$' -benchmem
```

数据库相关测试依赖：`utils/test_conf.json`。

## 6) 精确导出清单（补充）

为避免遗漏，这里补充可机读的精确入口命令。大模型在分析本包时，优先先执行这些命令获取当前版本完整导出面。

```bash
go doc nwgit.gzhhit.com/BD/hhitcommcode.git/utils/gdb
go doc nwgit.gzhhit.com/BD/hhitcommcode.git/utils/gdb.Db
go doc nwgit.gzhhit.com/BD/hhitcommcode.git/utils/gdb.Sql
go doc nwgit.gzhhit.com/BD/hhitcommcode.git/utils/gdb.DbT
go doc nwgit.gzhhit.com/BD/hhitcommcode.git/utils/gdb.OrderedMap
```

当前版本已确认高频导出面：

- 顶层函数族：初始化、Option、轻量 CRUD、事务 CRUD、查询执行、日志设置、错误工具、工具函数
- `Db`：完整链式查询/写入/事务/日志控制/高级条件与搜索
- `Sql`：完整 SQL 构建链路与 getter 访问
- `DbT[T]`：与 `Db` 同构的强类型 API
- `OrderedMap`：有序键值维护与 JSON 编解码
