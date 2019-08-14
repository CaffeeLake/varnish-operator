# Image URL to use in all building/pushing image targets
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
VERSION ?= $(shell cat ${ROOT_DIR}version.txt)
PUBLISH_IMG ?= varnish-controller:${VERSION}
VARNISH_PUBLISH_IMG ?= varnish:${VERSION}
VARNISH_IMG ?= ${VARNISH_PUBLISH_IMG}-dev
IMG ?= ${PUBLISH_IMG}-dev
NAMESPACE ?= "default"

# all: test manager
all: test manager kwatcher

# Run tests
test: generate fmt vet manifests
	go test icm-varnish-k8s-operator/pkg/... icm-varnish-k8s-operator/cmd/... -coverprofile cover.out

# Run lint tools
lint:
	golangci-lint run

# Build manager binary
manager: generate fmt vet
	go build -o ${ROOT_DIR}bin/manager icm-varnish-k8s-operator/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	NAMESPACE=${NAMESPACE} LOGLEVEL=debug LOGFORMAT=console CONTAINER_IMAGE=us.icr.io/icm-varnish/${VARNISH_IMG} LEADERELECTION_ENABLED=false go run ${ROOT_DIR}cmd/manager/main.go

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
install: manifests
	kustomize build ${ROOT_DIR}config/default | kubectl apply -f -

uninstall:
	kustomize build ${ROOT_DIR}config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run ${ROOT_DIR}vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

# Run goimports against code
fmt:
	cd ${ROOT_DIR} && goimports -w ./pkg ./cmd

# Run go vet against code
vet:
	cd ${ROOT_DIR} && go vet ./pkg/... ./cmd/...

# Generate code
generate:
	cd ${ROOT_DIR} && go generate ./pkg/... ./cmd/...

# Prepare .yaml files for helm
helm-prepare: manifests
	${ROOT_DIR}hack/create_helm_files.sh ${ROOT_DIR}varnish-operator/templates

helm-upgrade: helm-prepare
	helm upgrade --install --namespace ${NAMESPACE} --force varnish-operator --set operator.controllerImage.tag=${VERSION} --set namespace=${NAMESPACE} ${ROOT_DIR}varnish-operator

# Build the docker image
# docker-build: test
docker-build: test
	docker build ${ROOT_DIR} -t ${IMG} -f Dockerfile

# Tag and push the docker image
docker-tag-push:
ifndef REPO_PATH
	$(error must set REPO_PATH, eg "make docker-tag-push REPO_PATH=us.icr.io/icm-varnish")
endif
ifndef PUBLISH
	docker tag ${IMG} ${REPO_PATH}/${IMG}
	docker push ${REPO_PATH}/${IMG}
else
	docker tag ${IMG} ${REPO_PATH}/${PUBLISH_IMG}
	docker push ${REPO_PATH}/${PUBLISH_IMG}
endif

kwatcher: fmt vet
	go build -o ${ROOT_DIR}bin/kwatcher ${ROOT_DIR}cmd/kwatcher/

# Build the docker image
docker-build-varnish: fmt vet
	docker build ${ROOT_DIR} -t ${VARNISH_IMG} -f Dockerfile.Varnish

docker-tag-push-varnish:
ifndef REPO_PATH
	$(error must set REPO_PATH, eg "make docker-tag-push REPO_PATH=us.icr.io/icm-varnish")
endif
ifndef PUBLISH
	docker tag ${VARNISH_IMG} ${REPO_PATH}/${VARNISH_IMG}
	docker push ${REPO_PATH}/${VARNISH_IMG}
else
	docker tag ${VARNISH_IMG} ${REPO_PATH}/${VARNISH_PUBLISH_IMG}
	docker push ${REPO_PATH}/${VARNISH_PUBLISH_IMG}
endif
