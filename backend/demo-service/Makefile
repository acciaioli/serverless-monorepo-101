.PHONY: default test compile

default: test

test:
	@ echo ">> running tests..."
	@ go test --count=1 ./...
	@ echo ">> done"

compile:
	@ echo ">> compiling binaries..."
	@ go build -ldflags="-s -w" -o .bin/echo echo/main.go
	@ echo ">> done"
