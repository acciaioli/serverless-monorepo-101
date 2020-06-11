.PHONY: default printenv compile dist

default: printenv

MK_ENV := $(ENV)
MK_BUCKET_NAME := $(DEPLOYMENT_BUCKET)-$(ENV)

MK_SERVICE := $(shell basename $(dir $(abspath $(firstword $(MAKEFILE_LIST)))))

printenv:
	@ echo ">> printing env..."
	@ echo "env: $(MK_ENV)"
	@ echo "bucket: $(MK_BUCKET_NAME)"
	@ echo "service-name: $(MK_SERVICE)"
	@ echo ">> done"

compile:
	@ echo ">> compiling binaries..."
	@ env GOOS=linux go build -ldflags="-s -w" -o .bin/echo echo/main.go
	@ echo ">> done"

dist:
	@ echo ">> running dist script..."
	@ ./dist.sh $(MK_BUCKET_NAME) $(MK_SERVICE)
	@ echo ">> done"
