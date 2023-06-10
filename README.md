# Expense Splitter
This project is an app allowing to split expenses inside a group.
Basically, it provides a frontend for ordering and managing project as well as services that perform the actions requested by the user. There are services for each resource type the API exposes and some additional services for the purpose of UX (e.g. gRPC reflection). The services only handle the requests by the users, i.e.
- reading requests collect the data the user requests and return the results
- writing requests only create a task on the message queue
- writing requests do not wait for tasks on the message queue to be finished but respond to the user immediately telling that a task to write was created with a related identifier
- tasks on the message queue are processed by dedicated MQ processing containers

## General notes
Dockerfiles are expected to have the repository root as their context

## Adding a service
TODO

## Adding a processor
TODO

## Prerequisites
You are required to have a Kubernetes cluster with an ingress controller (preferably Nginx) and a gRPC load-balancing solution like Linkerd for the gRPC pods installed. Then, you can use Skaffold to build and deploy the app.
