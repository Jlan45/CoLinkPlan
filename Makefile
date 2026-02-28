.PHONY: all build build-web build-server build-client clean

all: build

build: build-web build-server build-client

build-web:
	@echo "Building React frontend..."
	cd web && npm i && npm run build

build-server:
	@echo "Building Go server with embedded frontend..."
	go build -o bin/server ./cmd/server

build-client:
	@echo "Building Go client..."
	go build -o bin/client ./cmd/client

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/dist/
