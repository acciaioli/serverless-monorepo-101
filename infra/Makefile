.PHONY: default compile deploy

default: compile

MK_ENV := $(ENV)
MK_BUCKET_NAME := $(DEPLOYMENT_BUCKET)


BIN ?= .bin/main

compile-ci-hash:
	@ echo ">> compiling ci-hash..."
	@ go build -o $(BIN) internal/cmds-ci/hash/main.go
	@ echo ">> done"

compile-ci-build:
	@ echo ">> compiling ci-hash..."
	@ go build -o $(BIN) internal/cmds-ci/build/main.go
	@ echo ">> done"

compile-ci-deploy:
	@ echo ">> compiling ci-hash..."
	@ go build -o $(BIN) internal/cmds-ci/deploy/main.go
	@ echo ">> done"

compile-user-deploy:
	@ echo ">> compiling ci-hash..."
	@ go build -o $(BIN) internal/cmds-user/deploy/main.go
	@ echo ">> done"


test-compile: compile-ci-hash compile-ci-build compile-ci-deploy compile-user-deploy
	@ rm -rf $(BIN)

deploy:
	@ echo ">> deploying stack..."
	@ aws cloudformation deploy \
		--region eu-west-1 \
		--template-file s3.yml \
		--stack-name serverless-monorepo-101--deployments\
		--parameter-overrides \
			BucketName=$(MK_BUCKET_NAME)
	@ echo ">> done"
