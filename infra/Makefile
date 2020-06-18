.PHONY: default deploy

default: deploy

MK_ENV := $(ENV)
MK_BUCKET_NAME := $(DEPLOYMENT_BUCKET)


deploy:
	@ echo ">> deploying stack..."
	@ aws cloudformation deploy \
		--region eu-west-1 \
		--template-file s3.yml \
		--stack-name serverless-monorepo-101--deployments\
		--parameter-overrides \
			BucketName=$(MK_BUCKET_NAME)
	@ echo ">> done"
