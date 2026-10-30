using Pulumi;
using Z = Zerops.Pulumi;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var project = new Z.Project("project", new Z.ProjectArgs
    {
        Name = config.Require("appName"),
        Mode = "LIGHT",
    });
    var service = new Z.Service("app", new Z.ServiceArgs
    {
        ProjectId = project.Id,
        Hostname = "app",
        Type = "nodejs@22",
    });
    var message = new Z.ServiceEnv("message", new Z.ServiceEnvArgs
    {
        ServiceId = service.Id,
        Key = "MESSAGE",
        Content = config.Get("message") ?? "Hello from Zerops",
    });
    var deploy = new Z.ServiceDeploy("deploy", new Z.ServiceDeployArgs
    {
        ServiceId = service.Id,
        SourceDir = "./app",
        ZeropsYaml = File.ReadAllText("./app/zerops.yml"),
        Setup = "app",
    }, new CustomResourceOptions { DependsOn = { message } });
    var access = new Z.ServiceSubdomainAccess("access", new Z.ServiceSubdomainAccessArgs
    {
        ServiceId = service.Id,
    }, new CustomResourceOptions { DependsOn = { deploy } });

    return new Dictionary<string, object?>
    {
        ["projectId"] = project.Id,
        ["serviceId"] = service.Id,
        ["url"] = access.Urls.Apply(urls => urls[0]),
    };
});
