# Changelog

本文件记录 `log-analyzer` 的重要版本变更。版本号遵循语义化版本约定。

## [1.0.0] - 2026-07-20

首个稳定版本，完成单机日志“读取、过滤、统计、可视化”闭环。

### Added

- UC-003：校验 UC-002 CSV 报表并生成自包含 HTML 柱状图。
- 使用 Python 3 和 `matplotlib` 渲染图表，通过 Go 子进程调用。
- HTML 支持放大、缩小、重置和下载 PNG，并可离线打开。
- 生成成功后尝试使用操作系统默认浏览器打开报告。
- 将格式错误日志的诊断保存到按次创建的系统临时告警文件。
- 在 CI 中生成 Go 覆盖率报告，并强制执行 85% 总语句覆盖率门槛。

### Fixed

- 日志文件不存在时返回稳定、友好的中文提示。
- 浏览器无法打开时保留已经生成的 HTML，并提示用户手动打开。
- 保留 Python 可视化失败的底层诊断，便于定位环境和写入问题。

## [0.2.0] - 2026-07-19

### Added

- UC-002：按包含首尾日期的时间范围统计日志级别。
- 生成包含 `level`、`count`、`percentage` 字段的 CSV 报表。
- 无匹配数据时生成仅含表头的空报表。

## [0.1.0] - 2026-07-18

### Added

- UC-001：流式读取本地日志文件。
- 按日志级别进行大小写不敏感的过滤。
- 跳过格式错误的日志行并继续处理后续记录。

[1.0.0]: https://github.com/Donking-36/log-analyzer/compare/v0.2.0...v1.0.0
[0.2.0]: https://github.com/Donking-36/log-analyzer/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Donking-36/log-analyzer/releases/tag/v0.1.0
