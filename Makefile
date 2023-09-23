GO_MODULE:=github.com/nico151999/high-availability-expense-splitter
STEP_ARCH:=amd64
GOMPLATE_ARCH:=amd64
BUF_ARCH:=x86_64
KIND_ARCH:=amd64
HELM_ARCH:=amd64
KUBECTL_ARCH:=amd64
SKAFFOLD_ARCH:=amd64
PNPM_ARCH:=x64
KIND_CLUSTER_NAME:=ha-expense-splitter-dev
KIND_VERSION:=0.19.0
HELM_VERSION:=3.12.0
PNPM_VERSION:=8.6.2
STEP_VERSION:=0.24.4
KUBECTL_VERSION:=1.27.3
SKAFFOLD_VERSION:=2.6.1
BUF_VERSION:=1.26.1
GOMPLATE_VERSION:=3.11.5
GOTAG_VERSION:=0.6.2
GOLANGCI_VERSION:=1.49.0
SKIP_BREAKING_CHANGES_CHECK:=false
# skips GOLANGCI-related tasks; useful during pipeline execution in dev mode to speed up development
SKIP_GOLANGCI:=true
REPO_ROOT_PATH:=$(shell pwd)
EXPENSESPLITTER_FRONTEND_DEV_PORT=8080
DOCUMENTATION_SVC_DIR:=$(REPO_ROOT_PATH)/cmd/service/documentation
REFLECTION_SVC_DIR:=$(REPO_ROOT_PATH)/cmd/service/reflection
GROUP_SVC_DIR:=$(REPO_ROOT_PATH)/cmd/service/group
PERSON_SVC_DIR:=$(REPO_ROOT_PATH)/cmd/service/person
CURRENCY_SVC_DIR:=$(REPO_ROOT_PATH)/cmd/service/currency
CATEGORY_SVC_DIR:=$(REPO_ROOT_PATH)/cmd/service/category
EXPENSE_CATEGORY_RELATION_SVC_DIR:=$(REPO_ROOT_PATH)/cmd/service/expensecategoryrelation
EXPENSE_STAKE_SVC_DIR:=$(REPO_ROOT_PATH)/cmd/service/expensestake
EXPENSE_SVC_DIR:=$(REPO_ROOT_PATH)/cmd/service/expense
GROUP_PROCESSOR_DIR:=$(REPO_ROOT_PATH)/cmd/processor/group
PERSON_PROCESSOR_DIR:=$(REPO_ROOT_PATH)/cmd/processor/person
CURRENCY_PROCESSOR_DIR:=$(REPO_ROOT_PATH)/cmd/processor/currency
CATEGORY_PROCESSOR_DIR:=$(REPO_ROOT_PATH)/cmd/processor/category
EXPENSE_CATEGORY_RELATION_PROCESSOR_DIR:=$(REPO_ROOT_PATH)/cmd/processor/expensecategoryrelation
EXPENSE_STAKE_PROCESSOR_DIR:=$(REPO_ROOT_PATH)/cmd/processor/expensestake
EXPENSE_PROCESSOR_DIR:=$(REPO_ROOT_PATH)/cmd/processor/expense
OUT_DIR:=$(REPO_ROOT_PATH)/gen
BIN_INSTALL_DIR:=$(OUT_DIR)/bin
HELM_PLUGIN_INSTALL_DIR:=$(BIN_INSTALL_DIR)/plugins/helm
STEP_INSTALL_LOCATION:=$(BIN_INSTALL_DIR)/step
HELM_INSTALL_LOCATION:=$(BIN_INSTALL_DIR)/helm
PNPM_INSTALL_LOCATION:=$(BIN_INSTALL_DIR)/pnpm
GOMPLATE_INSTALL_LOCATION:=$(BIN_INSTALL_DIR)/gomplate
GOTAG_INSTALL_LOCATION:=$(BIN_INSTALL_DIR)/protoc-gen-gotag
BUF_INSTALL_LOCATION:=$(BIN_INSTALL_DIR)/buf
KIND_INSTALL_LOCATION:=$(BIN_INSTALL_DIR)/kind
GOLANGCI_LINT_INSTALL_LOCATION:=$(BIN_INSTALL_DIR)/golangci-lint
KUBECTL_INSTALL_LOCATION:=$(BIN_INSTALL_DIR)/kubectl
SKAFFOLD_INSTALL_LOCATION:=$(BIN_INSTALL_DIR)/skaffold
CERT_OUT_DIR:=$(OUT_DIR)/cert
DOC_OUT_DIR:=$(OUT_DIR)/doc
LIB_OUT_DIR:=$(OUT_DIR)/lib
APPLICATION_OUT_DIR:=$(OUT_DIR)/application
GO_LIB_OUT_DIR:=$(LIB_OUT_DIR)/go
DOCUMENTATION_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(DOCUMENTATION_SVC_DIR))
REFLECTION_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(REFLECTION_SVC_DIR))
GROUP_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(GROUP_SVC_DIR))
PERSON_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(PERSON_SVC_DIR))
CURRENCY_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(CURRENCY_SVC_DIR))
CATEGORY_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(CATEGORY_SVC_DIR))
EXPENSE_CATEGORY_RELATION_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_CATEGORY_RELATION_SVC_DIR))
EXPENSE_STAKE_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_STAKE_SVC_DIR))
EXPENSE_SVC_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_SVC_DIR))
GROUP_PROCESSOR_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(GROUP_PROCESSOR_DIR))
PERSON_PROCESSOR_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(PERSON_PROCESSOR_DIR))
CURRENCY_PROCESSOR_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(CURRENCY_PROCESSOR_DIR))
CATEGORY_PROCESSOR_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(CATEGORY_PROCESSOR_DIR))
EXPENSE_CATEGORY_RELATION_PROCESSOR_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_CATEGORY_RELATION_PROCESSOR_DIR))
EXPENSE_STAKE_PROCESSOR_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_STAKE_PROCESSOR_DIR))
EXPENSE_PROCESSOR_OUT_DIR:=$(APPLICATION_OUT_DIR)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_PROCESSOR_DIR))

