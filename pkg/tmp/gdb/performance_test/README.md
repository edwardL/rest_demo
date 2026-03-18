# GDB 性能测试套件

## 概述

本目录包含 GDB (Go Database) 包的性能测试代码，用于评估和对比 GDB 与其他 ORM 框架（如 GORM）的性能表现。

## 文件结构

- `db_performance_test.go` - 数据库性能对比测试代码
- `db_performance_report.md` - 性能测试报告
- `struct_map_conversion_performance_test.go` - 结构体与Map转换性能测试
- `struct_map_conversion_performance_report.md` - 转换性能测试报告
- `sql_performance_test.go` - SQL构建性能测试
- `sql_performance_report.md` - SQL性能测试报告

## 测试内容

### 1. 数据库操作性能测试

对比 GDB 与 GORM 在以下操作中的性能表现：

- 单条记录查询
- 批量记录查询
- 数据插入
- 数据更新
- 数据删除

### 2. 结构体与Map转换性能测试

评估 GDB 在结构体与 Map 之间转换的性能表现。

### 3. SQL构建性能测试

测试 GDB SQL构建器在复杂查询构建中的性能表现。

## 运行测试

### 前置条件

1. 确保数据库服务可用
2. 配置正确的数据库连接信息 (`test_conf.json`)

### 运行命令

```bash
# 运行所有性能测试
go test -bench=. -v ./gdb/performance_test/

# 运行特定的性能测试
go test -bench=BenchmarkGdbSingleQuery -v ./gdb/performance_test/

# 运行所有测试并生成性能分析报告
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./gdb/performance_test/
```

## 测试结果

详细的性能测试结果请参考 [db_performance_report.md](db_performance_report.md)。

## 配置说明

测试使用 `test_conf.json` 文件进行数据库连接配置，如果该文件不存在，将使用默认配置连接到本地 MySQL 服务。

## 注意事项

1. 性能测试结果可能因环境不同而有所差异
2. 运行性能测试前确保数据库连接正常
3. 建议在相同环境下进行对比测试以保证结果准确性
