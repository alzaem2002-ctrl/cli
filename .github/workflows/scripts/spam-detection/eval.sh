#!/bin/bash

# Run the eval tests for the spam detection AI model.
#
# This script must be run from the root directory of the repository.

set -euo pipefail

_prompt_file=".github/workflows/scripts/spam-detection/prompt.yml"
_generate_sys_prompt_script=".github/workflows/scripts/spam-detection/generate-sys-prompt.sh"

_system_prompt="$($_generate_sys_prompt_script)"
_updated_prompt_file="$(_value="$_system_prompt" yq eval '.messages[0].content = strenv(_value)' "$_prompt_file")"

# We should be able to just run the following command:
#
# ```
# gh models eval <(echo "$_updated_prompt_file")
# ```
#
# But since `gh-models` does not throttle the rate of API requests, we need to
# modify the extension code and introduce a deliberate delay between the runs.
# Here, we assume a binary of the `gh-models` extension (with appropriate
# throttling) is available in the root directory of the repository and we're
# calling it directly (not though `gh`).
./gh-models eval <(echo "$_updated_prompt_file")
