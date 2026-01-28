#!/usr/bin/env bash
# bash

set -euo pipefail

# Configurable paths
STEPS_FILE="${1:-/Users/C5399877/Library/CloudStorage/OneDrive-SAPSE/Documents/Piper/jenkins-library/documentation/deprecated_steps.yaml}"
DOCS_DIR="${2:-documentation/docs/steps}"

# Deprecation block (exact text)
DEPR_BLOCK=$'!!! warning "Deprecation notice"\n    This step will soon be deprecated!'

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
  local step="$2"

  if grep -q 'Deprecation notice' "$file"; then
    echo "SKIP (already deprecated): \`$file\`"
    return
  fi

  awk -v block="$DEPR_BLOCK" '
    BEGIN { inserted=0 }
    {
      if (!inserted && $0 == # ${docGenStepName) {
        print $0
        print ""     # first blank line
        print ""     # second blank line
        n = split(block, b, "\n")
        for (i=1;i<=n;i++) print b[i]
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
    BEGIN { inblock=0 }
    # start of block
    /^!!! warning "Deprecation notice"$/ { inblock=1; next }
    # while in block skip indented lines (4 spaces or tabs) and empty lines that belong to block
    inblock && ($0 ~ /^[[:space:]]/ || $0 == "") { next }
    # first non-indented/non-empty line after block -> end block and print it
    inblock { inblock=0; print; next }
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
