using Pulumi;
using Zerops.Pulumi.App;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    // Your source directory is deployed with a public URL; PostgreSQL is optional.
    var web = new WebApp("web", new WebAppArgs
    {
        AppName = config.Require("appName"),
        Runtime = "python@3.12",
        Source = "./app",
        Database = config.GetBoolean("database") ?? false,
    });

    return new Dictionary<string, object?>
    {
        ["url"] = web.Url,
        ["projectId"] = web.ProjectId,
        ["serviceId"] = web.ServiceId,
    };
});
