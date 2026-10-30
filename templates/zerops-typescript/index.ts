import * as pulumi from "@pulumi/pulumi";
import * as zerops from "@zerops/pulumi-app";

const config = new pulumi.Config();

// Deploy the app folder; Zerops builds it and provides a public URL.
const web = new zerops.WebApp("web", {
    appName: config.require("appName"),
    runtime: "python@3.12",
    source: "./app",
    // Enable PostgreSQL with: pulumi config set database true
    database: config.getBoolean("database") ?? false,
});

export const url = web.url;
export const projectId = web.projectId;
export const serviceId = web.serviceId;
