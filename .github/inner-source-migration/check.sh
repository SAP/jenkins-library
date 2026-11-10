#!/usr/bin/env bash
# Classifies the changed files of a pull request against moved.txt / not-moved.txt
# and renders the result for the workflow run and the sticky PR comment.
#
# Usage:  <changed file paths, one per line> | check.sh [LIST_DIR]
# Local:  git diff --name-only origin/master... | .github/inner-source-migration/check.sh
#
# Result per file (first match wins): moved -> kept (not moved) -> shared.
# Status: fail if any moved file, warn if any shared file, ok otherwise.
#
# Writes:
#   stdout                 log banner, ::error/::warning/::notice annotations, status=<status>
#   $GITHUB_STEP_SUMMARY   job summary (default: stdout)
#   $COMMENT_FILE          sticky PR comment body (default: not written)
#   $GITHUB_OUTPUT         status=ok|warn|fail (default: not written)
set -euo pipefail

ARCHIVE_DATE="2026-12-31"
MARKER="<!-- inner-source-migration-check -->"
MAX_LISTED_SHARED=30

list_dir="${1:-$(dirname "$0")}"
summary_file="${GITHUB_STEP_SUMMARY:-/dev/stdout}"
comment_file="${COMMENT_FILE:-/dev/null}"
output_file="${GITHUB_OUTPUT:-/dev/null}"
repo="${GITHUB_REPOSITORY:-SAP/jenkins-library}"
list_url="${GITHUB_SERVER_URL:-https://github.com}/$repo/blob/${GITHUB_BASE_REF:-master}/.github/inner-source-migration/moved.txt"
inner_source="SAP-internal Piper library"
inner_source_caps="SAP-INTERNAL PIPER LIBRARY"
inner_source_link="<https://url.sap/piper> (SAP-internal)"

for list in moved.txt not-moved.txt; do
  if [[ ! -r "$list_dir/$list" ]]; then
    echo "::error::cannot read $list_dir/$list" >&2
    exit 2
  fi
done

