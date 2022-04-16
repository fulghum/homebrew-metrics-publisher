
Configuration
===
* Make sure DoltHub API token has been generated and stored in SSM Parameter store under "dolthub-auth-token" in the configured AWS account.
* Install CDK with `npm install -g aws-cdk`
* TODO: To run Go code locally, make sure AWS_PROFILE env is set

Deploying
===
* Make sure infra code has been compiled; `npm run watch` helps
* Run `cdk deploy`
