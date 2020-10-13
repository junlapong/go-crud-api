default: build

build:
	go build -o api

dev:
	go run api.go
