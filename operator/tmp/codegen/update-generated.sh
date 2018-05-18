#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

DOCKER_REPO_ROOT="/go/src/icm-varnish-k8s-operator/operator"
IMAGE=${IMAGE:-"gcr.io/coreos-k8s-scale-testing/codegen:1.9.3"}

docker run --rm \
  -v "$PWD":"$DOCKER_REPO_ROOT" \
  -w "$DOCKER_REPO_ROOT" \
  "$IMAGE" \
  "/go/src/k8s.io/code-generator/generate-groups.sh"  \
  "deepcopy" \
  "icm-varnish-k8s-operator/operator/pkg/generated" \
  "icm-varnish-k8s-operator/operator/pkg/apis" \
  "icm:v1alpha1" \
  --go-header-file "./tmp/codegen/boilerplate.go.txt" \
  $@
