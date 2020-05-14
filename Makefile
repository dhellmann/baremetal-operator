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

##
## To auto-generate help text, include a comment starting with "##" on
## the line with the target, as below for the "help" target.
##
.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: test ## Run local developer tests
test: generate lint lint-gofmt unit

.PHONY: lint-all ## Run all linter checks
lint-all: lint gosec lint-generate lint-gofmt vet markdownlint shellcheck

.PHONY: generate ## Run code generator for go, CRD, and OpenAPI
generate: bin/operator-sdk
	./bin/operator-sdk generate $(VERBOSE) k8s
	./bin/operator-sdk generate $(VERBOSE) crds
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

.PHONY: unit
unit: ## Run unit tests
	./hack/unit.sh

.PHONY: unit-local
unit-local: ## Run unit tests outside of a container
	go test $(GO_TEST_FLAGS) ./cmd/... ./pkg/...

.PHONY: unit-cover
unit-cover: ## Run unit tests outside of a container with coverage
	go test -coverprofile=cover.out $(GO_TEST_FLAGS) ./cmd/... ./pkg/...
	go tool cover -func=cover.out

.PHONY: unit-cover-html
unit-cover-html:
	go test -coverprofile=cover.out $(GO_TEST_FLAGS) ./cmd/... ./pkg/...
	go tool cover -html=cover.out

.PHONY: lint
lint: ## Run go lint
	./hack/golint.sh

.PHONY: lint-generate
lint-generate: ## Run the code generator and error if it makes any changes
	./hack/generate.sh

.PHONY: lint-generate-local
lint-generate-local:
	IS_CONTAINER=local ./hack/generate.sh

.PHONY: gosec
gosec: ## Run gosec
	./hack/gosec.sh

.PHONY: gofmt
gofmt: ## Run gofmt and let it update files locally
	gofmt -l -w ./pkg ./cmd

.PHONY: lint-gofmt
lint-gofmt: ## Run gofmt and error if it makes any changes
	./hack/gofmt.sh

.PHONY: vet
vet: ## Run go vet
	./hack/govet.sh

.PHONY: markdownlint
markdownlint: ## Run markdown text linter
	./hack/markdownlint.sh

.PHONY: shellcheck
shellcheck: ## Run shell script linter
	./hack/shellcheck.sh

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
run: ## Run the operator outside of a cluster in developer mode
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

.PHONY: docker ## Build docker images
docker: docker-operator docker-sdk docker-golint

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
	go build -o build/_output/bin/baremetal-operator cmd/manager/main.go

.PHONY: tools
tools:
	@echo LDFLAGS=$(LDFLAGS)
	go build -o build/_output/bin/get-hardware-details cmd/get-hardware-details/main.go

.PHONY: deploy
deploy:
	cd deploy && kustomize edit set namespace $(RUN_NAMESPACE) && cd ..
	kustomize build deploy | kubectl apply -f -
