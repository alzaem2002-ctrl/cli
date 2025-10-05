#!/usr/bin/env bash
set -euo pipefail

: "${GH_TOKEN:?GH_TOKEN required}"
: "${DEPLOY_HOST:?DEPLOY_HOST required}"
: "${DEPLOY_USER:?DEPLOY_USER required}"
: "${REPO:?REPO required}"

DEPLOY_PORT="${DEPLOY_PORT:-22}"
ALLOW_WRITE="${ALLOW_WRITE:-false}"
KEY_TITLE="${KEY_TITLE:-ci-deploy-key-$(date -u +%Y%m%dT%H%M%SZ)}"

WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

PRIVATE_KEY_PATH="$WORKDIR/$KEY_TITLE"
PUBLIC_KEY_PATH="$PRIVATE_KEY_PATH.pub"

ssh-keygen -o -a 100 -t ed25519 -C "$KEY_TITLE" -f "$PRIVATE_KEY_PATH" -N "" >/dev/null

PRIVATE_KEY="$(<"$PRIVATE_KEY_PATH")"
PUBLIC_KEY="$(<"$PUBLIC_KEY_PATH")"

# Append the public key to remote authorized_keys if missing (no quoting issues)
ssh -p "$DEPLOY_PORT" -o StrictHostKeyChecking=no "$DEPLOY_USER@$DEPLOY_HOST" 'set -euo pipefail
SSH_DIR="$HOME/.ssh"
AUTH="$SSH_DIR/authorized_keys"
umask 077
mkdir -p "$SSH_DIR"
touch "$AUTH"
chmod 700 "$SSH_DIR"
chmod 600 "$AUTH"
IFS= read -r KEY
if ! grep -qxF "$KEY" "$AUTH"; then
  printf "%s\n" "$KEY" >> "$AUTH"
fi
' <<<"$PUBLIC_KEY"

# Test SSH using the generated key
SSH_TEST_OUTPUT="$(ssh -i "$PRIVATE_KEY_PATH" -p "$DEPLOY_PORT" -o BatchMode=yes -o StrictHostKeyChecking=no "$DEPLOY_USER@$DEPLOY_HOST" 'echo SSH_OK:$HOSTNAME' 2>&1 || true)"

# Install GitHub CLI if missing (Debian/Ubuntu)
if ! command -v gh >/dev/null 2>&1; then
  if command -v apt-get >/dev/null 2>&1; then
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -y
    apt-get install -y gh >/dev/null
  else
    echo "ERROR: gh not found and automatic install unsupported on this OS." >&2
    exit 1
  fi
fi

# Authenticate GitHub CLI
printf %s "$GH_TOKEN" | gh auth login --with-token >/dev/null

# Determine read-only/read-write for deploy key
READ_ONLY_FLAG=true
if [[ "${ALLOW_WRITE,,}" == "true" ]]; then
  READ_ONLY_FLAG=false
fi

# Create deploy key in the repository (ignore if already exists)
gh api --method POST "/repos/$REPO/keys" \
  -f title="$KEY_TITLE" \
  -f key="$PUBLIC_KEY" \
  -f read_only="$READ_ONLY_FLAG" >/dev/null 2>&1 || true

# Store private key as repository secret DEPLOY_KEY
printf %s "$PRIVATE_KEY" | gh secret set DEPLOY_KEY -b- --repo "$REPO"

echo "✅ Completed. SSH_OK: $SSH_TEST_OUTPUT"