# prioritise executables in the repo's bin dir
export PATH=$(BIN_INSTALL_DIR):$(shell echo $$PATH)

# define where helm will put its plugins
export HELM_PLUGINS=$(HELM_PLUGIN_INSTALL_DIR)

# install the buf commandline tool required to run buf commands
.PHONY: install-buf
install-buf:
# TODO: also check if correct version is installed
ifeq ($(wildcard $(BUF_INSTALL_LOCATION)),)
	mkdir -p $(BIN_INSTALL_DIR)
	curl -o $(BIN_INSTALL_DIR)/buf -sSL https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-Linux-$(BUF_ARCH)
	chmod 755 $(BIN_INSTALL_DIR)/buf
ifneq ($(BIN_INSTALL_DIR)/buf,$(BUF_INSTALL_LOCATION))
	mkdir -p $(shell dirname $(BUF_INSTALL_LOCATION))
	mv $(BIN_INSTALL_DIR)/buf $(BUF_INSTALL_LOCATION)
endif
endif

# install the kind commandline tool required to create and manage local K8s clusters
.PHONY: install-kind
install-kind:
# TODO: also check if correct version is installed
ifeq ($(wildcard $(KIND_INSTALL_LOCATION)),)
	mkdir -p $(BIN_INSTALL_DIR)
	curl -o $(BIN_INSTALL_DIR)/kind -sSL https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-linux-$(KIND_ARCH)
	chmod 755 $(BIN_INSTALL_DIR)/kind
ifneq ($(BIN_INSTALL_DIR)/kind,$(KIND_INSTALL_LOCATION))
	mkdir -p $(shell dirname $(KIND_INSTALL_LOCATION))
	mv $(BIN_INSTALL_DIR)/kind $(KIND_INSTALL_LOCATION)
