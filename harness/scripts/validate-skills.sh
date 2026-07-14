#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
failures=0

for skill in "$root"/harness/skills/*; do
  [[ -d "$skill" ]] || continue
  file="$skill/SKILL.md"
  if [[ ! -f "$file" ]]; then
    echo "missing SKILL.md: $skill"
    failures=$((failures + 1))
    continue
  fi

  dir_name="$(basename "$skill")"
  if ! grep -qE "^name: ${dir_name}$" "$file"; then
    echo "name does not match directory: $file"
    failures=$((failures + 1))
  fi

  for required in "^---$" "^description:" "^metadata:" "^  author:" "^  version:" "^  mcp-server:" "^license:" "^compatibility:" "^## Instructions$" "^## Examples$" "^## Performance Notes$"; do
    if ! grep -qE "$required" "$file"; then
      echo "missing required pattern '$required': $file"
      failures=$((failures + 1))
    fi
  done

  if ! grep -qE "^## (Troubleshooting|Error Handling)$" "$file"; then
    echo "missing troubleshooting/error handling section: $file"
    failures=$((failures + 1))
  fi
done

if [[ "$failures" -gt 0 ]]; then
  echo "skill validation failed: $failures issue(s)"
  exit 1
fi

echo "skill validation passed"
