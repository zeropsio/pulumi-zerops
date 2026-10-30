"""Deploy a small Node.js application to Zerops using Python infrastructure."""
from pathlib import Path

import pulumi
import pulumi_zerops as zerops

config = pulumi.Config()
app_name = config.require("appName")
message = config.get("message") or "Hello from Zerops"
app_dir = Path(__file__).parent / "app"

project = zerops.Project("project", name=app_name, mode="LIGHT")
service = zerops.Service(
    "app", project_id=project.id, hostname="app", type="nodejs@22"
)
message_env = zerops.ServiceEnv(
    "message", service_id=service.id, key="MESSAGE", content=message
)
deploy = zerops.ServiceDeploy(
    "deploy",
    service_id=service.id,
    source_dir="./app",
    zerops_yaml=(app_dir / "zerops.yml").read_text(encoding="utf-8"),
    setup="app",
    opts=pulumi.ResourceOptions(depends_on=[message_env]),
)
access = zerops.ServiceSubdomainAccess(
    "access", service_id=service.id, opts=pulumi.ResourceOptions(depends_on=[deploy])
)

pulumi.export("projectId", project.id)
pulumi.export("serviceId", service.id)
pulumi.export("url", access.urls.apply(lambda urls: urls[0]))
