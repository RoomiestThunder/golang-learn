# Хостинг Go Learn. Платформе нужен компилятор Go в рантайме (она собирает и тестирует код ученика),
# поэтому образ — полный golang, а не минимальный.
FROM golang:1.25-bookworm

RUN useradd --create-home --uid 10001 learner \
 && mkdir /data && chown learner /data
WORKDIR /app
COPY --chown=learner . .
USER learner

# Платформа сама будет вызывать go build/test с CGO_ENABLED=0 — прогреваем кэш стандартной
# библиотеки, чтобы первая проверка не занимала минуты.
RUN CGO_ENABLED=0 go build std && go build -o /home/learner/golearn .

# Прогресс лежит в /data: подключи к нему постоянный диск (volume), иначе он сбросится при перезапуске.
# Обязательно задай GOLEARN_PASSWORD (от 12 символов) — без него сервер не запустится вне localhost.
ENV GOLEARN_DATA=/data/progress.json
VOLUME /data
EXPOSE 8088
ENV PORT=8088
CMD ["/home/learner/golearn"]
