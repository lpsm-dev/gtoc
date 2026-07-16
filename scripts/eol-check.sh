#!/usr/bin/env bash
# Checks the products pinned in this repository against the endoflife.date
# API v1 (https://endoflife.date/docs/api/v1/) and reports their lifecycle
# status. Versions are parsed from the repository's sources of truth, never
# hardcoded:
#   - go     <- .go-version
#   - alpine <- Dockerfile (final stage base image)
#   - ubuntu <- pinned runner labels in .github/workflows/
#
# Output: a markdown table on stdout (and in $GITHUB_STEP_SUMMARY when set),
# a markdown report at eol-report.md, and step outputs (has_findings, worst)
# in $GITHUB_OUTPUT when set. Always exits 0; the workflow decides whether to
# fail based on the outputs.
#
# Environment:
#   EOL_WARN_DAYS  days before EOL to start warning (default: 90)
#   EOL_API_BASE   API base URL override, mainly for tests
set -euo pipefail

API_BASE="${EOL_API_BASE:-https://endoflife.date/api/v1/products}"
WARN_DAYS="${EOL_WARN_DAYS:-90}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORT_FILE="${EOL_REPORT_FILE:-eol-report.md}"

rows=""
worst="ok" # escalates: ok -> warn -> eol

# discover prints one "product cycle source" line per pinned product.
discover() {
  local go_version alpine_tag ubuntu_pin
  go_version="$(tr -d '[:space:]' <"$ROOT/.go-version")"
  printf 'go %s .go-version\n' "$(printf '%s' "$go_version" | cut -d. -f1,2)"

  alpine_tag="$(sed -n 's/^FROM alpine:\([0-9][0-9.]*\).*/\1/p' "$ROOT/Dockerfile" | head -1)"
  if [ -n "$alpine_tag" ]; then
    printf 'alpine %s Dockerfile\n' "$(printf '%s' "$alpine_tag" | cut -d. -f1,2)"
  fi

  ubuntu_pin="$(grep -rhoE 'ubuntu-[0-9]+\.[0-9]+' "$ROOT/.github/workflows" | sort -u | head -1 | sed 's/ubuntu-//')"
  if [ -n "$ubuntu_pin" ]; then
    printf 'ubuntu %s workflows\n' "$ubuntu_pin"
  fi
}

escalate() {
  case "$1:$worst" in
    eol:*) worst="eol" ;;
    warn:ok) worst="warn" ;;
  esac
}

# check queries the API for one product cycle and appends a table row.
check() {
  local product="$1" cycle="$2" source="$3"
  local json fields is_eol is_maintained eol_from latest days status

  if ! json="$(curl -fsSL --retry 3 --max-time 30 "$API_BASE/$product/releases/$cycle")"; then
    rows+="| $product | $cycle | ⚠️ lookup failed | - | - | $source |"$'\n'
    escalate "warn"
    return
  fi

  fields="$(jq -r '.result | [
    (.isEol | tostring),
    (.isMaintained | tostring),
    (.eolFrom // "-"),
    (.latest.name // "-"),
    (if .eolFrom then
       ((((.eolFrom | strptime("%Y-%m-%d") | mktime) - now) / 86400) | floor | tostring)
     else "-" end)
  ] | join(" ")' <<<"$json")"
  read -r is_eol is_maintained eol_from latest days <<<"$fields"

  if [ "$is_eol" = "true" ]; then
    status="❌ end-of-life"
    escalate "eol"
  elif [ "$days" != "-" ] && [ "$days" -le "$WARN_DAYS" ]; then
    status="⚠️ EOL in ${days}d"
    escalate "warn"
  elif [ "$is_maintained" != "true" ]; then
    status="⚠️ unmaintained"
    escalate "warn"
  else
    status="✅ maintained"
  fi

  rows+="| $product | $cycle | $status | $eol_from | $latest | $source |"$'\n'
}

main() {
  while read -r product cycle source; do
    check "$product" "$cycle" "$source"
  done < <(discover)

  local report
  report="## EOL watch — pinned products

Data: [endoflife.date](https://endoflife.date) API v1 · warn window: ${WARN_DAYS} days

| Product | Pinned cycle | Status | EOL date | Latest release | Pinned in |
| ------- | ------------ | ------ | -------- | -------------- | --------- |
${rows}"

  printf '%s\n' "$report"
  printf '%s\n' "$report" >"$REPORT_FILE"
  if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
    printf '%s\n' "$report" >>"$GITHUB_STEP_SUMMARY"
  fi

  local has_findings="false"
  if [ "$worst" != "ok" ]; then
    has_findings="true"
  fi
  if [ -n "${GITHUB_OUTPUT:-}" ]; then
    {
      printf 'has_findings=%s\n' "$has_findings"
      printf 'worst=%s\n' "$worst"
    } >>"$GITHUB_OUTPUT"
  fi
}

main "$@"
