# ---- сборка платформы (контент и веб-интерфейс вшиваются в бинарь через go:embed) ----
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY *.go login.html ./
COPY web ./web
COPY content ./content
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /golearn .

# ---- рантайм: платформа запускает код ученика, поэтому внутри нужен компилятор Go ----
FROM golang:1.25-bookworm
RUN useradd --system --uid 10001 --create-home --home-dir /home/app app \
 && mkdir -p /app/seed /data /cache \
 && chown app:app /data /cache
ENV GOCACHE=/cache/build GOPATH=/cache/go GOTOOLCHAIN=local

COPY --from=build /golearn /app/golearn
COPY progress.json /app/seed/progress.json
COPY docker-entrypoint.sh warm-cache.sh /app/
RUN chmod +x /app/docker-entrypoint.sh /app/warm-cache.sh

USER app

# Кэш сборки со стандартной библиотекой: первая проверка задания не компилирует stdlib с нуля.
# При первом создании тома /cache Docker копирует этот кэш в том.
RUN /app/warm-cache.sh

WORKDIR /app
EXPOSE 8088
ENTRYPOINT ["/app/docker-entrypoint.sh"]
