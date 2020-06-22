TEST_NAMESPACE = operator-test
RUN_NAMESPACE = metal3
GO_TEST_FLAGS = $(VERBOSE)
DEBUG = --debug
SETUP = --no-setup

# See pkg/version.go for details
GIT_COMMIT="$(shell git rev-parse --verify 'HEAD^{commit}')"
export LDFLAGS="-X github.com/metal3-io/baremetal-operator/pkg/version.Raw=$(shell git describe --always --abbrev=40 --dirty) -X github.com/metal3-io/baremetal-operator/pkg/version.Commit=${GIT_COMMIT}"

# Set some variables the operator expects to have in order to work
# Those need to be the same as in deploy/ironic_ci.env
export OPERATOR_NAME=baremetal-operator
export DEPLOY_KERNEL_URL=http://172.22.0.1:6180/images/ironic-python-agent.kernel
export DEPLOY_RAMDISK_URL=http://172.22.0.1:6180/images/ironic-python-agent.initramfs
export IRONIC_ENDPOINT=http://localhost:6385/v1/
export IRONIC_INSPECTOR_ENDPOINT=http://localhost:5050/v1/
export GO111MODULE=on
export GOFLAGS=

CONTROLLER_GEN=./tools/bin/controller-gen
CRD_OPTIONS ?= "crd:trivialVersions=true"

.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo
	@echo "Variables:"
	@echo "  TEST_NAMESPACE   -- project name to use ($(TEST_NAMESPACE))"
	@echo "  SETUP            -- controls the --no-setup flag ($(SETUP))"
	@echo "  GO_TEST_FLAGS    -- flags to pass to --go-test-flags ($(GO_TEST_FLAGS))"
	@echo "  DEBUG            -- debug flag, if any ($(DEBUG))"

.PHONY: test
test: fmt generate lint vet unit ## Run common developer tests

.PHONY: generate
generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."
	$(CONTROLLER_GEN) $(CRD_OPTIONS) \
		rbac:roleName=manager-role \
		webhook \
		paths="./..." \
		output:crd:artifacts:config=deploy/crds
	openapi-gen \
		--input-dirs ./pkg/apis/metal3/v1alpha1 \
		--output-package ./pkg/apis/metal3/v1alpha1 \
		--output-base "" \
		--output-file-base zz_generated.openapi \
		--report-filename "-" \
		--go-header-file /dev/null

bin/operator-sdk: bin
	make -C tools/operator-sdk install

bin:
	mkdir -p bin

.PHONY: travis
travis: unit-verbose lint

.PHONY: unit
unit: ## Run unit tests
	go test $(GO_TEST_FLAGS) ./cmd/... ./pkg/...

.PHONY: unit-cover
unit-cover: ## Run unit tests with code coverage
	go test -coverprofile=cover.out $(GO_TEST_FLAGS) ./cmd/... ./pkg/...
	go tool cover -func=cover.out

.PHONY: unit-cover-html
unit-cover-html:
	go test -coverprofile=cover.out $(GO_TEST_FLAGS) ./cmd/... ./pkg/...
	go tool cover -html=cover.out

.PHONY: unit-verbose
unit-verbose: ## Run unit tests with verbose output
	VERBOSE=-v make unit

.PHONY: linters
linters: sec lint generate-check fmt-check vet ## Run all linters

.PHONY: vet
vet: ## Run go vet
	go vet ./pkg/... ./cmd/...

.PHONY: lint
lint: golint-binary ## Run golint
	find ./pkg ./cmd -type f -name \*.go  |grep -v zz_ | xargs -L1 golint -set_exit_status

.PHONY: generate-check
generate-check:
	./hack/generate.sh

.PHONY: generate-check-local
generate-check-local:
	IS_CONTAINER=local ./hack/generate.sh

.PHONY: sec
sec: $GOPATH/bin/gosec
	gosec -severity medium --confidence medium -quiet ./pkg/... ./cmd/...

$GOPATH/bin/gosec:
	go get -u github.com/securego/gosec/cmd/gosec

.PHONY: golint-binary
golint-binary:
	which golint 2>&1 >/dev/null || $(MAKE) $GOPATH/bin/golint
$GOPATH/bin/golint:
	go get -u golang.org/x/lint/golint

.PHONY: fmt
fmt: ## Run gofmt and write changes to each file
	gofmt -l -w ./pkg ./cmd

.PHONY: fmt-check
fmt-check: ## Run gofmt and report an error if any changes are made
	./hack/gofmt.sh

.PHONY: docs
docs: $(patsubst %.dot,%.png,$(wildcard docs/*.dot))

%.png: %.dot
	dot -Tpng $< >$@

.PHONY: e2e-local
e2e-local:
	operator-sdk test local ./test/e2e \
		--namespace $(TEST_NAMESPACE) \
		--up-local $(SETUP) \
		$(DEBUG) --go-test-flags "$(GO_TEST_FLAGS)"

.PHONY: run
run: ## Run the operator outside of a cluster in development mode
	operator-sdk run --local \
		--go-ldflags=$(LDFLAGS) \
		--watch-namespace=$(RUN_NAMESPACE) \
		--operator-flags="-dev"

.PHONY: demo
demo: ## Run the operator outside of a cluster using the demo driver
	operator-sdk run --local \
		--go-ldflags=$(LDFLAGS) \
		--watch-namespace=$(RUN_NAMESPACE) \
		--operator-flags="-dev -demo-mode"

.PHONY: docker
docker: docker-operator docker-sdk docker-golint ## Build docker images

.PHONY: docker-operator
docker-operator:
	docker build . -f build/Dockerfile

.PHONY: docker-sdk
docker-sdk:
	docker build . -f hack/Dockerfile.operator-sdk

.PHONY: docker-golint
docker-golint:
	docker build . -f hack/Dockerfile.golint

.PHONY: build
build: ## Build the operator binary
	@echo LDFLAGS=$(LDFLAGS)
	go build -ldflags $(LDFLAGS) -o build/_output/bin/baremetal-operator cmd/manager/main.go

.PHONY: tools
tools:
	go build -o build/_output/bin/get-hardware-details cmd/get-hardware-details/main.go

.PHONY: deploy
deploy:
	cd deploy && kustomize edit set namespace $(RUN_NAMESPACE) && cd ..
	kustomize build deploy | kubectl apply -f -

$(CONTROLLER_GEN): tools/bin
	cd tools/controller-tools && go build ./cmd/controller-gen
	cp tools/controller-tools/controller-gen tools/bin/

tools/bin:
	mkdir -p tools/bin
