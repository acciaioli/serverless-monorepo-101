# SERVERLESS-MONOREPO-101 INFRA

#### Notes

##### github actions list of permissions

For serverless deploy:
```
"cloudformation:CreateStack",
"cloudformation:DescribeStacks",
"cloudformation:DescribeStackResources",
"cloudformation:UpdateStack",
"cloudformation:ListStacks",
"iam:GetRole",
"lambda:UpdateFunctionCode",
"lambda:UpdateFunctionConfig",
"lambda:GetFunctionConfiguration",
"lambda:AddPermission",
"s3:DeleteObject",
"s3:GetObject",
"s3:ListBucket",
"s3:PutObject"
```
