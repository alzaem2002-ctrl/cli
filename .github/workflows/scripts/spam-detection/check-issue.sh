#!/bin/bash

# Check if an issue is spam or not and output "PASS" (not spam) or "FAIL" (spam).
#
# Regardless of the spam detection result, the script always exits with a zero
# exit code, unless there's a runtime error.
#
# This script must be run from the root directory of the repository.

set -euo pipefail

_prompt_file=".github/workflows/scripts/spam-detection/prompt.yml"
_generate_sys_prompt_script=".github/workflows/scripts/spam-detection/generate-sys-prompt.sh"
_generate_prompt_script=".github/workflows/scripts/spam-detection/generate-prompt.sh"

_issue_url="$1"
if [[ -z "$_issue_url" ]]; then
    echo "error: issue URL is empty" >&2
    exit 1
fi

_issue="$(gh issue view --json title,body "$_issue_url")"

_issue_body="$(jq -r ".body" <<< "$_issue")"
_issue_title="$(jq -r ".title" <<< "$_issue")"

_system_prompt="$($_generate_sys_prompt_script)"
_input_prompt="$($_generate_prompt_script "$_issue_title" "$_issue_body")"

_updated_prompt_file_content="$(
    cat "$_prompt_file" |
    yq eval 'del(.testData, .evaluators)' | # drop test data
    _system="$_system_prompt" _input="$_input_prompt" yq eval ".messages[0].content = strenv(_system) | .messages[1].content = strenv(_input)"
)"

gh extension install github/gh-models 2>/dev/null

_result="$(gh models run --file <(echo "$_updated_prompt_file_content") | cat)"

if [[ "$_result" != "PASS" && "$_result" != "FAIL" ]]; then
    echo "error: expected PASS or FAIL but got an unexpected result: $_result" >&2
    exit 1
fi

echo "$_result"
