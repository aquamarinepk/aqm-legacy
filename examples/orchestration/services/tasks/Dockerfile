# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder
WORKDIR /workspace

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd examples/orchestration && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /out/tasks ./services/tasks

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/tasks /app/tasks

ENV TASKS_HTTP_PORT=:8082
EXPOSE 8082
ENTRYPOINT ["/app/tasks"]
