#!/bin/sh
# Прогревает кэш сборки Go стандартной библиотекой (с теми же настройками, что у runner.go),
# чтобы первая проверка задания не компилировала stdlib с нуля. Запускается при сборке образа.
set -e
d=$(mktemp -d)
cd "$d"
printf 'module play\n\ngo 1.25\n' > go.mod
pkgs="bufio bytes cmp container/heap container/list context crypto/hmac crypto/rand crypto/sha256 crypto/subtle database/sql encoding/base64 encoding/csv encoding/hex encoding/json errors fmt hash/fnv io io/fs iter log log/slog maps math math/rand/v2 net net/http net/http/httptest net/url os os/signal path/filepath reflect regexp runtime slices sort strconv strings sync sync/atomic syscall testing testing/iotest time unicode unicode/utf8"
{
  echo 'package main'
  echo 'import ('
  for p in $pkgs; do printf '\t_ "%s"\n' "$p"; done
  echo ')'
  echo 'func main() {}'
} > main.go
printf 'package main\n\nimport "testing"\n\nfunc TestWarm(t *testing.T) {}\n' > warm_test.go
export CGO_ENABLED=0 GOPROXY=off GOFLAGS=
go build -o prog .
go test -count=1 .
cd /
rm -rf "$d"
