"""Standard-library tests for the embedded UC-003 Python renderer."""

from contextlib import redirect_stderr
import importlib.util
import io
import os
from pathlib import Path
import tempfile
import unittest
from unittest import mock


SCRIPT_PATH = Path(__file__).with_name("visualize_report.py")
SPEC = importlib.util.spec_from_file_location("visualize_report_under_test", SCRIPT_PATH)
SUBJECT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(SUBJECT)


class ReportReadingTests(unittest.TestCase):
    def write_report(self, directory, content):
        report_path = Path(directory, "report.csv")
        report_path.write_text(content, encoding="utf-8", newline="")
        return report_path

    def test_reads_valid_report(self):
        with tempfile.TemporaryDirectory() as directory:
            report_path = self.write_report(
                directory,
                "level,count,percentage\nERROR,2,100.00\n",
            )

            self.assertEqual(
                SUBJECT.read_report(report_path),
                [("ERROR", 2, 100.0)],
            )

    def test_accepts_header_only_report(self):
        with tempfile.TemporaryDirectory() as directory:
            report_path = self.write_report(
                directory,
                "level,count,percentage\n",
            )

            self.assertEqual(SUBJECT.read_report(report_path), [])

    def test_rejects_invalid_report(self):
        invalid_reports = (
            "wrong,count,percentage\n",
            "level,count,percentage\nERROR,-1,100.00\n",
            "level,count,percentage\nERROR,1,nan\n",
        )

        for content in invalid_reports:
            with self.subTest(content=content):
                with tempfile.TemporaryDirectory() as directory:
                    report_path = self.write_report(directory, content)
                    with self.assertRaises(SUBJECT.InvalidReportError):
                        SUBJECT.read_report(report_path)


class HTMLGenerationTests(unittest.TestCase):
    def test_builds_self_contained_interactive_html(self):
        html = SUBJECT.build_html(b"\x89PNG\r\n", empty=False)

        self.assertIn("data:image/png;base64,", html)
        self.assertIn("Content-Security-Policy", html)
        self.assertIn("放大", html)
        self.assertIn("缩小", html)
        self.assertIn("重置", html)
        self.assertIn("下载 PNG", html)
        self.assertNotIn("https://", html)
        self.assertNotIn("http://", html)

    def test_builds_empty_report_message(self):
        html = SUBJECT.build_html(b"png", empty=True)

        self.assertIn("暂无统计数据", html)

    def test_writes_html_atomically(self):
        with tempfile.TemporaryDirectory() as directory:
            output_path = Path(directory, "report.html")

            SUBJECT.write_html_atomically(output_path, "<html>ok</html>")

            self.assertEqual(
                output_path.read_text(encoding="utf-8"),
                "<html>ok</html>",
            )
            self.assertEqual(list(Path(directory).glob("*.tmp")), [])


class MainExitCodeTests(unittest.TestCase):
    def test_reports_underlying_write_error(self):
        stderr = io.StringIO()
        with (
            mock.patch.dict(os.environ, {}, clear=False),
            mock.patch.object(SUBJECT, "paths_refer_to_same_file", return_value=False),
            mock.patch.object(SUBJECT, "read_report", return_value=[("ERROR", 1, 100.0)]),
            mock.patch.object(SUBJECT, "load_pyplot", return_value=object()),
            mock.patch.object(SUBJECT, "render_png", return_value=b"png"),
            mock.patch.object(
                SUBJECT,
                "write_html_atomically",
                side_effect=OSError("permission denied"),
            ),
            redirect_stderr(stderr),
        ):
            exit_code = SUBJECT.main(["--csv", "report.csv", "--output", "report.html"])

        self.assertEqual(exit_code, SUBJECT.EXIT_WRITE_FAILED)
        self.assertIn(
            "VISUALIZE_WRITE_FAILED: permission denied",
            stderr.getvalue(),
        )

    def test_maps_missing_matplotlib(self):
        stderr = io.StringIO()
        with (
            mock.patch.dict(os.environ, {}, clear=False),
            mock.patch.object(SUBJECT, "paths_refer_to_same_file", return_value=False),
            mock.patch.object(SUBJECT, "read_report", return_value=[]),
            mock.patch.object(SUBJECT, "load_pyplot", side_effect=ImportError),
            redirect_stderr(stderr),
        ):
            exit_code = SUBJECT.main(["--csv", "report.csv", "--output", "report.html"])

        self.assertEqual(exit_code, SUBJECT.EXIT_MISSING_MATPLOTLIB)
        self.assertIn("VISUALIZE_MATPLOTLIB_MISSING", stderr.getvalue())


if __name__ == "__main__":
    unittest.main()
