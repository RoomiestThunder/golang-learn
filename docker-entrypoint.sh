#!/bin/sh
# При первом запуске кладёт в /data прогресс из репозитория, дальше он живёт только там.
# Порт берётся из PORT (его задаёт Cloud Run), по умолчанию 8088.
set -e
data="${GOLEARN_DATA:-/data/progress.json}"
if [ ! -f "$data" ] && [ -f /app/seed/progress.json ]; then
  cp /app/seed/progress.json "$data"
fi
exec /app/golearn -addr "0.0.0.0:${PORT:-8088}" -data "$data" "$@"
