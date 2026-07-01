# syntax=docker/dockerfile:1

# --- builder: toolchain yalnız burada yaşar ---
FROM golang:1.22 AS builder
WORKDIR /src
# Bağımlılık cache'i için önce modül dosyaları
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO'suz statik binary (cross-compile dostu)
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /voltium ./main.go

# --- final: küçük runtime imajı ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /voltium /usr/local/bin/voltium
EXPOSE 80
ENTRYPOINT ["voltium"]
