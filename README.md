# log-analyzer

单机版日志采集与分析工具，使用 Go 编写。

## 当前进度

- 已完成 UC-001：流式读取日志并按级别过滤。
- 已完成 UC-002：按日期范围统计日志级别并生成 CSV 报告。

## 功能

- 逐行流式读取本地日志文件。
- 按日志级别过滤，级别匹配不区分大小写。
- 遇到格式错误的日志行时输出诊断并继续处理。
- 统计指定日期范围内各日志级别的数量和占比。
- 将统计结果按日志级别升序写入 CSV 文件。

## 日志格式

每行日志应包含日期、时间、级别和消息：

```text
2026-03-01 10:02:00 ERROR database connection failed
```

## 使用方法

### 按级别过滤日志

```powershell
go run ./cmd/log-tool --file ./testdata/sample.log --level ERROR
```

省略 `--level` 时输出文件中的全部有效日志。

### 生成统计报告

```powershell
go run ./cmd/log-tool `
  --file ./testdata/sample.log `
  --stat `
  --start 2026-03-01 `
  --end 2026-03-02 `
  --output report.csv
```

统计模式说明：

- `--start` 和 `--end` 都包含在统计范围内。
- `--output` 默认值为 `report.csv`。
- `--stat` 不能与 `--level` 同时使用。
- 日志级别按大小写不敏感的方式合并，并统一以大写形式输出。
- `percentage` 不带百分号并保留两位小数。
- 没有匹配日志时，CSV 仍保留表头。

CSV 示例：

```csv
level,count,percentage
ERROR,2,40.00
INFO,2,40.00
WARN,1,20.00
```

详细需求和验收标准见 [UC-002 文档](docs/UC-002.md)。

## 本地验证

```powershell
gofmt -l .
go test ./...
go vet ./...
go build ./cmd/log-tool
```