# Backticks would break the markdown rendering of file names; they never occur in real paths.
classified=$(tr -d '`\r' | awk -v moved="$list_dir/moved.txt" -v kept="$list_dir/not-moved.txt" '
  function load(file, cls,   line) {
    while ((getline line < file) > 0) {
      sub(/#.*/, "", line); gsub(/[ \t]/, "", line)
      if (line == "") continue
      if (index(line, "/")) pre[cls, ++npre[cls]] = line
      else name[cls, ++nname[cls]] = line
    }
    close(file)
  }
  function owner(f, cls,   i, n) {
    for (i = 1; i <= nname[cls]; i++) {
      n = name[cls, i]
      if (f == "cmd/" n ".go" || f == "cmd/" n "_test.go" || f == "cmd/" n "_generated.go" || f == "cmd/" n "_generated_test.go" ||
          f == "resources/metadata/" n ".yaml" || f == "vars/" n ".groovy" || f == "documentation/docs/steps/" n ".md" ||
          tolower(f) == tolower("test/groovy/" n "Test.groovy")) return n
    }
    for (i = 1; i <= npre[cls]; i++) if (index(f, pre[cls, i]) == 1) return pre[cls, i]
    return ""
  }
  BEGIN { load(moved, "moved"); load(kept, "kept") }
  $0 == "" || seen[$0]++ { next }
  {
    if ((o = owner($0, "moved")) != "") print "moved\t" o "\t" $0
    else if ((o = owner($0, "kept")) != "") print "kept\t" o "\t" $0
    else print "shared\t-\t" $0
  }')

count() { printf '%s\n' "$classified" | awk -F'\t' -v c="$1" '$1 == c { n++ } END { print n + 0 }'; }
n_moved=$(count moved)
n_kept=$(count kept)
n_shared=$(count shared)

if ((n_moved > 0)); then status=fail
elif ((n_shared > 0)); then status=warn
else status=ok
fi

# "<owner>\t<file>, <file>" per moved step/path, sorted by owner
moved_groups=$(printf '%s\n' "$classified" | awk -F'\t' '
  $1 == "moved" { if (!($2 in files)) order[++n] = $2; files[$2] = files[$2] (files[$2] == "" ? "" : ", ") $3 }
  END { for (i = 1; i <= n; i++) print order[i] "\t" files[order[i]] }' | sort)
shared_files=$(printf '%s\n' "$classified" | awk -F'\t' '$1 == "shared" { print $3 }')

# Workflow command data must not contain raw %, CR or LF.
esc() { local s=${1//\%/%25}; s=${s//$'\r'/%0D}; printf '%s' "${s//$'\n'/%0A}"; }

banner() {
  local hr="################################################################################"
  local line
  echo "$hr"
  printf '##%76s##\n' ''
  for line in "$@"; do printf '##   %-72s ##\n' "$line"; done
  printf '##%76s##\n' ''
  echo "$hr"
}

moved_table() {
  echo "| Moved step / path | Changed files |"
  echo "| --- | --- |"
  local owner files
  while IFS=$'\t' read -r owner files; do
    [[ -n "$owner" ]] || continue
    echo "| \`$owner\` | \`${files//, /\`<br>\`}\` |"
  done <<< "$moved_groups"
}

shared_list() {
  local shown=0 file
  while IFS= read -r file; do
    [[ -n "$file" ]] || continue
    if ((shown == MAX_LISTED_SHARED)); then
      echo "- … and $((n_shared - shown)) more"
      break
    fi
    echo "- \`$file\`"
    shown=$((shown + 1))
  done <<< "$shared_files"
}

shared_section() {
  ((n_shared > 0)) || return 0
  echo
  echo "#### ⚠️ Shared code: $n_shared file(s)"
  echo
  echo "These files are not owned by a single step that stays here. They are also used by steps that moved to the $inner_source, so the change likely needs to be ported there too."
  echo
  shared_list
}

body() {
  case "$status" in
    fail)
      echo "> [!CAUTION]"
      echo "> ### 🚫 THIS PULL REQUEST CHANGES STEPS THAT MOVED TO THE $inner_source_caps. CHANGES TO THESE STEPS ARE NO LONGER ACCEPTED IN THIS REPOSITORY."
      echo
      echo "\`$repo\` will be **archived on $ARCHIVE_DATE**. The steps below moved to the $inner_source, where their development continues. Please make this change there instead: $inner_source_link."
      echo
      echo "Until the archive date, pull requests are only accepted for steps that were **not** moved. Full list: [moved.txt]($list_url)."
      echo
      moved_table
      shared_section
      ;;
    warn)
      echo "> [!WARNING]"
      echo "> ### ⚠️ THIS PULL REQUEST CHANGES SHARED CODE THAT ALSO EXISTS IN THE $inner_source_caps."
      echo
      echo "\`$repo\` will be **archived on $ARCHIVE_DATE**. This change is not blocked, but it likely needs to be ported to the $inner_source as well: $inner_source_link."
      shared_section
      ;;
    ok)
      echo "> [!NOTE]"
      echo "> ✅ This pull request does not change any steps that moved to the $inner_source. \`$repo\` will be archived on $ARCHIVE_DATE."
      ;;
  esac
}

# --- log + annotations ---
case "$status" in
  fail)
    banner "THIS PULL REQUEST CANNOT BE MERGED." \
      "IT CHANGES STEPS THAT MOVED TO THE SAP-INTERNAL PIPER LIBRARY." \
      "CHANGES TO THESE STEPS ARE NO LONGER ACCEPTED IN THIS REPOSITORY." \
      "THIS REPOSITORY WILL BE ARCHIVED ON $ARCHIVE_DATE."
    while IFS=$'\t' read -r owner files; do
      [[ -n "$owner" ]] || continue
      echo "::error title=MOVED TO INNER SOURCE - $(esc "$owner")::$(esc "$files")"
    done <<< "$moved_groups"
    ;;
  warn)
    banner "THIS PULL REQUEST CHANGES SHARED CODE." \
      "THE SAME CODE ALSO LIVES IN THE SAP-INTERNAL PIPER LIBRARY." \
      "THE CHANGE IS NOT BLOCKED, BUT LIKELY NEEDS TO BE PORTED THERE TOO." \
      "THIS REPOSITORY WILL BE ARCHIVED ON $ARCHIVE_DATE."
    ;;
esac
if ((n_shared > 0)); then
  echo "::warning title=SHARED CODE - ALSO PORT TO INNER SOURCE::$n_shared changed file(s) are shared with steps that moved to the $inner_source. See the job summary for the list."
fi
if [[ "$status" == ok ]]; then
  echo "::notice title=NO MOVED STEPS CHANGED::Only steps that stay in this repository are changed. Accepted until the archive date $ARCHIVE_DATE."
fi
echo "moved=$n_moved kept=$n_kept shared=$n_shared"
echo "status=$status"

# --- job summary ---
{
  case "$status" in
    fail) echo "# 🚫 THIS PR CANNOT BE MERGED: IT CHANGES STEPS MOVED TO INNER SOURCE" ;;
    warn) echo "# ⚠️ SHARED CODE CHANGED: ALSO PORT IT TO INNER SOURCE" ;;
    ok) echo "# ✅ NO MOVED STEPS CHANGED" ;;
  esac
  echo
  body
  echo
  echo "_Changed files: $n_moved moved, $n_kept not moved, $n_shared shared._"
} >> "$summary_file"

# --- sticky comment body ---
{
  echo "$MARKER"
  body
} > "$comment_file"

echo "status=$status" >> "$output_file"
