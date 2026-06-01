FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY bernard-protos ./bernard-protos
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o task-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/task-service .
COPY config.yaml .

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8082
CMD ["./task-service"]