endif
endif

# install the gomplate commandline tool for rendering go template files
.PHONY: install-gomplate
install-gomplate:
# TODO: also check if correct version is installed
ifeq ($(wildcard $(GOMPLATE_INSTALL_LOCATION)),)
	mkdir -p $(BIN_INSTALL_DIR)
	curl -o $(BIN_INSTALL_DIR)/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/download/v$(GOMPLATE_VERSION)/gomplate_linux-$(GOMPLATE_ARCH)
	chmod 755 $(BIN_INSTALL_DIR)/gomplate
ifneq ($(BIN_INSTALL_DIR)/gomplate,$(GOMPLATE_INSTALL_LOCATION))
	mkdir -p $(shell dirname $(GOMPLATE_INSTALL_LOCATION))
	mv $(BIN_INSTALL_DIR)/gomplate $(GOMPLATE_INSTALL_LOCATION)
endif
endif

# install the gotag protoc plugin for tagging structs in generated go files
.PHONY: install-gotag
install-gotag:
# TODO: also check if correct version is installed
ifeq ($(wildcard $(GOTAG_INSTALL_LOCATION)),)
	GOBIN=$(BIN_INSTALL_DIR) go install github.com/srikrsna/protoc-gen-gotag@v$(GOTAG_VERSION)
ifneq ($(BIN_INSTALL_DIR)/protoc-gen-gotag,$(GOTAG_INSTALL_LOCATION))
	mkdir -p $(shell dirname $(GOTAG_INSTALL_LOCATION))
	mv $(BIN_INSTALL_DIR)/protoc-gen-gotag $(GOTAG_INSTALL_LOCATION)
endif
endif

# install the kubeconform helm plugin used to check validity of helm charts
.PHONY: install-kubeconform
install-kubeconform: install-helm
	mkdir -p $(HELM_PLUGIN_INSTALL_DIR)
	if ! $(HELM_INSTALL_LOCATION) plugin list | grep -q 'kubeconform'; then \
		$(HELM_INSTALL_LOCATION) plugin install 'https://github.com/jtyr/kubeconform-helm'; \
	fi

# install golang-ci llinter
.PHONY: install-golangci-lint
install-golangci-lint:
# TODO: also check if correct version is installed
ifneq (true,$(SKIP_GOLANGCI))
ifeq ($(wildcard $(GOLANGCI_LINT_INSTALL_LOCATION)),)
	GOBIN=$(BIN_INSTALL_DIR) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_VERSION)
ifneq ($(BIN_INSTALL_DIR)/golangci-lint,$(GOLANGCI_LINT_INSTALL_LOCATION))
	mkdir -p $(shell dirname $(GOLANGCI_LINT_INSTALL_LOCATION))
	mv $(BIN_INSTALL_DIR)/golangci-lint $(GOLANGCI_LINT_INSTALL_LOCATION)
endif
endif
endif

# install step
.PHONY: install-step
install-step:
# TODO: also check if correct version is installed
ifeq (,$(wildcard $(STEP_INSTALL_LOCATION)))
	mkdir -p $(OUT_DIR)/tmp
	curl -fsSL https://github.com/smallstep/cli/releases/download/v$(STEP_VERSION)/step_linux_$(STEP_VERSION)_$(STEP_ARCH).tar.gz -o $(OUT_DIR)/tmp/step.tar.gz
	tar -xzf $(OUT_DIR)/tmp/step.tar.gz -C $(OUT_DIR)/tmp
	mkdir -p $(BIN_INSTALL_DIR)
	mv $(OUT_DIR)/tmp/step_$(STEP_VERSION)/bin/step $(STEP_INSTALL_LOCATION)
	rm -r $(OUT_DIR)/tmp
endif

