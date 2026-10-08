# Host'a Go kurulmaz — tüm toolchain golang:1.22 container'ında koşar.
IMAGE := voltium:dev
GO := docker run --rm -v "$(PWD)":/src -w /src golang:1.22

.PHONY: build test lint tidy run stop clean

## build: statik binary'yi imaja derle
build:
	docker build -t $(IMAGE) .

## test: tüm testleri container'da koştur
test:
	$(GO) go test ./...

## lint: go vet
lint:
	$(GO) go vet ./...

## tidy: go.mod/go.sum'ı güncelle (go.sum'ı ilk kez üretir)
tidy:
	$(GO) go mod tidy

## run: dev ortamını ayağa kaldır (voltium + test upstream)
run:
	docker compose up --build

## stop: compose servislerini durdur
stop:
	docker compose down

## clean: compose + volume temizliği
clean:
	docker compose down -v
