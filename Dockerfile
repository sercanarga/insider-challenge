FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY . .

RUN go mod download && go mod verify
RUN CGO_ENABLED=0 go build -o service cmd/service/main.go

FROM gcr.io/distroless/static-debian12 AS runner
WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /app/service .
COPY --from=builder --chown=nonroot:nonroot /app/.env .

EXPOSE ${SERVER_PORT}

ENTRYPOINT ["./service"]