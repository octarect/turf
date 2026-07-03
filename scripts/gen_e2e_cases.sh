#!/usr/bin/env bash
set -euo pipefail

TURF_BIN="${TURF_BIN:-turf}"
SLEEP_SEC="${SLEEP_SEC:-1}"
MAX_TOTAL=100
MAX_PER_MONTH=3

usage() {
  echo "Usage: $(basename "$0") OUTPUT_FILE [--bin PATH] [--sleep SEC]"
  echo ""
  echo "Arguments:"
  echo "  OUTPUT_FILE     Path to write the generated cases TSV"
  echo ""
  echo "Options:"
  echo "  --bin PATH      Path to turf binary (default: turf)"
  echo "  --sleep SEC     Sleep between requests in seconds (default: 1)"
  exit 1
}

if [[ $# -lt 1 ]]; then usage; fi
OUTPUT_FILE="$1"; shift

while [[ $# -gt 0 ]]; do
  case "$1" in
    --bin)    TURF_BIN="$2";  shift 2 ;;
    --sleep)  SLEEP_SEC="$2"; shift 2 ;;
    -h|--help) usage ;;
    *) echo "Unknown option: $1"; usage ;;
  esac
done

if ! command -v "$TURF_BIN" &>/dev/null; then
  echo "Error: turf binary not found: $TURF_BIN" >&2
  exit 1
fi
if ! command -v jq &>/dev/null; then
  echo "Error: jq is required" >&2
  exit 1
fi

shuffle() {
  jq -c 'to_entries | sort_by(.key * 2654435761 % 4294967296) | map(.value)'
}

declare -A seen_combos
total=0

{
  echo -e "# DATE\tCOURSE\tRACE_NO\tSURFACE\tDISTANCE"
} > "$OUTPUT_FILE"

START_YEAR=2020
START_MONTH=1
END_YEAR=2026
END_MONTH=6

year=$START_YEAR
month=$START_MONTH

echo "Generating E2E test cases..." >&2
echo "Target: $MAX_TOTAL cases, up to $MAX_PER_MONTH per month" >&2
echo "Output: $OUTPUT_FILE" >&2
echo "" >&2

while true; do
  if [[ $year -gt $END_YEAR ]]; then break; fi
  if [[ $year -eq $END_YEAR && $month -gt $END_MONTH ]]; then break; fi
  if [[ $total -ge $MAX_TOTAL ]]; then break; fi

  ym=$(printf "%04d-%02d" "$year" "$month")
  echo "[${ym}] Fetching fixtures..." >&2

  fixtures_json=$("$TURF_BIN" fixtures --month "$ym" -o json 2>/dev/null || echo "[]")
  sleep "$SLEEP_SEC"

  fixture_count=$(echo "$fixtures_json" | jq 'length')
  if [[ $fixture_count -eq 0 ]]; then
    echo "[${ym}] No fixtures found, skipping." >&2
    month=$((month + 1))
    if [[ $month -gt 12 ]]; then month=1; year=$((year + 1)); fi
    continue
  fi

  count_this_month=0

  shuffled_fixtures=$(echo "$fixtures_json" | jq -c '.[]' | shuf)

  while IFS= read -r fixture; do
    if [[ $count_this_month -ge $MAX_PER_MONTH ]]; then break; fi
    if [[ $total -ge $MAX_TOTAL ]]; then break; fi

    date=$(echo "$fixture" | jq -r '.date')
    course=$(echo "$fixture" | jq -r '.course')

    echo "[${ym}] Fetching races for $date $course..." >&2

    races_json=$("$TURF_BIN" races --date "$date" --course "$course" -o json 2>/dev/null || echo "[]")
    sleep "$SLEEP_SEC"

    race_count=$(echo "$races_json" | jq 'length')
    if [[ $race_count -eq 0 ]]; then
      echo "[${ym}] No races found for $date $course, skipping." >&2
      continue
    fi

    chosen_num=""
    chosen_surface=""
    chosen_distance=""
    fallback_num=""
    fallback_surface=""
    fallback_distance=""

    shuffled_races=$(echo "$races_json" | jq -c '.[]' | shuf)

    while IFS= read -r race; do
      num=$(echo "$race" | jq -r '.num')
      surface=$(echo "$race" | jq -r '.surface')
      distance=$(echo "$race" | jq -r '.distance')
      combo="${surface}_${distance}"

      if [[ -z "$fallback_num" ]]; then
        fallback_num="$num"
        fallback_surface="$surface"
        fallback_distance="$distance"
      fi

      if [[ -z "${seen_combos[$combo]+_}" ]]; then
        chosen_num="$num"
        chosen_surface="$surface"
        chosen_distance="$distance"
        break
      fi
    done <<< "$shuffled_races"

    if [[ -z "$chosen_num" ]]; then
      chosen_num="$fallback_num"
      chosen_surface="$fallback_surface"
      chosen_distance="$fallback_distance"
    fi

    combo="${chosen_surface}_${chosen_distance}"
    seen_combos[$combo]=1

    echo -e "${date}\t${course}\t${chosen_num}\t${chosen_surface}\t${chosen_distance}" >> "$OUTPUT_FILE"
    echo "[${ym}] Added: $date $course race=$chosen_num (${chosen_surface} ${chosen_distance}m)" >&2

    count_this_month=$((count_this_month + 1))
    total=$((total + 1))

  done <<< "$shuffled_fixtures"

  echo "[${ym}] Done ($count_this_month cases added, total=$total)" >&2

  month=$((month + 1))
  if [[ $month -gt 12 ]]; then month=1; year=$((year + 1)); fi
done

echo "" >&2
echo "Generated $total cases -> $OUTPUT_FILE" >&2
