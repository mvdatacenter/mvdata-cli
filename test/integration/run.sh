#!/usr/bin/env bash
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$here/../.." && pwd)"
ref="$(tr -d '[:space:]' <"$here/console-api.ref")"
src="$here/console-api"
# A project name per run keeps two checkouts sharing one Docker daemon out of each other's way.
compose=(docker compose -f "$here/docker-compose.yml" -p "mvdata-cli-integration-$$")

# Hooks are off: the clone is a build input, not a checkout anyone works in.
if [ ! -d "$src/.git" ]; then
  git -c core.hooksPath=/dev/null clone --quiet https://github.com/mvdatacenter/console-api.git "$src"
fi
if [ "$(git -C "$src" rev-parse HEAD)" != "$ref" ]; then
  git -C "$src" -c core.hooksPath=/dev/null fetch --quiet origin "$ref"
  git -C "$src" -c core.hooksPath=/dev/null checkout --quiet --detach "$ref"
fi

# The seed stores only the key's hash, so nothing outside this run ever holds the key itself.
ACCEPTANCE_API_KEY="mvd_$(head -c 30 /dev/urandom | base64 | tr -dc 'A-Za-z0-9' | head -c 40)"
export ACCEPTANCE_API_KEY
# pyx-setup.js mints the reconciler's and the tenant administrator's OpenBao tokens under these ids.
PYX_RECONCILER_TOKEN="acc-$(head -c 30 /dev/urandom | base64 | tr -dc 'A-Za-z0-9' | head -c 40)"
PYX_TENANT_ADMIN_TOKEN="acc-$(head -c 30 /dev/urandom | base64 | tr -dc 'A-Za-z0-9' | head -c 40)"
export PYX_RECONCILER_TOKEN PYX_TENANT_ADMIN_TOKEN

work="$(mktemp -d)"
cleanup() {
  local project="mvdata-cli-integration-$$" out
  if ! out="$("${compose[@]}" down --volumes --remove-orphans 2>&1)"; then
    echo "teardown of compose project $project failed: $out" >&2
    echo "remove what is left with: docker compose -p $project down --volumes" >&2
  elif ! out="$("${compose[@]}" ps --all --quiet 2>&1)"; then
    echo "could not confirm compose project $project is gone: $out" >&2
    echo "check what is left with: docker compose -p $project ps --all" >&2
  elif [ -n "$out" ]; then
    echo "containers of compose project $project survived teardown;" \
      "remove them with: docker compose -p $project down --volumes" >&2
  fi
  rm -rf "$work" || echo "could not remove $work" >&2
}
trap cleanup EXIT

(cd "$repo" && go build -o "$work/mvdata" .)

# Three steps, because the reconciler and the tenant administrator need tokens that only exist once
# pyx-setup.js has run against OpenBao, and it runs inside the console's container.
if ! "${compose[@]}" up --build --detach --wait db openbao audit_bucket console; then
  "${compose[@]}" logs console >&2 || true
  exit 1
fi
# Without token ids pyx-setup.js must refuse before touching OpenBao; the console's own environment
# carries none.
if out="$("${compose[@]}" exec -T console node - <"$here/pyx-setup.js" 2>&1)" ||
  ! grep -q "no token id for pyx-reconciler" <<<"$out"; then
  echo "pyx-setup.js did not refuse a run without token ids: $out" >&2
  exit 1
fi
if ! "${compose[@]}" exec -T -e PYX_RECONCILER_TOKEN -e PYX_TENANT_ADMIN_TOKEN console node - <"$here/pyx-setup.js"; then
  "${compose[@]}" logs openbao audit_bucket >&2 || true
  exit 1
fi
if ! "${compose[@]}" up --build --detach --wait reconciler tenant_admin; then
  "${compose[@]}" logs reconciler tenant_admin >&2 || true
  exit 1
fi

published="$("${compose[@]}" port console 7734)"
url="http://localhost:${published##*:}"
echo "console-api $ref serving on $url"

# The CLI reads ~/.mvdata/config.json, so the run gets a home of its own and the operator's
# profile never answers for the console.
export HOME="$work/home" MVDATA_API_URL="$url" MVDATA_API_TOKEN="$ACCEPTANCE_API_KEY" MVDATA_PROFILE=
mkdir -p "$HOME"

failures=0
# expect <status> <args...> runs mvdata and records a failure unless it exits with <status>.
expect() {
  local want="$1" got=0
  shift
  "$work/mvdata" "$@" >"$work/out" 2>&1 || got=$?
  if [ "$got" -eq "$want" ]; then
    echo "ok   mvdata $* exited $got"
  else
    echo "FAIL mvdata $* exited $got, want $want:" >&2
    cat "$work/out" >&2
    failures=$((failures + 1))
  fi
}

expect 0 secret-store list
expect 3 secret-store get integration-missing
expect 3 secret list --store integration-missing
expect 3 secret get integration-missing/absent

if [ -e "$HOME/.mvdata/config.json" ]; then
  echo "FAIL the run wrote $HOME/.mvdata/config.json:" >&2
  cat "$HOME/.mvdata/config.json" >&2
  failures=$((failures + 1))
else
  echo "ok   no configuration file was written"
fi

if [ "$failures" -ne 0 ]; then
  "${compose[@]}" logs console >&2 || true
  echo "$failures check(s) failed" >&2
  exit 1
fi
