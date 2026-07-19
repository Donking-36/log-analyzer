#!/usr/bin/env python3
"""Render a UC-002 CSV report as a self-contained HTML visualization."""

import argparse
import base64
import csv
import io
import math
import os
from pathlib import Path
import sys
import tempfile

EXIT_INVALID_REPORT = 3
EXIT_MISSING_MATPLOTLIB = 4
EXIT_WRITE_FAILED = 5
EXIT_RENDER_FAILED = 6


class InvalidReportError(Exception):
    """Raised when the CSV report violates the UC-002 schema."""


def parse_args(argv):
    parser = argparse.ArgumentParser(description="Render a log statistics CSV report")
    parser.add_argument("--csv", required=True, help="input CSV report path")
    parser.add_argument("--output", required=True, help="output HTML report path")
    return parser.parse_args(argv)


def read_report(file_path):
    rows = []
    try:
        with open(file_path, "r", encoding="utf-8-sig", newline="") as handle:
            reader = csv.reader(handle, strict=True)
            try:
                header = next(reader)
            except StopIteration as exc:
                raise InvalidReportError from exc

            if header != ["level", "count", "percentage"]:
                raise InvalidReportError

            for record in reader:
                if not record:
                    continue
                if len(record) != 3 or not record[0].strip():
                    raise InvalidReportError

                try:
                    count = int(record[1].strip())
                    percentage = float(record[2].strip())
                except ValueError as exc:
                    raise InvalidReportError from exc

                if count < 0 or not math.isfinite(percentage) or not 0 <= percentage <= 100:
                    raise InvalidReportError

                rows.append((record[0].strip(), count, percentage))
    except (OSError, UnicodeError, csv.Error) as exc:
        raise InvalidReportError from exc

    return rows


def load_pyplot():
    try:
        import matplotlib

        matplotlib.use("Agg", force=True)
        from matplotlib import pyplot
    except Exception as exc:
        raise ImportError from exc
    return pyplot


def render_png(rows, pyplot):
    width = max(8.0, min(18.0, 4.0 + len(rows) * 1.1))
    figure, axis = pyplot.subplots(figsize=(width, 6.0))
    try:
        if rows:
            levels = [row[0] for row in rows]
            counts = [row[1] for row in rows]
            positions = list(range(len(rows)))
            bars = axis.bar(positions, counts, color="#2563eb", edgecolor="#1e3a8a")
            axis.set_title("Log Level Statistics")
            axis.set_xlabel("Log Level")
            axis.set_ylabel("Count")
            axis.set_xticks(positions)
            axis.set_xticklabels(
                levels,
                rotation=30 if len(rows) > 5 else 0,
                ha="right" if len(rows) > 5 else "center",
            )
            axis.set_ylim(bottom=0)
            axis.grid(axis="y", linestyle="--", alpha=0.35)
            for bar, count in zip(bars, counts):
                axis.annotate(
                    str(count),
                    xy=(bar.get_x() + bar.get_width() / 2, bar.get_height()),
                    xytext=(0, 3),
                    textcoords="offset points",
                    ha="center",
                    va="bottom",
                )
        else:
            axis.axis("off")
            axis.text(
                0.5,
                0.5,
                "No statistics available",
                ha="center",
                va="center",
                fontsize=18,
                color="#475569",
                transform=axis.transAxes,
            )

        figure.tight_layout()
        buffer = io.BytesIO()
        figure.savefig(buffer, format="png", dpi=160, bbox_inches="tight")
        return buffer.getvalue()
    finally:
        pyplot.close(figure)


