FROM golang:1.22 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/proxy ./cmd/app

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/proxy /app/proxy
COPY config.yaml /app/config.yaml

ENV CONFIG_PATH=/app/config.yaml

EXPOSE 8090

ENTRYPOINT ["/app/proxy"]
