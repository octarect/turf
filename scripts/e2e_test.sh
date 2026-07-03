#!/usr/bin/env bash
set -euo pipefail

TURF_BIN="${TURF_BIN:-turf}"
SLEEP_SEC="${SLEEP_SEC:-1}"

usage() {
  echo "Usage: $(basename "$0") CASES_FILE [--bin PATH] [--sleep SEC]"
  echo ""
  echo "Arguments:"
  echo "  CASES_FILE      Path to cases TSV file"
  echo ""
  echo "Options:"
  echo "  --bin PATH      Path to turf binary (default: turf)"
  echo "  --sleep SEC     Sleep between requests in seconds (default: 1)"
  exit 1
}

if [[ $# -lt 1 ]]; then usage; fi
CASES_FILE="$1"; shift

while [[ $# -gt 0 ]]; do
  case "$1" in
    --bin)   TURF_BIN="$2";  shift 2 ;;
    --sleep) SLEEP_SEC="$2"; shift 2 ;;
    -h|--help) usage ;;
    *) echo "Unknown option: $1"; usage ;;
  esac
done

if ! command -v "$TURF_BIN" &>/dev/null; then
  echo "Error: turf binary not found: $TURF_BIN" >&2
  exit 1
fi

if [[ ! -f "$CASES_FILE" ]]; then
  echo "Error: cases file not found: $CASES_FILE" >&2
  exit 1
fi

pass=0
fail=0
failed_cases=()

total=$(grep -v '^#' "$CASES_FILE" | grep -v '^[[:space:]]*$' | wc -l | tr -d ' ')
echo "Running E2E tests: $total cases"
echo "Binary: $TURF_BIN"
echo "Cases:  $CASES_FILE"
echo ""

index=0
while IFS=$'\t' read -r date course race_no surface distance; do
  [[ "$date" =~ ^# ]] && continue
  [[ -z "$date" ]] && continue

  index=$((index + 1))
  label="[${index}/${total}] ${date} ${course} race=${race_no} (${surface} ${distance}m)"

  output=$("$TURF_BIN" result --date "$date" --course "$course" --race "$race_no" -o json 2>&1) && exit_code=0 || exit_code=$?

  if [[ $exit_code -eq 0 ]]; then
    echo "[PASS] ${label}"
    pass=$((pass + 1))
  else
    echo "[FAIL] ${label}"
    echo "       ${output}"
    failed_cases+=("${date} ${course} race=${race_no}")
    fail=$((fail + 1))
  fi

  if [[ $index -lt $total ]]; then
    sleep "$SLEEP_SEC"
  fi
done < "$CASES_FILE"

echo ""
echo "================================"
echo "PASS: ${pass}  FAIL: ${fail}  TOTAL: $((pass + fail))"

if [[ ${fail} -gt 0 ]]; then
  echo ""
  echo "Failed cases:"
  for c in "${failed_cases[@]}"; do
    echo "  ${c}"
  done
  exit 1
fi
