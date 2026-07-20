# log-analyzer v1.0.0 测试报告

## 1. 测试概况

- 测试日期：2026-07-20
- 发布分支：`release/1.0.0`
- 验收基线：`5b877d2`（UC-003 合入 `develop`）
- 测试范围：UC-001、UC-002、UC-003、异常处理、构建与静态检查
- 结论：通过，满足 v1.0.0 发布条件

## 2. 测试环境

### 本地环境

- 操作系统：Windows amd64
- Go：1.26.5
- Python：3.14.6
- matplotlib：3.11.1

### CI 环境

- GitHub Actions：`ubuntu-latest`
- Go：读取 `go.mod` 中的版本
- Python：3.x
- CI 工作流：格式检查、Go 测试及覆盖率门禁、Python 测试、`go vet`、Go 构建

## 3. 自动化测试结果

| 检查项 | 命令 | 结果 |
|---|---|---|
| Go 格式检查 | `gofmt -l .` | 通过，无未格式化文件 |
| Go 单元与命令测试 | `go test ./... -count=1` | 通过，7 个包全部成功 |
| Go 覆盖率 | `go test ./... -covermode=atomic -coverprofile=coverage.out` | 通过，总语句覆盖率 90.7% |
| Python 渲染器测试 | `python -m unittest internal/visualization/visualize_report_test.py` | 通过，8 项测试全部成功 |
| Go 静态检查 | `go vet ./...` | 通过 |
| Go 构建 | `go build ./cmd/log-tool` | 通过 |

### Go 包覆盖率

| 包 | 语句覆盖率 |
|---|---:|
| `cmd/log-tool` | 88.3% |
| `internal/filter` | 100.0% |
| `internal/logfile` | 100.0% |
| `internal/parser` | 100.0% |
| `internal/report` | 88.2% |
| `internal/stats` | 100.0% |
| `internal/visualization` | 91.6% |
| **总计** | **90.7%** |

所有 Go 包均达到不低于 85% 的目标，CI 同时对项目总语句覆盖率执行 85% 门禁。

## 4. 端到端验证

| 用例 | 验证内容 | 结果 |
|---|---|---|
| UC-001 | 读取日志、按级别过滤、跳过异常行、保存临时告警文件 | 通过 |
| UC-002 | 对 `2026-03-01` 至 `2026-03-02` 的示例日志生成 CSV | 通过：ERROR=2、INFO=2、WARN=1 |
| UC-003 | 将 CSV 渲染为自包含 HTML | 通过：HTML 53,635 字节，包含内嵌 PNG，无远程 HTTP 依赖 |

生成的 CSV 内容：

```csv
level,count,percentage
ERROR,2,40.00
INFO,2,40.00
WARN,1,20.00
```

## 5. 异常与边界场景

- 日志文件不存在时返回“文件路径无效，请检查路径后重试”。
- 格式错误日志不会中断后续处理，并写入单次运行共享的临时告警文件。
- 日期格式错误、开始日期晚于结束日期时返回明确提示。
- 无统计数据时生成仅含 CSV 表头的空报表。
- CSV 损坏、Python 或 `matplotlib` 缺失时返回规定的友好提示。
- 浏览器无法自动打开时保留 HTML，并提示用户手动打开。

## 6. 已知限制

- UC-003 运行环境需要 Python 3 和 `matplotlib`。
- 无桌面环境、CI、SSH 或部分 WSL 环境可能无法自动打开浏览器，但不影响 HTML 生成。
- 本次发布未设置注释覆盖率的自动统计门禁；代码注释质量通过人工检查维护。

## 7. 发布结论

三个核心用例均已实现并通过自动化及端到端验证，Go 总语句覆盖率高于 85%，CI、静态检查和构建均通过。当前版本可以作为项目一的首个稳定版本 `v1.0.0` 发布。
