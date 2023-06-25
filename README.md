# Expense Splitter
This project is an app allowing to split expenses inside a group.
Basically, it provides a frontend for ordering and managing project as well as services that perform the actions requested by the user. There are services for each resource type the API exposes and some additional services for the purpose of UX (e.g. gRPC reflection). The services only handle the requests by the users, i.e.
- reading requests collect the data the user requests and return the results
- writing requests only create a task on the message queue
- writing requests do not wait for tasks on the message queue to be finished but respond to the user immediately telling that a task to write was created with a related identifier
- tasks on the message queue are processed by dedicated MQ processing containers

## General notes
Dockerfiles are expected to have the repository root as their context

# Prerequisites
- You are expected to have the go commandline tool installed
- You are expected to have kubectl installed
- You are expected to have a K8s cluster up and running (for development purposes kind is recommended)

## Adding a service
TODO: explain

## Adding a processor
TODO: explain

## Prerequisites
You are required to have a Kubernetes cluster with an ingress controller (preferably Nginx) and a gRPC load-balancing solution like Linkerd for the gRPC pods installed. Then, you can use Skaffold to build and deploy the app.

## Roadmap
Some elemental features that are intended to be implemented in the near future include:
- export protoc artefacts as libraries to registries
- host API documentation
- generate /cmd/*.Dockerfile.dockerignore files from root .dockerignore or find another way to extend the root .dockerignore (background for the current, weird structure: https://github.com/moby/moby/issues/12886)
- export OTEL to trace collector deployed with Jaeger
- inject Linkerd sidecar in nats