SHELL := /bin/bash
VERSION ?= 0.1.0
LANGUAGES := go python dotnet java nodejs
SCHEMA := cmd/pulumi-resource-zerops-app/schema.json

.PHONY: build schema primitive_plugin gen_sdks build_sdks $(addprefix generate_,$(LANGUAGES)) $(addprefix build_,$(LANGUAGES))
build:
	go build -ldflags '-X main.version=$(VERSION)' -o ../bin/pulumi-resource-zerops-app ./cmd/pulumi-resource-zerops-app
schema: build
	../bin/pulumi-resource-zerops-app --schema > $(SCHEMA)
# Foreign resource references require the matching primitive plugin even when the
# caller starts with an empty PULUMI_HOME. Build it through the root's versioned target.
primitive_plugin:
	$(MAKE) -C .. VERSION=$(VERSION) provider
	pulumi plugin install resource zerops $(VERSION) --file ../bin/pulumi-resource-zerops --reinstall
gen_sdks: $(addprefix generate_,$(LANGUAGES))
$(addprefix generate_,$(LANGUAGES)): generate_%: schema primitive_plugin
	pulumi package gen-sdk $(SCHEMA) --language $* --out ../sdk-app
	python stamp-sdk.py $* $(VERSION)
build_sdks: $(addprefix build_,$(LANGUAGES))
build_go: generate_go
	cd ../sdk-app && go mod tidy && go build ./...
build_python: generate_python
	uv venv --allow-existing ../.cache/python-build
	uv pip install --python ../.cache/python-build/bin/python build
	cd ../sdk-app/python && ../../.cache/python-build/bin/python -m build
build_dotnet: generate_dotnet
	cd ../sdk-app/dotnet && dotnet restore --source ../../sdk/dotnet/bin/Release --source https://api.nuget.org/v3/index.json && dotnet build --configuration Release --no-restore
build_java: generate_java
	cd ../sdk/java && gradle --no-daemon publishToMavenLocal
	cd ../sdk-app/java && gradle --no-daemon build
build_nodejs: generate_nodejs
	cd ../sdk-app/nodejs && npm install --no-save ../../sdk/nodejs && npx tsc
	cp ../sdk-app/nodejs/package.json ../sdk-app/nodejs/README.md ../LICENSE ../sdk-app/nodejs/bin/
