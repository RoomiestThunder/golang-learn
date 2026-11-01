#!/usr/bin/env bash
# Выкладывает Go Learn в Google Cloud Run. Запускать в Cloud Shell (или там, где настроен gcloud)
# из корня репозитория:   ./deploy/cloudrun.sh
#
# Что делает: включает API, создаёт бакет для прогресса, пароль в Secret Manager и отдельный
# сервис-аккаунт с минимальными правами, затем собирает образ (Cloud Build) и выкладывает сервис.
# Повторный запуск безопасен — это обновление; пароль и прогресс остаются.
set -euo pipefail
cd "$(dirname "$0")/.."

PROJECT="${PROJECT:-$(gcloud config get-value project 2>/dev/null)}"
REGION="${REGION:-europe-west1}"
SERVICE="${SERVICE:-golearn}"
BUCKET="${BUCKET:-${PROJECT}-golearn-data}"
SECRET="${SECRET:-golearn-password}"
SA_NAME="golearn-run"
SA="${SA_NAME}@${PROJECT}.iam.gserviceaccount.com"

[ -n "$PROJECT" ] || { echo "Проект не выбран: gcloud config set project <ID проекта>"; exit 1; }
echo "Проект: $PROJECT, регион: $REGION, сервис: $SERVICE"

echo "→ включаю API (первый раз может занять минуту)…"
gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com \
  secretmanager.googleapis.com storage.googleapis.com iam.googleapis.com --project "$PROJECT"

# Код из заданий исполняется внутри сервиса и может читать токен его сервис-аккаунта,
# поэтому используем отдельный аккаунт без прав на остальной проект.
if ! gcloud iam service-accounts describe "$SA" --project "$PROJECT" >/dev/null 2>&1; then
  echo "→ создаю сервис-аккаунт $SA"
  gcloud iam service-accounts create "$SA_NAME" --display-name "Go Learn runtime" --project "$PROJECT"
fi

if ! gcloud storage buckets describe "gs://$BUCKET" --project "$PROJECT" >/dev/null 2>&1; then
  echo "→ создаю бакет gs://$BUCKET для прогресса"
  gcloud storage buckets create "gs://$BUCKET" --location "$REGION" --uniform-bucket-level-access --project "$PROJECT"
fi
gcloud storage buckets add-iam-policy-binding "gs://$BUCKET" \
  --member "serviceAccount:$SA" --role roles/storage.objectUser --project "$PROJECT" >/dev/null

if ! gcloud secrets describe "$SECRET" --project "$PROJECT" >/dev/null 2>&1; then
  PASSWORD="${GOLEARN_PASSWORD:-$(openssl rand -hex 16)}"
  [ "${#PASSWORD}" -ge 12 ] || { echo "Пароль должен быть не короче 12 символов"; exit 1; }
  printf %s "$PASSWORD" | gcloud secrets create "$SECRET" --data-file=- --replication-policy=automatic --project "$PROJECT" >/dev/null
  echo
  echo "  ПАРОЛЬ ДЛЯ ВХОДА (показан один раз, сохрани в менеджер паролей):"
  echo "  $PASSWORD"
  echo
fi
gcloud secrets add-iam-policy-binding "$SECRET" \
  --member "serviceAccount:$SA" --role roles/secretmanager.secretAccessor --project "$PROJECT" >/dev/null

# max-instances=1 обязателен: прогресс держится в памяти и пишется в один файл,
# два экземпляра затирали бы друг другу записи.
echo "→ собираю образ и выкладываю (несколько минут)…"
gcloud run deploy "$SERVICE" --source . --project "$PROJECT" --region "$REGION" --quiet \
  --service-account "$SA" \
  --allow-unauthenticated \
  --max-instances 1 --concurrency 8 \
  --cpu 2 --memory 2Gi --cpu-boost --timeout 300 \
  --set-secrets "GOLEARN_PASSWORD=${SECRET}:latest" \
  --add-volume "name=data,type=cloud-storage,bucket=${BUCKET},mount-options=uid=10001;gid=10001;file-mode=644;dir-mode=755" \
  --add-volume-mount "volume=data,mount-path=/data"

echo
echo "Готово. Адрес платформы:"
gcloud run services describe "$SERVICE" --project "$PROJECT" --region "$REGION" --format 'value(status.url)'
echo "Пароль, если забыл:  gcloud secrets versions access latest --secret $SECRET"
