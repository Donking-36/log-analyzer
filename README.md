# log-analyzer

单机版日志采集与分析工具，使用 Go 编写。

## 当前进度

- 已完成 UC-001：流式读取日志并按级别过滤。
- 已完成 UC-002：按日期范围统计日志级别并生成 CSV 报告。
- 已完成 UC-003：将 CSV 统计报告转换为可离线打开的 HTML 柱状图。

## 功能

- 逐行流式读取本地日志文件。
- 按日志级别过滤，级别匹配不区分大小写。
- 遇到格式错误的日志行时输出诊断，将告警写入系统临时文件并继续处理。
- 统计指定日期范围内各日志级别的数量和占比。
- 将统计结果按日志级别升序写入 CSV 文件。
- 使用 Python 3 和 `matplotlib` 生成日志级别数量柱状图。
- 生成自包含 HTML，支持放大、缩小、重置和下载 PNG。

## 可视化依赖

UC-001 和 UC-002 只需要 Go。使用 UC-003 前，需要安装 Python 3 和 `matplotlib`；工具不会自动安装依赖。

Windows：

```powershell
py -3 --version
py -3 -m pip install matplotlib
```

如果系统没有 `py` 启动器，可以改用：

```powershell
python --version
python -m pip install matplotlib
```

macOS 或 Linux：

```bash
python3 --version
python3 -m pip install matplotlib
```

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

### 生成可视化报告

```powershell
go run ./cmd/log-tool --visual --csv ./report.csv
```

可视化模式说明：

- 不需要同时提供 `--file`。
- `report.csv` 会在同一目录生成 `report.html`。
- 成功信息会显示 HTML 的绝对路径，并尝试使用默认浏览器打开。
- 仅包含 `level,count,percentage` 表头的空报表是合法输入，会生成“暂无统计数据”页面。
- HTML 不依赖网络资源，可以离线打开。
- 页面支持放大、缩小、重置和下载图表 PNG。
- `--visual` 不能与 `--file`、`--level`、`--stat`、`--start`、`--end` 或 `--output` 混用。

Windows 使用系统文件关联打开 HTML，macOS 使用 `open`，Linux 使用 `xdg-open`。在 CI、SSH、无桌面 Linux 或部分 WSL 环境中，自动打开可能失败，但已生成的 HTML 不会被删除。

## 故障排查

- 日志行格式错误：工具会跳过该行，并在标准错误中显示临时告警文件的绝对路径。
- CSV 不存在、无法读取或格式损坏：`报表文件无效或损坏，请重新生成`
- Python 3 或 `matplotlib` 不可用：`请安装Python及matplotlib库后重试`
- 浏览器无法自动打开：根据错误信息中的绝对路径手动打开 HTML。
- HTML 无法写入：检查错误信息中的目标路径和底层系统原因，例如目录权限或磁盘空间。

## 需求文档

- [UC-001：日志读取与级别过滤](docs/UC-001.md)
- [UC-002：日志统计与 CSV 报告](docs/UC-002.md)
- [UC-003：统计结果可视化](docs/UC-003.md)

## 本地验证

```powershell
gofmt -l .
go test ./... -count=1 -coverprofile=coverage.out
go tool cover -func=coverage.out
py -3 -m unittest internal/visualization/visualize_report_test.py
go vet ./...
go build ./cmd/log-tool
```

CI 会检查 Go 测试总语句覆盖率，低于 85% 时检查失败。`coverage.out` 已被 Git 忽略。

macOS 或 Linux 请将最后一组命令中的 `py -3` 改为 `python3`。
