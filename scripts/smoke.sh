#!/usr/bin/env sh
set -eu

base_url="${1:-http://127.0.0.1:${APP_PORT:-8088}}"
demo_username="${DEMO_USERNAME:-demo}"
demo_password="${DEMO_PASSWORD:-demo123}"

check_ok() {
  name="$1"
  url="$2"
  body="$(curl -fsS "$url")"
  printf '%s' "$body" | grep -q '"code":0'
  printf 'ok  %s\n' "$name"
}

check_ok "health" "$base_url/healthz"
check_ok "home" "$base_url/api/v1/home"
check_ok "movies" "$base_url/api/v1/movies?status=ON_SALE"

login_body="$(curl -fsS \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$demo_username\",\"password\":\"$demo_password\"}" \
  "$base_url/api/v1/auth/login")"
printf '%s' "$login_body" | grep -q '"code":0'
printf '%s' "$login_body" | grep -q '"token"'
printf 'ok  demo login\n'

printf 'smoke test passed: %s\n' "$base_url"
