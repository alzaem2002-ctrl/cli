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

# Secure cleanup: shred key files if available, then remove temp dir
secure_cleanup() {
  set +e
  if command -v shred >/dev/null 2>&1; then
    [ -f "$PRIVATE_KEY_PATH" ] && shred -u "$PRIVATE_KEY_PATH" 2>/dev/null || true
    [ -f "$PUBLIC_KEY_PATH" ] && shred -u "$PUBLIC_KEY_PATH" 2>/dev/null || true
  else
    rm -f "$PRIVATE_KEY_PATH" "$PUBLIC_KEY_PATH" 2>/dev/null || true
  fi
  rm -rf "$WORKDIR" 2>/dev/null || true
}
trap secure_cleanup EXIT

PRIVATE_KEY_PATH="$WORKDIR/$KEY_TITLE"
PUBLIC_KEY_PATH="$PRIVATE_KEY_PATH.pub"

ssh-keygen -o -a 100 -t ed25519 -C "$KEY_TITLE" -f "$PRIVATE_KEY_PATH" -N "" >/dev/null
echo "✅ SSH keypair generated"

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
echo "✅ Public key appended to $DEPLOY_USER@$DEPLOY_HOST:~/.ssh/authorized_keys (if not present)"

# Test SSH using the generated key
SSH_TEST_OUTPUT="$(ssh -i "$PRIVATE_KEY_PATH" -p "$DEPLOY_PORT" -o BatchMode=yes -o StrictHostKeyChecking=no "$DEPLOY_USER@$DEPLOY_HOST" 'echo SSH_OK:$HOSTNAME' 2>&1 || true)"
if echo "$SSH_TEST_OUTPUT" | grep -q 'SSH_OK:'; then
  echo "✅ SSH connectivity verified"
else
  echo "❌ SSH connectivity failed: $SSH_TEST_OUTPUT" >&2
  exit 1
fi

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
if command -v gh >/dev/null 2>&1; then
  echo "✅ gh installed"
fi

# Authenticate GitHub CLI
printf %s "$GH_TOKEN" | gh auth login --with-token >/dev/null
echo "✅ gh authenticated"

# Determine read-only/read-write for deploy key
READ_ONLY_FLAG=true
if [[ "${ALLOW_WRITE,,}" == "true" ]]; then
  READ_ONLY_FLAG=false
fi

# Create deploy key (read-only by default). If it already exists, skip creation.
EXISTING_KEY_ID="$(gh api "/repos/$REPO/keys" --jq ".[] | select(.key == \"$PUBLIC_KEY\") | .id" 2>/dev/null || true)"
if [ -z "$EXISTING_KEY_ID" ]; then
  gh api --method POST "/repos/$REPO/keys" \
    -f title="$KEY_TITLE" \
    -f key="$PUBLIC_KEY" \
    -f read_only="$READ_ONLY_FLAG" >/dev/null
  echo "✅ Deploy key created"
else
  echo "✅ Deploy key created (already existed)"
fi

# Store private key as repository secret DEPLOY_KEY
printf %s "$PRIVATE_KEY" | gh secret set DEPLOY_KEY -b- --repo "$REPO"
echo "✅ DEPLOY_KEY secret set"

echo "✅ Completed. SSH_OK: $SSH_TEST_OUTPUT"