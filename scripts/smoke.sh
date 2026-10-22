#!/bin/sh
# Calls /ready on every service from inside the compose network and fails if any
# dependency check is not ok. Used by CI and `make smoke`.
set -eu

NETWORK="${NETWORK:-atlas_default}"
targets="gateway:8080 brokerage-sync:8000 auth:8000 portfolio:8000 analytics:8000 risk:8000 ai:8000 notifications:8000"

script='
fail=0
for t in '"$targets"'; do
  body=$(curl -s -m 5 -w " %{http_code}" "http://$t/ready") || body="unreachable 000"
  code=${body##* }
  printf "%-22s %s %s\n" "${t%%:*}" "$code" "${body% *}"
  [ "$code" = "200" ] || fail=1
done
exit $fail
'
MSYS_NO_PATHCONV=1 docker run --rm --network "$NETWORK" curlimages/curl:8.10.1 sh -c "$script"
