#!/usr/bin/env bash
# Emit a Redoc HTML page for a given release tag (e.g. v2.59.0).
# Used by CircleCI swagger-deploy and tools/redoc-spec-backfill.sh so they stay in sync.
# Hosted pages pin Redoc v2 (cdn.redoc.ly .../redoc/v2.x/...) rather than latest.
set -euo pipefail

tag="${1:-}"
if [ -z "$tag" ]; then
  echo "usage: $0 <tag>" >&2
  exit 1
fi

cat <<EOF
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Algorea API ${tag}</title>
</head>
<body>
<redoc spec-url="https://franceioi-algorea.s3.eu-west-3.amazonaws.com/spec/swagger-${tag}.yaml"></redoc>
<script src="https://cdn.redoc.ly/redoc/v2.x/bundles/redoc.standalone.js"></script>
</body>
</html>
EOF
