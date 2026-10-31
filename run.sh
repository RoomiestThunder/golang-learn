#!/usr/bin/env bash
# Запуск платформы с синхронизацией прогресса через git:
#   1) подтягивает свежий progress.json с других устройств,
#   2) запускает платформу,
#   3) после остановки (Ctrl+C) коммитит и отправляет прогресс.
set -u
cd "$(dirname "$0")"

echo "↓ git pull (подтягиваю прогресс с других устройств)…"
git pull --rebase --autostash || echo "⚠ pull не удался — продолжаю локально (проверь конфликты в progress.json)"

go run . "$@"

echo
echo "↑ сохраняю прогресс в git…"
git add progress.json
if git diff --cached --quiet; then
  echo "прогресс не изменился"
else
  git commit -m "progress: $(date '+%Y-%m-%d %H:%M')" -q
  git pull --rebase --autostash -q || echo "⚠ конфликт при pull — реши его вручную (см. README)"
  git push -q && echo "✓ прогресс отправлен" || echo "⚠ push не удался — повтори: git push"
fi