def build_html(png_bytes, empty):
    image_data = base64.b64encode(png_bytes).decode("ascii")
    status = "暂无统计数据" if empty else "柱状图展示各日志级别的数量。"
    template = """<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src data:; style-src 'unsafe-inline'; script-src 'unsafe-inline'">
  <title>日志级别统计</title>
  <style>
    :root { color-scheme: light; font-family: system-ui, -apple-system, "Segoe UI", sans-serif; }
    body { margin: 0; background: #f8fafc; color: #0f172a; }
    main { width: min(1120px, calc(100% - 32px)); margin: 32px auto; }
    h1 { margin: 0 0 8px; font-size: 28px; }
    .status { margin: 0 0 20px; color: #475569; }
    .toolbar { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
    button, .download { border: 1px solid #cbd5e1; border-radius: 8px; background: white; color: #0f172a; padding: 8px 14px; font: inherit; cursor: pointer; text-decoration: none; }
    button:hover, .download:hover { background: #eff6ff; border-color: #60a5fa; }
    .viewport { overflow: auto; border: 1px solid #dbe3ee; border-radius: 12px; background: white; padding: 16px; box-shadow: 0 8px 24px rgb(15 23 42 / 8%); }
    #chart { display: block; width: 100%; height: auto; max-width: none; transform-origin: top left; }
  </style>
</head>
<body>
  <main>
    <h1>日志级别统计</h1>
    <p class="status">__STATUS__</p>
    <div class="toolbar" aria-label="图表操作">
      <button type="button" onclick="zoomBy(0.25)">放大</button>
      <button type="button" onclick="zoomBy(-0.25)">缩小</button>
      <button type="button" onclick="resetZoom()">重置</button>
      <a class="download" id="download" download="log-level-statistics.png">下载 PNG</a>
    </div>
    <div class="viewport">
      <img id="chart" alt="日志级别数量柱状图" src="data:image/png;base64,__IMAGE_DATA__">
    </div>
  </main>
  <script>
    const chart = document.getElementById("chart");
    const download = document.getElementById("download");
    let scale = 1;
    download.href = chart.src;
    function applyZoom() {
      chart.style.width = `${scale * 100}%`;
    }
    function zoomBy(delta) {
      scale = Math.min(4, Math.max(0.5, scale + delta));
      applyZoom();
    }
    function resetZoom() {
      scale = 1;
      applyZoom();
    }
  </script>
</body>
</html>
"""
    return template.replace("__STATUS__", status).replace("__IMAGE_DATA__", image_data)


def write_html_atomically(output_path, html_text):
    output = Path(output_path)
    temporary_path = None
    try:
        with tempfile.NamedTemporaryFile(
            mode="w",
            encoding="utf-8",
            newline="\n",
            dir=output.parent,
            prefix=f".{output.name}.",
            suffix=".tmp",
            delete=False,
        ) as handle:
            temporary_path = Path(handle.name)
            handle.write(html_text)
            handle.flush()
            os.fsync(handle.fileno())
        os.replace(temporary_path, output)
    except OSError:
        if temporary_path is not None:
            try:
                temporary_path.unlink(missing_ok=True)
            except OSError:
                pass
        raise


def paths_refer_to_same_file(first_path, second_path):
    first = Path(first_path)
    second = Path(second_path)
    try:
        return first.samefile(second)
    except OSError:
        return first.resolve() == second.resolve()


def main(argv=None):
    args = parse_args(argv)

    if paths_refer_to_same_file(args.csv, args.output):
        print("VISUALIZE_INVALID_REPORT", file=sys.stderr)
        return EXIT_INVALID_REPORT

    try:
        rows = read_report(args.csv)
    except InvalidReportError:
        print("VISUALIZE_INVALID_REPORT", file=sys.stderr)
        return EXIT_INVALID_REPORT

    try:
        with tempfile.TemporaryDirectory(prefix="log-tool-matplotlib-") as config_dir:
            os.environ["MPLCONFIGDIR"] = config_dir
            pyplot = load_pyplot()
            png_bytes = render_png(rows, pyplot)
    except ImportError:
        print("VISUALIZE_MATPLOTLIB_MISSING", file=sys.stderr)
        return EXIT_MISSING_MATPLOTLIB
    except Exception:
        print("VISUALIZE_RENDER_FAILED", file=sys.stderr)
        return EXIT_RENDER_FAILED

    try:
        html_text = build_html(png_bytes, empty=not rows)
        write_html_atomically(args.output, html_text)
    except OSError as exc:
        print(f"VISUALIZE_WRITE_FAILED: {exc}", file=sys.stderr)
        return EXIT_WRITE_FAILED
    except Exception:
        print("VISUALIZE_RENDER_FAILED", file=sys.stderr)
        return EXIT_RENDER_FAILED

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
