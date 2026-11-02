#!/usr/bin/env bash
# Build the hosted API version index (stdout), or list the selected tags.
# Reads release tags or *.html names from stdin (one per line); keeps the latest
# 10 by version sort (GNU sort -V). Shared by CI and the backfill script.
# Usage:
#   tools/redoc-spec-index.sh < tags.txt > index.html
#   tools/redoc-spec-index.sh --list-tags < tags.txt
set -euo pipefail

limit=10
# Aligns with CircleCI release-tag filter /^v\d+\.\d+\..*/
tag_re='^v[0-9]+\.[0-9]+\..*'
list_tags=0

for arg in "$@"; do
  case "$arg" in
    --list-tags) list_tags=1 ;;
    -h|--help)
      echo "usage: $0 [--list-tags]" >&2
      echo "reads tags or *.html names from stdin" >&2
      exit 0
      ;;
    *)
      echo "unknown argument: $arg" >&2
      echo "usage: $0 [--list-tags]" >&2
      exit 1
      ;;
  esac
done

tags="$(
  while IFS= read -r line || [ -n "$line" ]; do
    [ -n "$line" ] || continue
    tag="${line%.html}"
    printf '%s\n' "$tag" | grep -Eq "$tag_re" || continue
    printf '%s\n' "$tag"
  done | sort -urV | head -n "$limit"
)"

if [ "$list_tags" -eq 1 ]; then
  printf '%s\n' "$tags"
  exit 0
fi

cat <<EOF
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Algorea API versions</title>
<style>
  body {
    margin: 0;
    padding: 1.25rem 1.5rem;
    font-family: "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    font-size: 16px;
    line-height: 1.5;
    color: #1a1a1a;
    background: #f7f7f5;
  }
  h1 {
    margin: 0 0 0.25rem;
    font-size: 1.25rem;
    font-weight: 600;
  }
  .note {
    margin: 0 0 1rem;
    color: #5c5c5c;
    font-size: 0.875rem;
  }
  ul {
    margin: 0;
    padding: 0;
    list-style: none;
    border: 1px solid #e2e2de;
    border-radius: 6px;
    background: #fff;
    overflow: hidden;
  }
  li + li {
    border-top: 1px solid #e2e2de;
  }
  a {
    display: block;
    padding: 0.65rem 0.9rem;
    color: #0b57d0;
    text-decoration: none;
  }
  a:hover,
  a:focus-visible {
    background: #eef3fc;
    outline: none;
  }
</style>
</head>
<body>
<h1>Algorea API</h1>
<p class="note">Latest ${limit} releases</p>
<ul>
EOF

if [ -z "$tags" ]; then
  printf '%s\n' '  <li><span style="display:block;padding:0.65rem 0.9rem;color:#5c5c5c;">No versions found</span></li>'
else
  while IFS= read -r tag; do
    [ -n "$tag" ] || continue
    printf '  <li><a href="./%s.html">%s</a></li>\n' "$tag" "$tag"
  done <<EOF
$tags
EOF
fi

cat <<EOF
</ul>
</body>
</html>
EOF
