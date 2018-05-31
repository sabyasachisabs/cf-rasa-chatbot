# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

.PHONY: check-env-deploy build package deploy
.DEFAULT_GOAL := help

build: ## setup env and build
	@docker build -f Dockerfile_build -t dos-docker.artifactory.swisscom.com/go-builder:latest .
	@mkdir -p build
	@docker run --rm=true -w /usr/local/go/src/nlp-hipchat-integration/ -v `pwd`:/usr/local/go/src/nlp-hipchat-integration/ dos-docker.artifactory.swisscom.com/go-builder:latest env GOOS=linux GOARCH=amd64 go build -o build/nlp-hipchat-integration .

package: ## package the built library for deployment
	@echo "Package the library to deploy to cf"
	@rm -rf deploy_cf && mkdir -p deploy_cf
	@cp build/nlp-hipchat-integration deploy_cf/nlp-hipchat-integration
	@cp -r build/* deploy_cf/
	@cp -r static deploy_cf/
	@echo 'web: ./nlp-hipchat-integration' > deploy_cf/Procfile
	@cd deploy_cf && zip -r nlp-hipchat-integration.zip .
	@mv deploy_cf/nlp-hipchat-integration.zip .

cf-login: ## login to cloud foundry
	echo "Deploy the app hipchat integration to App Cloud"
	echo "Login to internal app cloud..."
	@cf login -a https://api.scapp-console.swisscom.com -u ${CF_USER} -p ${CF_PASS}
	@cf target -o ${CF_ORG} -s ${ENVIRONMENT}
	echo "List currently running apps"
	@cf apps

scapp_appname=$(shell echo nlp-hipchat-integration)

deploy: check-env-deploy ## deploy the binary to AppCloud
	@echo "App Name is: ${scapp_appname}"

	# Push new app:
	@cf push ${scapp_appname}-${ENVIRONMENT} -n ${scapp_appname}-${ENVIRONMENT} -m 64M -i 1 --no-start -p ${scapp_appname}.zip -b binary_buildpack
	@cf set-env ${scapp_appname}-${ENVIRONMENT} BASE_URL https://${scapp_appname}-${ENVIRONMENT}.scapp-corp.swisscom.com
	@cf set-env ${scapp_appname}-${ENVIRONMENT} CHATBOT_URL ${CHATBOT_URL}
	@cf start ${scapp_appname}-${ENVIRONMENT}

	# Map traffic to new app too:
	@cf map-route ${scapp_appname}-${ENVIRONMENT} scapp-corp.swisscom.com -n ${scapp_appname}-${ENVIRONMENT}
	@cf routes

	@rm ${scapp_appname}.zip

check-env-deploy:
ifndef ENVIRONMENT
	$(error ENVIRONMENT is not set)
else
	$(info ENVIRONMENT is set to ${ENVIRONMENT})
endif

ifndef CF_USER
	$(error CF_USER is not set)
else
	$(info CF_USER is set to ${CF_USER})
endif

ifndef CF_PASS
	$(error CF_PASS is not set)
else
	$(info CF_PASS is set)
endif

ifndef CF_ORG
	$(error CF_ORG is not set)
else
	$(info CF_ORG is set to ${CF_ORG})
endif

ifndef CHATBOT_URL
	$(error CHATBOT_URL is not set)
else
	$(info CHATBOT_URL is set to ${CHATBOT_URL})
endif

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
