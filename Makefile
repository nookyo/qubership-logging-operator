SHELL=/usr/bin/env bash -o pipefail

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
	GOBIN=$(shell go env GOPATH)/bin
else
	GOBIN=$(shell go env GOBIN)
endif

#############
# Constants #
#############

# Set build version
VERSION=14.20.0
ARTIFACT_NAME=qubership-logging-operator

# Helm charts directory
HELM_FOLDER=charts/qubership-logging-operator

# Directories and files
BUILD_DIR=build
RESULT_DIR=$(BUILD_DIR)/_result
CRDS_DIR=$(BUILD_DIR)/_crds
HELM_DIR=$(BUILD_DIR)/_helm

# Documents folders
PUBLIC_DOC_FOLDER := docs
CRD_FOLDER=$(HELM_FOLDER)/crds
CRD_PUBLIC_DOC_FOLDER=$(PUBLIC_DOC_FOLDER)/crds

# Directories to generate API documentation
TYPES_V1ALPHA1_TARGET=api/v1alpha1/loggingservice_types.go
API_DOC_GEN_BINARY_DIR?=$(shell pwd)/api

# Tools
CONTROLLER_GEN_PACKAGE=sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.5
GEN_CRD_API_PACKAGE=github.com/ahmetb/gen-crd-api-reference-docs@v0.3.0

# Detect the build environment, local or Jenkins builder
BUILD_DATE=$(shell date +"%Y%m%d-%T")
ifndef JENKINS_URL
	BUILD_USER?=$(USER)
	BUILD_BRANCH?=$(shell git branch --show-current)
	BUILD_REVISION?=$(shell git rev-parse --short HEAD)
else
	BUILD_USER=$(BUILD_USER)
	BUILD_BRANCH=$(LOCATION:refs/heads/%=%)
	BUILD_REVISION=$(REPO_HASH)
endif

# The Prometheus common library import path
LOGGING_OPERATOR_PKG=github.com/Netcracker/qubership-logging-operator

# The ldflags for the go build process to set the version related data.
GO_BUILD_LDFLAGS=\
	-s \
	-X $(LOGGING_OPERATOR_PKG)/version.Revision=$(BUILD_REVISION) \
	-X $(LOGGING_OPERATOR_PKG)/version.BuildUser=$(BUILD_USER) \
	-X $(LOGGING_OPERATOR_PKG)/version.BuildDate=$(BUILD_DATE) \
	-X $(LOGGING_OPERATOR_PKG)/version.Branch=$(BUILD_BRANCH) \
	-X $(LOGGING_OPERATOR_PKG)/version.Version=$(VERSION)

# Go build flags
GO_BUILD_RECIPE=\
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_ENABLED=0 \
	go build -ldflags="$(GO_BUILD_LDFLAGS)"

# Default test arguments
TEST_RUN_ARGS=-vet=off --shuffle=on

# List of packages, exclude integration tests that require "envtest"
pkgs = $(shell go list ./... | grep -v /e2e-tests)

# Container name
CONTAINER_CLI?=docker
CONTAINER_NAME="qubership-logging-operator"
DOCKERFILE=Dockerfile

###########
# Generic #
###########

# Default run without arguments
.PHONY: all
all: generate test build-binary image docs archives

# Run only build
.PHONY: build
build: generate build-binary image docs archives

# Run only build inside the Dockerfile
.PHONY: build-image
build-image: generate image docs archives

# Remove all files and directories ignored by git
.PHONY: clean
clean:
	echo "=> Cleanup repository ..."
	git clean -Xfd .

##############
# Generating #
##############

# Generate code (deepcopy files, CRDs)
generate: controller-gen
	echo "=> Generate CRDs and deepcopy ..."
	$(CONTROLLER_GEN) crd:crdVersions={v1} \
					object:headerFile="hack/boilerplate.go.txt" \
					paths="./api/v1alpha1" \
					output:artifacts:config=charts/logging-operator/crds/
	chmod +x ./scripts/build/append-operator-version.sh
	VERSION=$(VERSION) ./scripts/build/append-operator-version.sh

# Find or download controller-gen, download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
		set -e ;\
		CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
		cd $$CONTROLLER_GEN_TMP_DIR ;\
		go mod init tmp ;\
		go install $(CONTROLLER_GEN_PACKAGE) ;\
		rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

#########
# Build #
#########

# Build manager binary
.PHONY: build-binary
build-binary: generate fmt
# TODO: go vet fail build, need to check why?
# build-binary: generate fmt vet
	echo "=> Build binary ..."
	$(GO_BUILD_RECIPE) -o build/_binary/manager cmd/operator/main.go

