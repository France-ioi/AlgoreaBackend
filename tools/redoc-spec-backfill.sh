#!/usr/bin/env bash
# Occasional backfill: regenerate only the latest 10 versioned Redoc HTML pages
# under s3://franceioi-algorea/spec/ (the ones linked from the index) via
# tools/redoc-spec-page.sh, then rebuild index.html via tools/redoc-spec-index.sh.
# Use after changing those templates without cutting a new release tag.
# Older version objects are left untouched.
# Requires AWS credentials with write access to that bucket (region eu-west-3).
# Usage: tools/redoc-spec-backfill.sh [--dry-run]
set -euo pipefail

bucket="s3://franceioi-algorea/spec"
region="eu-west-3"
dry_run=0

for arg in "$@"; do
  case "$arg" in
    --dry-run) dry_run=1 ;;
    -h|--help)
      echo "usage: $0 [--dry-run]" >&2
      exit 0
      ;;
    *)
      echo "unknown argument: $arg" >&2
      echo "usage: $0 [--dry-run]" >&2
      exit 1
      ;;
  esac
done

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
page_script="${script_dir}/redoc-spec-page.sh"
index_script="${script_dir}/redoc-spec-index.sh"

if [ ! -x "$page_script" ]; then
  echo "missing executable template script: $page_script" >&2
  exit 1
fi
if [ ! -x "$index_script" ]; then
  echo "missing executable index script: $index_script" >&2
  exit 1
fi

listing="$(aws s3 ls "${bucket}/" --region "$region")"
objects="$(printf '%s\n' "$listing" | awk '{print $4}' | grep '\.html$' || true)"
latest_tags="$(printf '%s\n' "$objects" | "$index_script" --list-tags)"

if [ -z "$latest_tags" ]; then
  echo "no release-tag HTML pages found under ${bucket}/"
  exit 0
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

while IFS= read -r tag; do
  [ -n "$tag" ] || continue
  object="${tag}.html"
  out="${tmpdir}/${object}"
  "$page_script" "$tag" > "$out"
  dest="${bucket}/${object}"
  if [ "$dry_run" -eq 1 ]; then
    echo "dry-run: would upload ${dest}"
  else
    echo "uploading ${dest}"
    aws s3 cp "$out" "$dest" \
      --region "$region" \
      --acl public-read \
      --content-type text/html
  fi
done <<EOF
$latest_tags
EOF

index_out="${tmpdir}/index.html"
printf '%s\n' "$latest_tags" | "$index_script" > "$index_out"
index_dest="${bucket}/index.html"
if [ "$dry_run" -eq 1 ]; then
  echo "dry-run: would upload ${index_dest} (latest 10 versions)"
else
  echo "uploading ${index_dest}"
  aws s3 cp "$index_out" "$index_dest" \
    --region "$region" \
    --acl public-read \
    --content-type text/html
fi
