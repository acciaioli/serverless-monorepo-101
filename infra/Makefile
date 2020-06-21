.PHONY: default compile deploy

default: compile

BIN ?= .bin/main

compile-ci-hash:
	@ echo ">> compiling ci-hash...  ($(BIN))"
	@ go build -o $(BIN) internal/cmds-ci/hash/main.go
	@ echo ">> done"

compile-ci-build:
	@ echo ">> compiling ci-build...  ($(BIN))"
	@ go build -o $(BIN) internal/cmds-ci/build/main.go
	@ echo ">> done"

compile-ci-deploy:
	@ echo ">> compiling ci-deploy...  ($(BIN))"
	@ go build -o $(BIN) internal/cmds-ci/deploy/main.go
	@ echo ">> done"

compile-user-deploy:
	@ echo ">> compiling user-deploy...  ($(BIN))"
	@ go build -o $(BIN) internal/cmds-user/deploy/main.go
	@ echo ">> done"

test-compile: compile-ci-hash compile-ci-build compile-ci-deploy compile-user-deploy
	@ echo ">> cleaning up..."
	@ rm -rf $(BIN)
	@ echo ">> done"

MK_DEPLOYMENT_BUCKET := $(INFRA_AWS_S3_BUCKET)

deploy:
	@ echo ">> deploying stack..."
	@ aws cloudformation deploy \
		--region eu-west-1 \
		--template-file s3.yml \
		--stack-name serverless-monorepo-101--deployments \
		--no-fail-on-empty-changeset \
		--parameter-overrides \
			BucketName=$(MK_DEPLOYMENT_BUCKET)
	@ echo ">> done"
