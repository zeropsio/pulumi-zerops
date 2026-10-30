SHELL := /bin/bash
VERSION ?= 0.1.0
PROVIDER_MODULE := github.com/zeropsio/pulumi-zerops/provider
LDFLAGS := -X $(PROVIDER_MODULE)/pkg/version.Version=$(VERSION)
LANGUAGES := python go dotnet java nodejs
SCHEMA := provider/cmd/pulumi-resource-zerops/schema.json

# Release: make VERSION=X.Y.Z tfgen provider build_sdks re-stamps all SDKs.
.PHONY: tfgen schema provider build_sdks install_plugins test_provider lint check_schema $(addprefix generate_,$(LANGUAGES)) $(addprefix build_,$(LANGUAGES))

bin/pulumi-tfgen-zerops: FORCE
	@mkdir -p bin
	cd provider && go build -ldflags '$(LDFLAGS)' -o ../bin/pulumi-tfgen-zerops ./cmd/pulumi-tfgen-zerops

FORCE:

tfgen schema: bin/pulumi-tfgen-zerops
	bin/pulumi-tfgen-zerops schema --out provider/cmd/pulumi-resource-zerops

provider: tfgen
	cd provider && go build -ldflags '$(LDFLAGS)' -o ../bin/pulumi-resource-zerops ./cmd/pulumi-resource-zerops

# Generate from the stamped schema, using the same CLI code generators for all SDKs.
$(addprefix generate_,$(LANGUAGES)): generate_%: tfgen
	pulumi package gen-sdk $(SCHEMA) --language $* --out sdk
	python scripts/stamp-sdk.py $* $(VERSION)

build_sdks: $(addprefix build_,$(LANGUAGES))

build_python: generate_python
	uv venv --allow-existing .cache/python-build
	uv pip install --python .cache/python-build/bin/python build
	cd sdk/python && ../../.cache/python-build/bin/python -m build

build_go: generate_go
	cd sdk && go mod tidy && go build ./go/...

build_dotnet: generate_dotnet
	cd sdk/dotnet && dotnet build --configuration Release

build_java: generate_java
	cd sdk/java && gradle --no-daemon build

build_nodejs: generate_nodejs
	cd sdk/nodejs && npm install && npx tsc
	cp sdk/nodejs/package.json sdk/nodejs/README.md sdk/nodejs/bin/
	cp LICENSE sdk/nodejs/bin/LICENSE

install_plugins: provider components
	pulumi plugin install resource zerops $(VERSION) --file bin/pulumi-resource-zerops
	pulumi plugin install resource zerops-app $(VERSION) --file bin/pulumi-resource-zerops-app

test_provider:
	cd provider && go test ./...

lint:
	cd provider && go vet ./...

# Git-independent snapshot check: generate into a temporary directory, then diff.
check_schema: bin/pulumi-tfgen-zerops
	@tmp=$$(mktemp -d); trap 'rm -rf "$$tmp"' EXIT; \
	bin/pulumi-tfgen-zerops schema --out "$$tmp" && \
	diff -u $(SCHEMA) "$$tmp/schema.json" && \
	diff -u provider/cmd/pulumi-resource-zerops/bridge-metadata.json "$$tmp/bridge-metadata.json"

.PHONY: components components_schema check_components_schema build_app_sdks test $(addprefix build_app_,$(LANGUAGES))
components:
	$(MAKE) -C components VERSION=$(VERSION) build

components_schema:
	$(MAKE) -C components VERSION=$(VERSION) schema

check_components_schema: components
	@tmp=$$(mktemp); trap 'rm -f "$$tmp"' EXIT; \
	bin/pulumi-resource-zerops-app --schema > "$$tmp" && \
	diff -u components/cmd/pulumi-resource-zerops-app/schema.json "$$tmp"

build_app_sdks: $(addprefix build_app_,$(LANGUAGES))

$(addprefix build_app_,$(LANGUAGES)): build_app_%:
	$(MAKE) -C components VERSION=$(VERSION) build_$*

test: test_provider
	cd components && go test ./...
