# Expense Splitter
This project is an app allowing to split expenses inside a group.
Basically, it provides a frontend for ordering and managing project as well as services that perform the actions requested by the user. There are services for each resource type the API exposes and some additional services for the purpose of UX (e.g. gRPC reflection). The services only handle the requests by the users, i.e.
- reading requests collect the data the user requests and return the results
- writing requests only create a task on the message queue
- writing requests do not wait for tasks on the message queue to be finished but respond to the user immediately telling that a task to write was created with a related identifier
- tasks on the message queue are processed by dedicated MQ processing containers

# Prerequisites
- You are expected to have the go commandline tool installed
- You are expected to have Docker installed
- You are expected to have a K8s cluster up and running (for development purposes kind is recommended which can be installed using the respective make target)

# Installing
You can use `make skaffold-run` to run a pipeline that builds and deploys the Helm charts according to your configuration.

# Development
For dev purposes you can run `make skaffold-dev`. If you only want to develop specific charts but want to have some of the others installed it is recommended to first install charts you rely on using `make skaffold-run` in combination with environment variables set that skip those parts you want to develop later. Then, invert the values of the environment variables responsible for skipping charts so that only those charts will be included that you want to develop. Finally, run `make skaffold-dev` to actually start development.

## General notes
- Dockerfiles are expected to have the repository root as their context
- environment variables can be set by creating a .env file (a sample `.env.dist` file is provided)

## Adding a service
TODO: explain

## Adding a processor
TODO: explain

## Roadmap
Some elemental features that are intended to be implemented in the near future include:
- export protoc artefacts as libraries to registries in a pipeline
- host API documentation
- generate /cmd/*.Dockerfile.dockerignore files from root .dockerignore or find another way to extend the root .dockerignore (background for the current, weird structure: https://github.com/moby/moby/issues/12886)
- export OTEL to trace collector deployed with Jaeger
- cleanup frontend Dockerfile
- auth (probably via Ory Stack in combination with Stackgres for persistence)