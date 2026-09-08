#!/usr/bin/env bash
set -euo pipefail

profile="${1:?usage: check_coverage.sh <coverprofile>}"
threshold="${COVERAGE_THRESHOLD:-80}"

if [ ! -f "$profile" ]; then
  echo "coverage profile not found: $profile" >&2
  exit 1
fi

total_line="$(go tool cover -func="$profile" | awk '/^total:/ {print $0}')"
if [ -z "$total_line" ]; then
  echo "failed to parse coverage from $profile" >&2
  exit 1
fi

pct="$(echo "$total_line" | awk '{gsub(/%/,"",$3); print $3}')"
echo "$total_line"

awk -v pct="$pct" -v threshold="$threshold" 'BEGIN {
  if (pct + 0 < threshold + 0) {
    printf "coverage %.1f%% is below required %s%%\n", pct, threshold > "/dev/stderr"
    exit 1
  }
  printf "coverage %.1f%% meets required %s%%\n", pct, threshold
}'
