# gdb 测试覆盖清单

## 说明

- 本清单用于快速查看 `utils/gdb` 的单元测试/集成测试覆盖范围。
- “已覆盖”表示有明确断言的测试路径，不代表每个分支 100% 覆盖。

## 已覆盖文件与能力

### 1) `db_config_test.go`（集成测试，依赖数据库）

- 连接初始化（读取 `config.json`）
- `Create / Select / One / Update / Count / Exists / Delete`
- 事务 `Transaction`
- `Raw + Map`、`Raw + Maps`

### 2) `utils_unit_test.go`

- `MapToMapAny`
- `MapsToMapsAny`
- `StructDbField`
- `strToArr`
- `replaceIndex` / `strIndex`
- `genPrePil` / `genPrePilGroup`
- `genQueryList` / `genQueryGroupAnyList`
- `arrToArrList`
- `InArr` / `As` / `EscapeLike`
- `GetTableName`
- `ValidateOrderParam`

### 3) `types_unit_test.go`

- `ReadOnlyMap`：`Get/Keys/Len/Range`
- `OrderedMap`：`Set/Get/Delete/Keys/Values`
- `OrderedMap` JSON：`MarshalJSON/UnmarshalJSON`
- `JSONType`：`Value/Scan/MarshalJSON`
- `ENUMType`：`Value/Scan`
- `NewPage`
- `QueryWrapper`：`GetSql/GetArgs/GetWhereSql/GetWhereArgs`
- `StreamCallback`

### 4) `search_param_unit_test.go`

- `GetSearchParams/GetSearchItem`
- `GetSearchValue/GetSearchIntValue/GetSearchStringValue/GetSearchOp`
- `ReplaceFieldName/ReplaceFieldValue`
- `RemoveSearch/KeepSearch`
- `ReplaceSort/KeepSort`
- `RequiredSearch/AddSearch`

### 5) `cte_query_unit_test.go`

- `NewCteQuery`
- `SetCet`
- `SetCteTable`（`*Sql`、`DbFace`、非法类型）

### 6) `sql_build_type_unit_test.go`

- `Raw`
- `RawAny`
- `Result.CompSql`

### 7) `db_generics_unit_test.go`

- `Model[T]` 表名推断
- `Scan/ScanAndCount`（不支持分支）
- `WhereSearchParams` 分页注入
- `PageReset`
- `setLastInsertId`

### 8) `db_helper_unit_test.go`

- `Exec` 错误路径
- `QueryOne` 错误路径

### 9) `db_quick_func_unit_test.go`

- `Db` 快捷条件：
    - `Eq/NotEq/Like/LikeRaw/In/Lt/Le/Gt/Ge/Between/IsNotNull`
- `DbT` 快捷条件：
    - `Eq/NotEq/Like/LikeRaw/In/Lt/Le/Gt/Ge/Between/IsNotNull`

### 10) `sql_chain_unit_test.go`

- `Sql` 链式构建：
    - `Where/WhereGroup/Order/Page/Select`
    - `Update/Delete/Exists`
    - `Join/Union`

### 11) `init_open_log_unit_test.go`

- `init_func.go` 全部 Option（`With*`）
- `Init` 对全局配置生效
- `open.go`：`Db.GenWhere/QueryWrapper`、`Sql.GenWhere/QueryWrapper`
- `log.go`：日志深度 ctx、`gdbLog` 输出、`Clone`、`writer.Write`

### 12) `mysql_unit_test.go`

- `getSearchWhere`：`equal/in/between/like/unknown`
- `genQuery`：`[]int`、`[][]any`、数组、`RawArg`、异常输入
- `ppNum`
- `GenWhere`

## 已修复的问题（由测试驱动发现）

- `utils.go`：`MapToMapAny` 处理 `map[string]RawBody`/`map[string]string` 时 map 未初始化，已修复。
- `db_quick_func.go`：`Like` 生成 SQL 错误（重复 `LIKE ?`），已修复。

## 后续可扩展测试点

- `mysql.go` 其余边界分支（更多组合参数）
- `db.go` 的高级链路分支（复杂 `WhereSearchView`、`Stream` 错误场景）
- `db_tx.go` 的失败回滚场景（需集成环境）
