package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"

	pf "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"
	zerops "github.com/zeropsio/pulumi-zerops/provider"
	"github.com/zeropsio/pulumi-zerops/provider/pkg/version"
)

//go:embed schema.json
var packageSchema []byte

//go:embed bridge-metadata.json
var bridgeMetadata []byte

func main() {
	schema, err := withVersion(packageSchema, version.Version)
	if err != nil {
		log.Fatalf("embedded schema: %v", err)
	}
	pf.Main(context.Background(), "zerops", zerops.Provider(), pf.ProviderMetadata{
		PackageSchema: schema, BridgeMetadata: bridgeMetadata,
	})
}

// withVersion stamps the release version into the served schema. The committed schema stays
// unversioned; dependent packages (zerops-app) need the version to generate their SDKs.
func withVersion(schema []byte, v string) ([]byte, error) {
	var spec map[string]json.RawMessage
	if err := json.Unmarshal(schema, &spec); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	spec["version"] = raw
	return json.Marshal(spec)
}
