#!/usr/bin/env python3
"""Fail if govulncheck JSON output contains findings outside GOVULNCHECK_IGNORE."""

import json
import os
import sys


def filter_govulncheck(data: str, ignore: set) -> int:
	"""Parse govulncheck -json output and exit-status semantics for findings.

	Returns 0 on success, 1 on unexpected findings or invalid/incomplete stream.
	"""
	dec = json.JSONDecoder()
	i, n = 0, len(data)
	osvs = set()
	saw_config = False
	try:
		while i < n:
			while i < n and data[i].isspace():
				i += 1
			if i >= n:
				break
			obj, end = dec.raw_decode(data, i)
			i = end
			if not isinstance(obj, dict):
				print("govulncheck: invalid JSON stream (expected objects)", file=sys.stderr)
				return 1
			if "config" in obj:
				saw_config = True
			# Top-level "osv" messages are vuln-DB catalog entries for required
			# modules and must not be treated as detections (golang/go#65132).
			finding = obj.get("finding")
			if finding and finding.get("osv"):
				osvs.add(finding["osv"])
	except json.JSONDecodeError as err:
		print(f"govulncheck: malformed JSON: {err}", file=sys.stderr)
		return 1

	if not saw_config:
		print(
			"govulncheck: missing config message (not real govulncheck -json output)",
			file=sys.stderr,
		)
		return 1

	unexpected = sorted(osvs - ignore)
	ignored = sorted(osvs & ignore)
	if ignored:
		print("govulncheck: ignored advisory(ies): " + ", ".join(ignored))
	if unexpected:
		print(
			"govulncheck: unexpected vulnerabilit(ies): " + ", ".join(unexpected),
			file=sys.stderr,
		)
		return 1
	if not osvs:
		print("govulncheck: no vulnerabilities found")
	else:
		print("govulncheck: ok (only ignored advisories remain)")
	return 0


def main() -> int:
	if len(sys.argv) != 2:
		print("usage: govulncheck_filter.py <govulncheck-json>", file=sys.stderr)
		return 2

	ignore = {x.strip() for x in os.environ.get("GOVULNCHECK_IGNORE", "").split(",") if x.strip()}
	data = open(sys.argv[1], encoding="utf-8").read()
	return filter_govulncheck(data, ignore)


if __name__ == "__main__":
	sys.exit(main())