# Run go fmt against code
.PHONY: fmt
fmt:
	echo "=> Formatting Golang code ..."
	go fmt ./...

# Run go vet against code
.PHONY: vet
vet:
	echo "=> Examines Golang code ..."
	go vet ./...

###############
# Build image #
###############

.PHONY: image
image:
	echo "=> Build image ..."
	docker build --pull -t $(CONTAINER_NAME) -f $(DOCKERFILE) .
	
	# Set image tag if build inside the Jenkins
	for id in $(DOCKER_NAMES) ; do \
		docker tag $(CONTAINER_NAME) "$$id"; \
	done

###########
# Testing #
###########

.PHONY: test
test: unit-test

# Run unit tests in all packages
.PHONY: unit-test
unit-test:
	echo "=> Run Golang unit-tests ..."
	go test -race $(TEST_RUN_ARGS) $(pkgs) -count=1 -v

#################
# Documentation #
#################

# Run document generation
docs: docs/api.md docs/crds

# Run gen-crd-api-reference-docs to generate API documents by operator API
docs/api.md: docs/api/gen $(TYPES_V1ALPHA1_TARGET)
	cd $(API_DOC_GEN_BINARY_DIR) \
	&& $(API_DOC_GEN_BINARY) -api-dir "./v1alpha1/" \
							-config "../scripts/docs/config.json" \
							-template-dir "../scripts/docs/templates" \
							-out-file "../docs/api.md" \
	&& rm -rf $(API_DOC_GEN_BINARY)
	chmod +x ./scripts/build/append-markdown-linter-comments.sh
	./scripts/build/append-markdown-linter-comments.sh

# Find or download gen-crd-api-reference-docs, download gen-crd-api-reference-docs if necessary
docs/api/gen:
ifeq (, $(shell which ./gen-crd-api-reference-docs))
	@{ \
		set -e ;\
		GOBIN=$(API_DOC_GEN_BINARY_DIR) go install $(GEN_CRD_API_PACKAGE) ;\
	}
API_DOC_GEN_BINARY=./gen-crd-api-reference-docs
else
API_DOC_GEN_BINARY=$(shell which gen-crd-api-reference-docs)
endif

# Copy CRDs from the Helm chart to documentation directory
docs/crds:
	rm -rf $(CRD_PUBLIC_DOC_FOLDER)/*.yaml
	cp $(CRD_FOLDER)/* $(CRD_PUBLIC_DOC_FOLDER)/

###################
# Running locally #
###################

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: generate fmt vet
	echo "=> Run ..."
	go run ./cmd/operator/main.go

############
# Archives #
############

# Run archives with helm chart and crds creation
.PHONY: archives
archives: cleanup prepare-charts archive-helm-chart archive-crds

# Remove build dir
.PHONY: cleanup
cleanup:
	rm -rf $(BUILD_DIR)

# Copy Helm charts to the /helm directory because the builder expect it in this dir
.PHONY: prepare-charts
prepare-charts:
	echo "=> Copy Helm charts to contract directory for build ..."
	mkdir -p $(RESULT_DIR)

	# Copy helm chart in folder
	mkdir -p $(HELM_DIR)
	cp -R charts $(HELM_DIR)/

	# Create directories to copy CRDs
	mkdir -p "$(CRDS_DIR)/logging-operator"

# Archive Helm chart separately from application manifest
.PHONY: archive-helm-chart
archive-helm-chart:
	echo "=> Archive Helm charts ..."

	# Navigate to dir to avoid unnecessary directories in result archive
	# name like: logging-operator-14.20.0-chart.zip
	cd "$(HELM_DIR)/charts/" && zip -r "../../../$(RESULT_DIR)/$(ARTIFACT_NAME)-$(VERSION)-chart.zip" ./*

# Archive CRDs separately from helm chart
.PHONY: archive-crds
archive-crds:
	echo "=> Archive CRDs ..."
	# Copy documentation how to apply CRDS
	cp docs/user-guides/manual-create-crds.md "${BUILD_DIR}"/_crds/README.md

	# Copy CRDs from different places in helm chart and sub-charts
	cp charts/logging-operator/crds/* "${BUILD_DIR}/_crds/logging-operator/"

	# Navigate to dir to avoid unnecessary directories in result archive\
	# name like: logging-operator-14.20.0-crds.zip
	cd "$(CRDS_DIR)" && zip -r "../../$(RESULT_DIR)/$(ARTIFACT_NAME)-$(VERSION)-crds.zip" ./*
