.PHONY: default test build upload update-stack deploy

default: test

MK_ENV=$(ENV)
MK_BUCKET_NAME=$(BUCKET_NAME)-$(ENV)

test:
	@ echo ">> running tests..."
	@ npm run lint
	@ echo ">> done"

build:
	@ echo ">> building..."
	@ npm run build
	@ echo ">> done"

upload:
	@ echo ">> uploading..."
	@ aws s3 sync dist s3://$(MK_BUCKET_NAME)
	@ echo ">> done"

update-stack:
	@ echo ">> updating stack..."
	@ aws cloudformation deploy \
		--region eu-west-1 \
		--template-file template.yml \
		--stack-name serverless-monorepo-101--frontend \
		--parameter-overrides \
			Env=$(MK_ENV) \
			BucketName=$(MK_BUCKET_NAME)
	@ echo ">> done"


deploy: build upload update-stack
