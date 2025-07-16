#!/bin/bash

# Generate the prompt for the spam detection AI model. The issue title and body
# should be provided as arguments.
#
# This script must be run from the root directory of the repository.

set -euo pipefail

_issue_title="$1"
_issue_body="$2"

_prompt="
<TITLE>
$_issue_title
</TITLE>

<BODY>
$_issue_body
</BODY>
"

echo "$_prompt"
