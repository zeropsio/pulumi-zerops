package io.zerops.examples;

import com.pulumi.Pulumi;
import io.zerops.zeropsapp.WebApp;
import io.zerops.zeropsapp.WebAppArgs;

public final class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            // WebApp creates the project, runtime, deployment and public URL.
            // Set database=true to add PostgreSQL and DATABASE_URL automatically.
            var web = new WebApp("web", WebAppArgs.builder()
                    .appName(config.require("appName"))
                    .runtime("python@3.12")
                    .source("./app")
                    .database(config.getBoolean("database").orElse(false))
                    .build());
            ctx.export("url", web.url());
            ctx.export("projectId", web.projectId());
            ctx.export("serviceId", web.serviceId());
        });
    }
}
