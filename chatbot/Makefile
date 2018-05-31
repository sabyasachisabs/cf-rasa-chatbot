# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

.PHONY: check-env-deploy build package deploy
.DEFAULT_GOAL := help

train-bot:
	@echo "Training the bot with data.."
	export LC_ALL=en_US.UTF-8
	export LANG=en_US.UTF-8
	pip3 install -r requirements.txt
	python3 -m rasa_nlu.train --config config_spacy.yml --data data/chatbot_training.json --path models
	python3 -m rasa_nlu.train --config config_spacy.yml --data data/chatbot_training.json --path models --project current --fixed_model_name nlu
	python3 -m rasa_core.train -d domain.yml -s data/stories.md -o models/dialogue --epochs 300

run: ## run the bot locally
	@echo "Starting the rasa core server locally it will be accessible on port 8080"
	export LC_ALL=en_US.UTF-8
	export LANG=en_US.UTF-8
	pip3 install -r requirements.txt
	python3 run.py

cf-login: train-bot ## login to cloud foundry
	@echo "Deploy the app rasa core to App Cloud"
	@echo "Login to internal app cloud..."
	@cf login -a https://api.scapp-console.swisscom.com -u ${CF_USER} -p ${CF_PASS}
	@cf target -o ${CF_ORG} -s ${ENVIRONMENT}
	echo "List currently running apps"
	@cf apps

scapp_appname=$(shell echo chatbot)

deploy: check-env-deploy  ## deploy the binary to AppCloud
	@echo "App Name is: ${scapp_appname}"

	# Push new app:
	@cf push ${scapp_appname}-${ENVIRONMENT} -n ${scapp_appname}-${ENVIRONMENT} -m 4G -i 1 --no-start -p ./ -b python_buildpack
	#cf set-env ${scapp_appname} HIPCHAT_AUTH_TOKEN ""
	#cf set-env ${scapp_appname} LDAP_USER ${LDAP_USER}
	#cf set-env ${scapp_appname} LDAP_PASS ""
	#cf set-env ${scapp_appname} LDAP_URL ${LDAP_URL}

	@cf start ${scapp_appname}-${ENVIRONMENT}

	# Map traffic to new app too:
	@cf map-route ${scapp_appname}-${ENVIRONMENT} scapp-corp.swisscom.com -n ${scapp_appname}-${ENVIRONMENT}

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

ifndef HIPCHAT_AUTH_TOKEN
	$(error HIPCHAT_AUTH_TOKEN is not set)
else
	$(info HIPCHAT_AUTH_TOKEN is set)
endif

ifndef LDAP_USER
	$(error LDAP_USER is not set)
else
	$(info LDAP_USER is set to ${LDAP_USER})
endif

ifndef LDAP_PASS
	$(error LDAP_PASS is not set)
else
	$(info LDAP_PASS is set)
endif

ifndef LDAP_URL
	$(error LDAP_URL is not set)
else
	$(info LDAP_URL is set to ${LDAP_URL})
endif

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'