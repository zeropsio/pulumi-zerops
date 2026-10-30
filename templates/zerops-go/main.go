package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/zeropsio/pulumi-zerops/sdk-app/go/zeropsapp"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		database := cfg.GetBool("database") // Optional PostgreSQL; off by default.
		source := "./app"
		// WebApp handles the project, deployment, public URL, and database wiring.
		web, err := zeropsapp.NewWebApp(ctx, "web", &zeropsapp.WebAppArgs{
			AppName:  cfg.Require("appName"),
			Runtime:  "python@3.12",
			Source:   &source,
			Database: &database,
		})
		if err != nil {
			return err
		}
		ctx.Export("url", web.Url)
		ctx.Export("projectId", web.ProjectId)
		ctx.Export("serviceId", web.ServiceId)
		return nil
	})
}
