import * as fs from "node:fs";
import * as pulumi from "@pulumi/pulumi";
import * as zerops from "@zerops/pulumi";

const config = new pulumi.Config();
const project = new zerops.Project("project", {
    name: config.require("appName"),
    mode: "LIGHT",
});
const service = new zerops.Service("app", {
    projectId: project.id,
    hostname: "app",
    type: "nodejs@22",
    horizontalAutoscaling: { minContainers: 1, maxContainers: 1 },
    verticalAutoscaling: { minCpu: 1, maxCpu: 1, minRamGb: 0.25, maxRamGb: 0.25, minDiskGb: 1 },
});
const message = new zerops.ServiceEnv("message", {
    serviceId: service.id,
    key: "MESSAGE",
    content: config.get("message") ?? "Hello from Zerops",
});
const deploy = new zerops.ServiceDeploy("deploy", {
    serviceId: service.id,
    sourceDir: "./app",
    zeropsYaml: fs.readFileSync("./app/zerops.yml", "utf8"),
    setup: "app",
}, { dependsOn: [message] });
const access = new zerops.ServiceSubdomainAccess("access", {
    serviceId: service.id,
}, { dependsOn: [deploy] });

export const projectId = project.id;
export const serviceId = service.id;
export const url = access.urls.apply(urls => urls[0]);
