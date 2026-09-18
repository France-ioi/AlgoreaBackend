#!/usr/bin/env python3
"""Unit tests for tools/govulncheck_filter.py.

Run: python3 -m unittest tools/govulncheck_filter_test.py
Or:  make govulncheck-filter-test
"""

import io
import os
import tempfile
import unittest
from contextlib import redirect_stderr, redirect_stdout
from unittest import mock

from tools import govulncheck_filter as filt


def _config_msg():
	return '{"config": {"protocol_version": "v1.0.0", "scanner_name": "govulncheck"}}\n'


def _finding(osv_id):
	return f'{{"finding": {{"osv": "{osv_id}", "trace": []}}}}\n'


def _osv_catalog(osv_id):
	# Top-level osv is a DB catalog entry, not a detection.
	return f'{{"osv": {{"id": "{osv_id}", "summary": "catalog only"}}}}\n'


class FilterGovulncheckTest(unittest.TestCase):
	def _run(self, data, ignore=None):
		ignore = set() if ignore is None else ignore
		out, err = io.StringIO(), io.StringIO()
		with redirect_stdout(out), redirect_stderr(err):
			code = filt.filter_govulncheck(data, ignore)
		return code, out.getvalue(), err.getvalue()

	def test_empty_file_fails_closed(self):
		code, _, err = self._run("")
		self.assertEqual(code, 1)
		self.assertIn("missing config", err)

	def test_config_only_success(self):
		code, out, err = self._run(_config_msg())
		self.assertEqual(code, 0)
		self.assertIn("no vulnerabilities found", out)
		self.assertEqual(err, "")

	def test_ignore_only_go_2026_4316(self):
		data = _config_msg() + _finding("GO-2026-4316")
		code, out, err = self._run(data, {"GO-2026-4316"})
		self.assertEqual(code, 0)
		self.assertIn("ignored advisory(ies): GO-2026-4316", out)
		self.assertIn("only ignored advisories remain", out)
		self.assertEqual(err, "")

	def test_unexpected_id(self):
		data = _config_msg() + _finding("GO-9999-0001")
		code, _, err = self._run(data, {"GO-2026-4316"})
		self.assertEqual(code, 1)
		self.assertIn("GO-9999-0001", err)

	def test_mixed_ignore_and_unexpected(self):
		data = _config_msg() + _finding("GO-2026-4316") + _finding("GO-9999-0001")
		code, out, err = self._run(data, {"GO-2026-4316"})
		self.assertEqual(code, 1)
		self.assertIn("ignored advisory(ies): GO-2026-4316", out)
		self.assertIn("GO-9999-0001", err)
		self.assertNotIn("GO-2026-4316", err.split("unexpected")[-1] if "unexpected" in err else err)

	def test_comma_separated_ignore_list(self):
		# Exercises main()'s env parsing via filter with a multi-id set.
		data = _config_msg() + _finding("GO-AAAA-1") + _finding("GO-BBBB-2")
		code, out, _ = self._run(data, {"GO-AAAA-1", "GO-BBBB-2"})
		self.assertEqual(code, 0)
		self.assertIn("GO-AAAA-1", out)
		self.assertIn("GO-BBBB-2", out)

	def test_malformed_json_nonzero(self):
		code, _, err = self._run(_config_msg() + "{not-json")
		self.assertEqual(code, 1)
		self.assertIn("malformed JSON", err)

	def test_top_level_osv_catalog_not_a_finding(self):
		data = _config_msg() + _osv_catalog("GO-2026-4316")
		code, out, err = self._run(data)
		self.assertEqual(code, 0)
		self.assertIn("no vulnerabilities found", out)
		self.assertEqual(err, "")

	def test_garbage_without_config_fails(self):
		code, _, err = self._run('{"progress": {"message": "scanning"}}\n')
		self.assertEqual(code, 1)
		self.assertIn("missing config", err)


class MainCliTest(unittest.TestCase):
	def test_usage_argc(self):
		out, err = io.StringIO(), io.StringIO()
		with mock.patch.object(filt.sys, "argv", ["govulncheck_filter.py"]), redirect_stdout(out), redirect_stderr(err):
			code = filt.main()
		self.assertEqual(code, 2)
		self.assertIn("usage:", err.getvalue())

	def test_main_comma_separated_ignore_env(self):
		data = _config_msg() + _finding("GO-AAAA-1") + _finding("GO-BBBB-2")
		with tempfile.NamedTemporaryFile("w", encoding="utf-8", suffix=".json", delete=False) as fh:
			fh.write(data)
			path = fh.name
		try:
			err = io.StringIO()
			out = io.StringIO()
			with mock.patch.object(filt.sys, "argv", ["govulncheck_filter.py", path]), mock.patch.dict(
				os.environ, {"GOVULNCHECK_IGNORE": " GO-AAAA-1 , GO-BBBB-2 "}
			), redirect_stdout(out), redirect_stderr(err):
				code = filt.main()
			self.assertEqual(code, 0)
			self.assertIn("only ignored advisories remain", out.getvalue())
		finally:
			os.unlink(path)


if __name__ == "__main__":
	unittest.main()
