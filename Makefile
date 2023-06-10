REPO:=github.com/nico151999/high-availability-expense-splitter
SKIP_CERT_GENERATION:=false # if set to true it expects the Linkerd cert to exist already; the old one will not be deleted and no new certs will be created
BUF_VERSION:=1.17.0
GOMPLATE_VERSION:=3.11.5
GOLANGCI_VERSION:=1.49.0
CHECK_BREAKING_CHANGES:=true
DOCUMENTATION_PATH:=cmd/service/documentation
REFLECTION_SVC_PATH:=cmd/service/reflection
GROUP_SVC_PATH:=cmd/service/group
GROUP_PROCESSOR_PATH:=cmd/processor/group
OUT_DIR:=./gen
LIB_OUT_DIR:=$(OUT_DIR)/lib
APPLICATION_OUT_DIR:=$(OUT_DIR)/application
GO_LIB_OUT_DIR:=$(LIB_OUT_DIR)/go
DOCUMENTATION_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(DOCUMENTATION_PATH)
REFLECTION_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(REFLECTION_SVC_PATH)
GROUP_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(GROUP_SVC_PATH)
INGRESS_URL_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(INGRESS_URL_SVC_PATH)
GROUP_PROCESSOR_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(GROUP_PROCESSOR_PATH)

# source the .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# install the buf commandline tool required to run buf commands
.PHONY: install-buf
install-buf:
# TODO: also check if correct version is installed
ifeq (, $(shell which buf))
	go install github.com/bufbuild/buf/cmd/buf@v$(BUF_VERSION)
endif

# install the buf commandline tool required to run buf commands
.PHONY: install-kubeconform
install-kubeconform:
ifeq ($(findstring kubeconform,$(shell helm plugin list)),)
	helm plugin install https://github.com/jtyr/kubeconform-helm
endif

# install golang-ci llinter
.PHONY: install-golangci-lint
install-golangci-lint:
# TODO: also check if correct version is installed
ifeq (, $(shell which golangci-lint))
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_VERSION)
endif

# updates protobuf dependencies
.PHONY: update-buf-dependencies
update-buf-dependencies: install-buf
	buf mod update proto

# format code; should be run before something is merged into main branch; consequently this should be part of our pipeline
.PHONY: format
format: install-buf
	buf format -w
	go fmt ./...

# generate files from proto definitions using buf
.PHONY: generate-proto
generate-proto: install-buf clean
	buf generate

# generate Dockerfile links required for using namespaced dockerignore files: https://github.com/moby/moby/issues/12886#issuecomment-480575928
.PHONY: generate-dockerfile-links
generate-dockerfile-links:
	ln -sf Dockerfile ./cmd/service/documentation.Dockerfile
	ln -sf Dockerfile ./cmd/service/group.Dockerfile
	ln -sf Dockerfile ./cmd/service/reflection.Dockerfile
	ln -sf Dockerfile ./cmd/processor/group.Dockerfile

# performs all code generation tasks
.PHONY: generate
generate: generate-proto generate-manifests

# initializes the ts workspace of the generated proto files
.PHONY: prepare-ts-proto
prepare-ts-proto: generate-proto
	cd $(LIB_OUT_DIR)/ts && pnpm init && pnpm install @bufbuild/protobuf @bufbuild/protoc-gen-es

# lint code; should be run before something is merged into master branch; consequently this should be part of our pipeline
.PHONY: lint
lint: install-buf install-golangci-lint generate-proto prepare-ts-proto install-kubeconform
	buf lint
ifeq ($(CHECK_BREAKING_CHANGES),true)
	buf breaking --against '.git#branch=main'
endif
	go vet ./...
	golangci-lint run
	helm kubeconform --verbose --summary ./charts/ha-expense-splitter

.PHONY: clean-lib
clean-lib:
	rm -rf $(LIB_OUT_DIR)

.PHONY: clean-application
clean-application:
	rm -rf $(APPLICATION_OUT_DIR)

# cleanup generated files
.PHONY: clean
clean: clean-lib clean-application

test: generate format lint prepare-ts-proto generate-proto
	go test ./... -coverprofile cover.out
# TODO: also test TS

# build documentation UI
.PHONY: build-documentation
build-documentation: generate-proto
	CGO_ENABLED=0 go build -o $(DOCUMENTATION_OUT_DIR) $(REPO)/$(DOCUMENTATION_PATH)

# builds group service
.PHONY: build-group-service
build-group-service: generate-proto
	CGO_ENABLED=0 go build -o $(GROUP_SVC_OUT_DIR) $(REPO)/$(GROUP_SVC_PATH)

# builds reflection service
.PHONY: build-reflection-service
build-reflection-service: generate-proto
	CGO_ENABLED=0 go build -o $(REFLECTION_SVC_OUT_DIR) $(REPO)/$(REFLECTION_SVC_PATH)

# builds group processor
.PHONY: build-group-processor
build-group-processor: generate-proto
	CGO_ENABLED=0 go build -o $(GROUP_PROCESSOR_OUT_DIR) $(REPO)/$(GROUP_PROCESSOR_PATH)

# starts the dev mode of skaffold
.PHONY: skaffold-dev
skaffold-dev: generate-manifests generate-dockerfile-links
	skaffold dev

# builds and deploys the entire app
.PHONY: skaffold-run
skaffold-run: lint test generate-manifests generate-dockerfile-links
	skaffold run

.PHONY: skaffold-delete
skaffold-delete: generate-dockerfile-links
	skaffold delete