# install kubectl
.PHONY: install-kubectl
install-kubectl:
# TODO: also check if correct version is installed
ifeq (,$(wildcard $(KUBECTL_INSTALL_LOCATION)))
	mkdir -p $(BIN_INSTALL_DIR)
	curl -L "https://dl.k8s.io/release/v$(KUBECTL_VERSION)/bin/linux/$(KUBECTL_ARCH)/kubectl" -o $(KUBECTL_INSTALL_LOCATION)
	chmod +x $(KUBECTL_INSTALL_LOCATION)
endif

# install helm
.PHONY: install-helm
install-helm: install-kubectl
# TODO: also check if correct version is installed
ifeq (,$(wildcard $(HELM_INSTALL_LOCATION)))
	mkdir -p $(OUT_DIR)/tmp
	curl https://get.helm.sh/helm-v$(HELM_VERSION)-linux-$(HELM_ARCH).tar.gz -o $(OUT_DIR)/tmp/helm.tar.gz
	tar -xzf $(OUT_DIR)/tmp/helm.tar.gz -C $(OUT_DIR)/tmp
	mkdir -p $(BIN_INSTALL_DIR)
	mv $(OUT_DIR)/tmp/linux-$(HELM_ARCH)/helm $(HELM_INSTALL_LOCATION)
	rm -r $(OUT_DIR)/tmp
endif

.PHONY: install-pnpm
install-pnpm:
# TODO: also check if correct version is installed
ifeq (,$(wildcard $(PNPM_INSTALL_LOCATION)))
	mkdir -p $(BIN_INSTALL_DIR)
	curl -fsSL "https://github.com/pnpm/pnpm/releases/download/v${PNPM_VERSION}/pnpm-linuxstatic-$(PNPM_ARCH)" -o $(PNPM_INSTALL_LOCATION)
	chmod +x $(PNPM_INSTALL_LOCATION)
endif

.PHONY: install-skaffold
install-skaffold: install-helm install-kubectl
# TODO: also check if correct version is installed
ifeq (,$(wildcard $(SKAFFOLD_INSTALL_LOCATION)))
	mkdir -p $(BIN_INSTALL_DIR)
	curl -L https://storage.googleapis.com/skaffold/releases/v$(SKAFFOLD_VERSION)/skaffold-linux-$(SKAFFOLD_ARCH) -o $(SKAFFOLD_INSTALL_LOCATION)
	chmod +x $(SKAFFOLD_INSTALL_LOCATION)
endif

.PHONY: pnpm-install
pnpm-install: install-pnpm
	pnpm install

# updates protobuf dependencies
.PHONY: update-buf-dependencies
update-buf-dependencies: install-buf generate-buf
	$(BUF_INSTALL_LOCATION) mod update proto

# format code; should be run before something is merged into main branch; consequently this should be part of our pipeline
.PHONY: format
format: install-buf generate-buf
	$(BUF_INSTALL_LOCATION) format -w
	go fmt ./...

# generates config files required to run buf correctly
.PHONY: generate-buf
generate-buf: install-gomplate
	echo '{"goModule": "$(GO_MODULE)", "relativeGoLibOutDir": "$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(GO_LIB_OUT_DIR))"}' | \
	$(GOMPLATE_INSTALL_LOCATION) -d 'data=stdin:?type=application/json' -f buf.gen.yaml.tpl -o buf.gen.yaml
	echo '{"relativeGoLibOutDir": "$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(GO_LIB_OUT_DIR))"}' | \
	$(GOMPLATE_INSTALL_LOCATION) -d 'data=stdin:?type=application/json' -f buf.gen.tag.yaml.tpl -o buf.gen.tag.yaml

# generate files from proto definitions using buf
.PHONY: generate-proto
generate-proto: install-buf generate-buf
# if there are neither docs nor libs generated by buf
ifeq (,$(wildcard $(DOC_OUT_DIR))$(wildcard $(LIB_OUT_DIR)))
	$(BUF_INSTALL_LOCATION) generate
else
	@echo 'There are buf-generated files already. Consider cleaning them first. Skipping proto generation...'
endif

