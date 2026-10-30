package main

import (
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"github.com/zeropsio/pulumi-zerops/sdk/go/zerops"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		appName := cfg.Require("appName")
		message := "Hello from Zerops"
		if value, err := cfg.Try("message"); err == nil {
			message = value
		}
		deploymentYAML, err := os.ReadFile("./app/zerops.yml")
		if err != nil {
			return err
		}
		project, err := zerops.NewProject(ctx, "project", &zerops.ProjectArgs{
			Name: pulumi.String(appName),
			Mode: pulumi.String("LIGHT"),
		})
		if err != nil {
			return err
		}
		service, err := zerops.NewService(ctx, "app", &zerops.ServiceArgs{
			ProjectId: project.ID(),
			Hostname:  pulumi.String("app"),
			Type:      pulumi.String("nodejs@22"),
		})
		if err != nil {
			return err
		}
		environment, err := zerops.NewServiceEnv(ctx, "message", &zerops.ServiceEnvArgs{
			ServiceId: service.ID(),
			Key:       pulumi.String("MESSAGE"),
			Content:   pulumi.String(message),
		})
		if err != nil {
			return err
		}
		deployment, err := zerops.NewServiceDeploy(ctx, "deploy", &zerops.ServiceDeployArgs{
			ServiceId:  service.ID(),
			SourceDir:  pulumi.String("./app"),
			ZeropsYaml: pulumi.String(deploymentYAML),
			Setup:      pulumi.String("app"),
		}, pulumi.DependsOn([]pulumi.Resource{environment}))
		if err != nil {
			return err
		}
		access, err := zerops.NewServiceSubdomainAccess(ctx, "public", &zerops.ServiceSubdomainAccessArgs{
			ServiceId: service.ID(),
		}, pulumi.DependsOn([]pulumi.Resource{deployment}))
		if err != nil {
			return err
		}
		ctx.Export("projectId", project.ID())
		ctx.Export("serviceId", service.ID())
		ctx.Export("url", access.Urls.Index(pulumi.Int(0)))
		return nil
	})
}
