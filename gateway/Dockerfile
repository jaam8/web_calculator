# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS builder
LABEL authors="jaam8"

WORKDIR /build

COPY gateway/go.* ./gateway/
COPY common-lib/go.* ./common-lib/

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go work init ./common-lib ./gateway && go mod download

COPY gateway/ ./gateway/
COPY common-lib/ ./common-lib/
COPY .env/ ./.env

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o gateway-service ./gateway/cmd/main.go

FROM gcr.io/distroless/base-debian12 AS runner

WORKDIR /app

COPY --from=builder /build/.env ../.env
COPY --from=builder /build/gateway-service .

CMD ["./gateway-service"]