.PHONY: generate-proto-with-gotag
generate-proto-with-gotag: install-buf generate-proto install-gotag generate-buf
	$(BUF_INSTALL_LOCATION) generate --template buf.gen.tag.yaml

# generate files from proto definitions using buf and initializes node package
.PHONY: generate-proto-with-node
generate-proto-with-node: generate-proto install-pnpm
ifeq (,$(wildcard $(LIB_OUT_DIR)/ts/package.json))
	cd $(LIB_OUT_DIR)/ts && \
	$(PNPM_INSTALL_LOCATION) init && \
	$(PNPM_INSTALL_LOCATION) install @bufbuild/protobuf @bufbuild/protoc-gen-es
else
	@echo 'There are node package files already. Consider cleaning them first. Skipping node package generation...'
endif

# generate Dockerfile links required for using namespaced dockerignore files: https://github.com/moby/moby/issues/12886#issuecomment-480575928
.PHONY: generate-dockerfile-links
generate-dockerfile-links:
	ln -sf Dockerfile ./cmd/service/documentation.Dockerfile
	ln -sf Dockerfile ./cmd/service/reflection.Dockerfile
	ln -sf Dockerfile ./cmd/service/group.Dockerfile
	ln -sf Dockerfile ./cmd/service/person.Dockerfile
	ln -sf Dockerfile ./cmd/service/currency.Dockerfile
	ln -sf Dockerfile ./cmd/service/category.Dockerfile
	ln -sf Dockerfile ./cmd/service/expensecategoryrelation.Dockerfile
	ln -sf Dockerfile ./cmd/service/expense.Dockerfile
	ln -sf Dockerfile ./cmd/service/expensestake.Dockerfile
	ln -sf Dockerfile ./cmd/processor/group.Dockerfile
	ln -sf Dockerfile ./cmd/processor/person.Dockerfile
	ln -sf Dockerfile ./cmd/processor/currency.Dockerfile
	ln -sf Dockerfile ./cmd/processor/category.Dockerfile
	ln -sf Dockerfile ./cmd/processor/expensecategoryrelation.Dockerfile
	ln -sf Dockerfile ./cmd/processor/expense.Dockerfile
	ln -sf Dockerfile ./cmd/processor/expensestake.Dockerfile

# generates new certs for Linkerd communication and overwrites existing ones
.PHONY: build
generate-cert: install-step
# if ca.crt has not been created yet
ifeq (,$(wildcard $(CERT_OUT_DIR)/ca.crt)$(wildcard $(CERT_OUT_DIR)/ca.key))
	mkdir -p $(CERT_OUT_DIR)
	$(STEP_INSTALL_LOCATION) certificate create root.linkerd.cluster.local $(CERT_OUT_DIR)/ca.crt $(CERT_OUT_DIR)/ca.key --profile root-ca --no-password --insecure
else
	@echo 'Skipping cert generation cause it was performed already'
endif

# performs all code generation tasks
.PHONY: generate
generate: generate-proto-with-node generate-proto-with-gotag generate-dockerfile-links generate-cert

# lint code; should be run before something is merged into master branch; consequently this should be part of our pipeline
.PHONY: lint
lint: install-buf generate-proto-with-node install-kubeconform install-helm generate-buf $(if $(findstring $(SKIP_GOLANGCI),false),install-golangci-lint)
	$(BUF_INSTALL_LOCATION) lint
ifeq (false,$(SKIP_BREAKING_CHANGES_CHECK))
	$(BUF_INSTALL_LOCATION) breaking --against '.git#branch=main'
endif
	$(HELM_INSTALL_LOCATION) kubeconform --verbose --summary '$(REPO_ROOT_PATH)/charts/ha-expense-splitter'
	$(HELM_INSTALL_LOCATION) kubeconform --verbose --summary '$(REPO_ROOT_PATH)/charts/linkerd-cert-config'
	$(HELM_INSTALL_LOCATION) kubeconform --verbose --summary '$(REPO_ROOT_PATH)/charts/stackgres-cluster'
	go vet ./...
