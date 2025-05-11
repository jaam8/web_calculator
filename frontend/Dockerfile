# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS builder
LABEL authors="jaam8"

WORKDIR /build

COPY frontend/go.* ./frontend/

#RUN --mount=type=cache,target=/go/pkg/mod \
#    --mount=type=cache,target=/root/.cache/go-build \
#    go mod download

COPY frontend/ ./frontend/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o frontend-service ./frontend/main.go

FROM gcr.io/distroless/base-debian12 AS runner

WORKDIR /app

COPY --from=builder /build/frontend/static ../static
COPY --from=builder /build/frontend-service .

CMD ["./frontend-service"]