#!/usr/bin/env bash
set -euo pipefail

API="https://api.github.com"
REPO="${REPO:?REPO required}"
BRANCH="${BRANCH:?BRANCH required}"
TITLE="${TITLE:?TITLE required}"
BODY="${BODY:-}"
TOKEN="${TOKEN:-${GITHUB_PAT:-${GH_TOKEN:-}}}"

if [[ -z "${TOKEN}" ]]; then
  echo "❌ Missing GitHub token (TOKEN/GITHUB_PAT/GH_TOKEN)"
  exit 1
fi

owner="${REPO%%/*}"

# Compose common headers
AUTH_HEADER=( -H "Authorization: Bearer ${TOKEN}" -H "Accept: application/vnd.github+json" -H "Content-Type: application/json" )

echo "🚀 Starting auto-deploy via GitHub API..."

# 1) Find existing PR
resp_create=$(curl -sS "${AUTH_HEADER[@]}" \
  "${API}/repos/${REPO}/pulls?head=${owner}:${BRANCH}&base=main&state=open")
pr_number=$(printf "%s" "$resp_create" | sed -n 's/.*"number"[[:space:]]*:[[:space:]]*\([0-9][0-9]*\).*/\1/p' | head -n1 || true)

# If not found, create PR
if [[ -z "${pr_number:-}" ]]; then
  create_payload=$(printf '{"title":"%s","head":"%s","base":"main","body":"%s","maintainer_can_modify":true}' \
    "$TITLE" "$BRANCH" "$BODY")
  resp_create=$(curl -sS "${AUTH_HEADER[@]}" -X POST \
    "${API}/repos/${REPO}/pulls" -d "$create_payload")
  pr_number=$(printf "%s" "$resp_create" | sed -n 's/.*"number"[[:space:]]*:[[:space:]]*\([0-9][0-9]*\).*/\1/p' | head -n1 || true)
  pr_url=$(printf "%s" "$resp_create" | sed -n 's/.*"html_url"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1 || true)
else
  pr_url=$(printf "%s" "$resp_create" | sed -n 's/.*"html_url"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1 || true)
fi

if [[ -z "${pr_number:-}" ]]; then
  echo "❌ Failed to create or find PR"
  printf "%s\n" "$resp_create"
  exit 2
fi

echo "🔗 PR: ${pr_url:-unknown} (#${pr_number})"

# 2) Merge PR (merge commit)
merge_payload='{"merge_method":"merge","commit_title":"Auto-merged refactored autodeploy.yml."}'
resp_merge=$(curl -sS "${AUTH_HEADER[@]}" -X PUT \
  "${API}/repos/${REPO}/pulls/${pr_number}/merge" -d "$merge_payload" || true)
merged=$(printf "%s" "$resp_merge" | sed -n 's/.*"merged"[[:space:]]*:[[:space:]]*\(true\|false\).*/\1/p' | head -n1 || true)
if [[ "${merged}" != "true" ]]; then
  # Fallback: base branch changed; try updating PR branch then re-merge
  if printf "%s" "$resp_merge" | grep -qi 'Base branch was modified'; then
    echo "ℹ️ Base branch changed; updating PR branch..."
    curl -sS "${AUTH_HEADER[@]}" -X PUT \
      "${API}/repos/${REPO}/pulls/${pr_number}/update-branch" -d '{}' >/dev/null || true

    echo "⏳ Waiting for PR to become mergeable..."
    end_update=$((SECONDS+300))
    while [[ $SECONDS -lt $end_update ]]; do
      pr_json=$(curl -sS "${AUTH_HEADER[@]}" \
        "${API}/repos/${REPO}/pulls/${pr_number}")
      mergeable_state=$(printf "%s" "$pr_json" | sed -n 's/.*"mergeable_state"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1 || true)
      if [[ "$mergeable_state" == "clean" ]]; then
        break
      fi
      sleep 5
    done

    echo "🔁 Retrying merge after update..."
    resp_merge=$(curl -sS "${AUTH_HEADER[@]}" -X PUT \
      "${API}/repos/${REPO}/pulls/${pr_number}/merge" -d "$merge_payload" || true)
    merged=$(printf "%s" "$resp_merge" | sed -n 's/.*"merged"[[:space:]]*:[[:space:]]*\(true\|false\).*/\1/p' | head -n1 || true)
    if [[ "${merged}" != "true" ]]; then
      echo "❌ Merge failed after update"
      printf "%s\n" "$resp_merge"
      exit 3
    fi
  else
    echo "❌ Merge failed"
    printf "%s\n" "$resp_merge"
    exit 3
  fi
fi

echo "✅ PR merged"

# 3) Dispatch workflow
curl -sS "${AUTH_HEADER[@]}" -X POST \
  "${API}/repos/${REPO}/actions/workflows/autodeploy.yml/dispatches" \
  -d '{"ref":"main"}' >/dev/null || true

echo "▶️ Workflow dispatched, polling..."

# 4) Poll for completion (up to 10 minutes)
end=$((SECONDS+600))
run_result=""
run_url=""
status=""
while [[ $SECONDS -lt $end ]]; do
  runs=$(curl -sS "${AUTH_HEADER[@]}" \
    "${API}/repos/${REPO}/actions/workflows/autodeploy.yml/runs?per_page=1")
  status=$(printf "%s" "$runs" | sed -n 's/.*"status"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1 || true)
  run_result=$(printf "%s" "$runs" | sed -n 's/.*"conclusion"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1 || true)
  run_url=$(printf "%s" "$runs" | sed -n 's/.*"html_url"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1 || true)

  if [[ "${status}" == "completed" && -n "${run_result}" ]]; then
    break
  fi
  sleep 10

done

printf "RESULT: %s\n" "${run_result:-unknown}"
printf "URL:    %s\n" "${run_url:-}"

# 5) Optional Slack notify
if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
  payload=$(printf '%s' "{\"text\":\"Workflow result: ${run_result:-unknown}\\n🔗 ${run_url:-}\"}")
  curl -sS -X POST -H 'Content-type: application/json' --data "$payload" "$SLACK_WEBHOOK_URL" >/dev/null \
    && echo "📣 Slack notified"
fi

if [[ "${run_result}" == "success" ]]; then
  echo "🎉 Deployment succeeded on main"
else
  echo "⚠️ Merged but workflow failed; see link above"
fi

echo "🏁 Done"