ifeq (false,$(SKIP_GOLANGCI))
	$(GOLANGCI_LINT_INSTALL_LOCATION) run --concurrency 1 --verbose
endif

.PHONY: clean-lib
clean-lib:
	rm -rf $(LIB_OUT_DIR)

.PHONY: clean-doc
clean-doc:
	rm -rf $(DOC_OUT_DIR)

.PHONY: clean-proto
clean-proto: clean-doc clean-lib

.PHONY: clean-application
clean-application:
	rm -rf $(APPLICATION_OUT_DIR)

.PHONY: clean-cert
clean-cert:
	rm -rf $(CERT_OUT_DIR)

.PHONY: clean-bin
clean-bin:
	rm -rf $(BIN_INSTALL_DIR)

# cleanup generated files
.PHONY: clean
clean: clean-lib clean-doc clean-proto clean-application clean-cert clean-bin

test: generate format lint generate-proto-with-node
	go test ./... -coverprofile cover.out
# TODO: also test TS

.PHONY: run-expensesplitter-frontend
run-expensesplitter-frontend: pnpm-install install-pnpm
	$(PNPM_INSTALL_LOCATION) -C '$(REPO_ROOT_PATH)/frontend/expense_splitter' dev --host --port $(EXPENSESPLITTER_FRONTEND_DEV_PORT)

.PHONY: build-expensesplitter-frontend
build-expensesplitter-frontend: pnpm-install install-pnpm
	$(PNPM_INSTALL_LOCATION) -C '$(REPO_ROOT_PATH)/frontend/expense_splitter' build

# build documentation UI
.PHONY: build-documentation
build-documentation: generate-proto
	CGO_ENABLED=0 go build -o $(DOCUMENTATION_SVC_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(DOCUMENTATION_SVC_DIR))

# builds reflection service
.PHONY: build-reflection-service
build-reflection-service: generate-proto
	CGO_ENABLED=0 go build -o $(REFLECTION_SVC_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(REFLECTION_SVC_DIR))

# builds group service
.PHONY: build-group-service
build-group-service: generate-proto
	CGO_ENABLED=0 go build -o $(GROUP_SVC_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(GROUP_SVC_DIR))

# builds person service
.PHONY: build-person-service
build-person-service: generate-proto
	CGO_ENABLED=0 go build -o $(PERSON_SVC_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(PERSON_SVC_DIR))

# builds currency service
.PHONY: build-currency-service
build-currency-service: generate-proto
	CGO_ENABLED=0 go build -o $(CURRENCY_SVC_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(CURRENCY_SVC_DIR))

# builds category service
.PHONY: build-category-service
build-category-service: generate-proto
	CGO_ENABLED=0 go build -o $(CATEGORY_SVC_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(CATEGORY_SVC_DIR))

# builds expense category relation service
.PHONY: build-expensecategoryrelation-service
build-expensecategoryrelation-service: generate-proto
	CGO_ENABLED=0 go build -o $(EXPENSE_CATEGORY_RELATION_SVC_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_CATEGORY_RELATION_SVC_DIR))

# builds expense stake service
.PHONY: build-expensestake-service
build-expensestake-service: generate-proto
	CGO_ENABLED=0 go build -o $(EXPENSE_STAKE_SVC_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_STAKE_SVC_DIR))

# builds expense service
.PHONY: build-expense-service
build-expense-service: generate-proto
	CGO_ENABLED=0 go build -o $(EXPENSE_SVC_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_SVC_DIR))

# builds group processor
.PHONY: build-group-processor
build-group-processor: generate-proto
	CGO_ENABLED=0 go build -o $(GROUP_PROCESSOR_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(GROUP_PROCESSOR_DIR))

# builds person processor
.PHONY: build-person-processor
build-person-processor: generate-proto
	CGO_ENABLED=0 go build -o $(PERSON_PROCESSOR_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(PERSON_PROCESSOR_DIR))

