package main

import (
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfgen"
	zerops "github.com/zeropsio/pulumi-zerops/provider"
	"os/exec"
	"strings"
)

func main() {
	info := zerops.Provider()
	// Resolve documentation from the pinned upstream release's module directory.
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/zeropsio/terraform-provider-zerops")
	command.Dir = "provider"
	directory, err := command.Output()
	if err != nil {
		panic(err)
	}
	info.UpstreamRepoPath = strings.TrimSpace(string(directory))
	tfgen.Main("zerops", info)
}
