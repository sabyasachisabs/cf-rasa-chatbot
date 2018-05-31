# Simple HipChat Golang Integration

The integration is used to forward the message from Hipchat room to Rasa core running in iAPC.

## Running locally

- python3


### Requirements


### Train the bot

You can train the bot using below make target. You need to run it when you change the model.

```bash
make train
```

### Run the bot

You can run rasa core locally using below make target.

```bash
export LDAP_URL="ldap.corproot.net"
make run
```

Rasa core server will be available on port 8080.

## Deploy to iAPC

You can deploy the bot to iAPC using make script

### Requirements

- make

### Steps for build and deploy

```bash
make cf-login
make deploy
```

You need to set below variables for authenticating against iAPC

```bash
export CF_USER=<?>
export CF_PASS=<?>
export ENVIRONMENT=<Space_Name>
export CF_ORG=<?>
export HIPCHAT_AUTH_TOKEN=<?>
export LDAP_URL=<?>
export LDAP_USER=<?>
export LDAP_PASS=<?>
```