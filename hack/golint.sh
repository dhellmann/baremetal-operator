#!/bin/sh

set -eux

IS_CONTAINER=${IS_CONTAINER:-false}
CONTAINER_RUNTIME="${CONTAINER_RUNTIME:-podman}"

if [ "${IS_CONTAINER}" != "false" ]; then
  export XDG_CACHE_HOME="/tmp/.cache"

  find ./pkg ./cmd -type f -name \*.go  \
      | grep -v zz_ \
      | xargs -I'{}' golint -set_exit_status '{}'
else
  "${CONTAINER_RUNTIME}" run --rm \
    --env IS_CONTAINER=TRUE \
    --volume "${PWD}:/go/src/github.com/metal3-io/baremetal-operator:ro,z" \
    --entrypoint sh \
    --workdir /go/src/github.com/metal3-io/baremetal-operator \
    quay.io/metal3-io/golint:latest \
    /go/src/github.com/metal3-io/baremetal-operator/hack/golint.sh "${@}"
fi;
