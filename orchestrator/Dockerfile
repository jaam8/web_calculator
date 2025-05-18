# syntax=docker/dockerfile:1.4
FROM golang:alpine AS builder
LABEL authors="jaam8"

WORKDIR /build

COPY orchestrator/go.* ./orchestrator/
COPY common-lib/go.* ./common-lib/

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go work init ./common-lib ./orchestrator && go mod download

COPY orchestrator/ ./orchestrator/
COPY db/migrations ./db/migrations/
COPY common-lib/ ./common-lib/
COPY .env/ ./.env

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o orchestrator-service orchestrator/cmd/main.go

FROM gcr.io/distroless/base-debian12 AS runner

WORKDIR /app

COPY --from=builder /build/.env ../.env
COPY --from=builder /build/db/migrations ../db/migrations
COPY --from=builder /build/orchestrator-service .

CMD ["./orchestrator-service"]