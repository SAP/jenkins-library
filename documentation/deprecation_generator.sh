#!/usr/bin/env bash
# bash

set -euo pipefail

# Configurable paths
STEPS_FILE="${1:-/yourAbsoluteDirPathForStepsYamlFile}"
DOCS_DIR="${2:-/yourAbsoluteDirPathForMDDocsFiles}"

# Deprecation block (exact text)
DEPR_TITLE=$'!!! warning "Deprecation notice"'
DEPR_BODY=$'This step will soon be deprecated!'

# Load steps from YAML into a plain array (compatible with macOS bash)
STEPS=()
if [[ -f "$STEPS_FILE" ]]; then
  while IFS= read -r line; do
    if [[ $line =~ ^[[:space:]]*-[[:space:]]*(.+) ]]; then
      name="${BASH_REMATCH[1]}"
      name="${name//\"/}"   # strip quotes if present
      name="${name//\'/}"
      name="${name// /}"    # strip spaces (assumes names have no spaces)
      STEPS+=("$name")
    fi
  done < <(grep -E '^[[:space:]]*-[[:space:]]*' "$STEPS_FILE" || true)
else
  echo "Steps file \`$STEPS_FILE\` not found."
  exit 1
fi

# Ensure docs dir exists
if [[ ! -d "$DOCS_DIR" ]]; then
  echo "Docs directory \`$DOCS_DIR\` not found."
  exit 1
fi

# Helper: check membership in STEPS
contains() {
  local seek="$1"
  for s in "${STEPS[@]}"; do
    [[ "$s" == "$seek" ]] && return 0
  done
  return 1
}

# Helper: insert deprecation block two lines under header "# <StepName>"
insert_deprecation() {
  local file="$1"
  if grep -q 'Deprecation notice' "$file"; then
    echo "SKIP (already deprecated): \`$file\`"
    return
  fi

  awk -v t="$DEPR_TITLE" -v b="$DEPR_BODY" '
    BEGIN { inserted=0 }
    {
      if (!inserted && $0 == "# ${docGenStepName}") {
        print $0
        print ""     # blank line
        print t
        print b
        inserted=1
        next
      }
      print $0
    }
  ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
  echo "INSERTED: \`$file\`"
}

# Helper: remove deprecation blocks from file (all occurrences)
remove_deprecation() {
  local file="$1"
  if ! grep -q 'Deprecation notice' "$file"; then
    echo "SKIP (no deprecation): \`$file\`"
    return
  fi

awk '
  BEGIN { skip=0 }
  /^!!! warning "Deprecation notice"$/ { skip=2; next }   # drop this line and the following line
  skip > 0 { skip--; next }
  { print }
' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
  echo "REMOVED: \`$file\`"
}

# 1) For each listed step: find its .md and insert block if missing
for step in "${STEPS[@]}"; do
  mdfile="$DOCS_DIR/${step}.md"
  if [[ -f "$mdfile" ]]; then
    insert_deprecation "$mdfile" "$step"
  else
    found="$(find "$DOCS_DIR" -maxdepth 1 -type f -name "*${step}*.md" -print -quit || true)"
    if [[ -n "$found" ]]; then
      insert_deprecation "$found" "$step"
    else
      echo "WARN: step \`$step\` has no matching .md in \`$DOCS_DIR\`"
    fi
  fi
done

# 2) For each .md in docs dir: if it contains Deprecation and its basename is NOT in STEPS -> remove
while IFS= read -r file; do
  [[ -f "$file" ]] || continue
  base="$(basename "$file" .md)"
  if ! contains "$base" && grep -q 'Deprecation notice' "$file"; then
    remove_deprecation "$file"
  fi
done < <(find "$DOCS_DIR" -maxdepth 1 -type f -name '*.md' -print)

echo "Done."
