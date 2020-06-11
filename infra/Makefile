.PHONY: default deploy

default: deploy

MK_ENV := $(ENV)
MK_BUCKET_NAME := $(DEPLOYMENT_BUCKET)-$(ENV)


deploy:
	@ echo ">> deploying stack..."
	@ aws cloudformation deploy \
		--region eu-west-1 \
		--template-file s3.yml \
		--stack-name serverless-monorepo-101--deployments-$(ENV) \
		--parameter-overrides \
			Env=$(MK_ENV) \
			BucketName=$(MK_BUCKET_NAME)
	@ echo ">> done"
