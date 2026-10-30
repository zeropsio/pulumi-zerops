"""Your application and its infrastructure, deployed together."""
import pulumi
from pulumi_zerops_app import WebApp

config = pulumi.Config()

# WebApp handles the project, runtime, deployment and public URL for you.
web = WebApp(
    "web",
    app_name=config.require("appName"),
    runtime="python@3.12",
    source="./app",
    # Enable PostgreSQL later with: pulumi config set database true
    database=config.get_bool("database") or False,
)

pulumi.export("url", web.url)
pulumi.export("projectId", web.project_id)
pulumi.export("serviceId", web.service_id)
