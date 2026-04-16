#!/usr/bin/env bash
# setup-emulator.sh — idempotently creates the Spanner emulator instance,
# database, and schema. Safe to run multiple times.
set -euo pipefail

SPANNER_PROJECT="${SPANNER_PROJECT:-cars-project}"
SPANNER_INSTANCE="${SPANNER_INSTANCE:-cars-instance}"
SPANNER_DATABASE="${SPANNER_DATABASE:-cars-db}"
EMULATOR_HOST="${SPANNER_EMULATOR_HOST:-localhost:9010}"

export SPANNER_EMULATOR_HOST="${EMULATOR_HOST}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATION="${SCRIPT_DIR}/../migrations/001_create_cars.sql"

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "ERROR: required command not found: $1" >&2
    exit 1
  fi
}

instance_exists() {
  gcloud spanner instances describe "${SPANNER_INSTANCE}" \
    --project="${SPANNER_PROJECT}" >/dev/null 2>&1
}

database_exists() {
  gcloud spanner databases describe "${SPANNER_DATABASE}" \
    --instance="${SPANNER_INSTANCE}" \
    --project="${SPANNER_PROJECT}" >/dev/null 2>&1
}

echo "==> Target emulator : ${EMULATOR_HOST}"
echo "==> Project         : ${SPANNER_PROJECT}"
echo "==> Instance        : ${SPANNER_INSTANCE}"
echo "==> Database        : ${SPANNER_DATABASE}"
echo ""

require_command gcloud
require_command nc

# ── Wait for emulator to be reachable ─────────────────────────────────────────
echo "--> Waiting for Spanner emulator..."
for i in $(seq 1 30); do
  if nc -z "${EMULATOR_HOST%%:*}" "${EMULATOR_HOST##*:}" 2>/dev/null; then
    echo "    Emulator is up."
    break
  fi
  echo "    Attempt ${i}/30 — retrying in 2s..."
  sleep 2
  if [ "${i}" -eq 30 ]; then
    echo "ERROR: Spanner emulator did not become available in time." >&2
    exit 1
  fi
done

# ── Create instance (idempotent) ───────────────────────────────────────────────
echo "--> Creating instance '${SPANNER_INSTANCE}' (skipping if exists)..."
if instance_exists; then
  echo "    Instance already exists — skipping."
else
  gcloud spanner instances create "${SPANNER_INSTANCE}" \
    --config=emulator-config \
    --description="Cars API Instance" \
    --nodes=1 \
    --project="${SPANNER_PROJECT}"
fi

# ── Create database (idempotent) ───────────────────────────────────────────────
echo "--> Creating database '${SPANNER_DATABASE}' (skipping if exists)..."
if database_exists; then
  echo "    Database already exists — skipping."
else
  gcloud spanner databases create "${SPANNER_DATABASE}" \
    --instance="${SPANNER_INSTANCE}" \
    --project="${SPANNER_PROJECT}"
fi

# ── Apply DDL ─────────────────────────────────────────────────────────────────
echo "--> Applying DDL from ${MIGRATION}..."
gcloud spanner databases ddl update "${SPANNER_DATABASE}" \
  --instance="${SPANNER_INSTANCE}" \
  --project="${SPANNER_PROJECT}" \
  --ddl-file="${MIGRATION}"

echo ""
echo "==> Emulator setup complete."
