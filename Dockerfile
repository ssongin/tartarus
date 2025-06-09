# Stage 1: Build
FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go build -o app .

# Stage 2: Minimal runtime
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/app .
ENTRYPOINT ["./app"]