# builds currency processor
.PHONY: build-currency-processor
build-currency-processor: generate-proto
	CGO_ENABLED=0 go build -o $(CURRENCY_PROCESSOR_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(CURRENCY_PROCESSOR_DIR))

# builds category processor
.PHONY: build-category-processor
build-category-processor: generate-proto
	CGO_ENABLED=0 go build -o $(CATEGORY_PROCESSOR_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(CATEGORY_PROCESSOR_DIR))

# builds expensecategoryrelation processor
.PHONY: build-expensecategoryrelation-processor
build-expensecategoryrelation-processor: generate-proto
	CGO_ENABLED=0 go build -o $(EXPENSE_CATEGORY_RELATION_PROCESSOR_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_CATEGORY_RELATION_PROCESSOR_DIR))

# builds expense stake processor
.PHONY: build-expensestake-processor
build-expensestake-processor: generate-proto
	CGO_ENABLED=0 go build -o $(EXPENSE_STAKE_PROCESSOR_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_STAKE_PROCESSOR_DIR))

# builds expense processor
.PHONY: build-expense-processor
build-expense-processor: generate-proto
	CGO_ENABLED=0 go build -o $(EXPENSE_PROCESSOR_OUT_DIR) $(GO_MODULE)/$(shell realpath -m --relative-to $(REPO_ROOT_PATH) $(EXPENSE_PROCESSOR_DIR))

# starts the dev mode of skaffold
.PHONY: skaffold-dev
skaffold-dev: install-skaffold generate-dockerfile-links
	$(SKAFFOLD_INSTALL_LOCATION) dev

# builds and deploys the entire app
.PHONY: skaffold-run
skaffold-run: install-skaffold $(if $(findstring $(HA_EXPENSE_SPLITTER_SKIP_EXPENSE_SPLITTER_INSTALLATION),false),lint test generate-dockerfile-links)
	$(SKAFFOLD_INSTALL_LOCATION) run

.PHONY: skaffold-delete
skaffold-delete: install-skaffold generate-dockerfile-links
	$(SKAFFOLD_INSTALL_LOCATION) delete

# creates a local cluster for dev purposes
.PHONY: kind-create
kind-create: install-kind install-helm install-kubectl
	$(KIND_INSTALL_LOCATION) create cluster --config ./kind-config.yaml --name $(KIND_CLUSTER_NAME)
	$(KUBECTL_INSTALL_LOCATION) wait --for=condition=Ready nodes --all --timeout=120s
	$(KUBECTL_INSTALL_LOCATION) create namespace metallb-system
	$(KUBECTL_INSTALL_LOCATION) label namespaces metallb-system pod-security.kubernetes.io/enforce=privileged pod-security.kubernetes.io/audit=privileged pod-security.kubernetes.io/warn=privileged
	$(HELM_INSTALL_LOCATION) repo add metallb https://metallb.github.io/metallb
	$(HELM_INSTALL_LOCATION) install metallb metallb/metallb -n metallb-system
	$(KUBECTL_INSTALL_LOCATION) wait --namespace metallb-system --for=condition=ready pod --selector=app.kubernetes.io/name=metallb --timeout=600s
	IP_PREFIX=$$(docker network inspect -f '{{(index .IPAM.Config 0).Subnet}}' kind | cut -f1 -d"/" | cut -f1-2 -d".") && \
	echo "{\"startIP\": \"$$IP_PREFIX.255.200\", \"endIP\": \"$$IP_PREFIX.255.255\"}" | \
	$(GOMPLATE_INSTALL_LOCATION) -d 'data=stdin:?type=application/json' -f metallb-config.yaml.tpl | \
	$(KUBECTL_INSTALL_LOCATION) apply -n metallb-system -f -

.PHONY: kind-delete
kind-delete: install-kind
	$(KIND_INSTALL_LOCATION) delete cluster --name $(KIND_CLUSTER_NAME)