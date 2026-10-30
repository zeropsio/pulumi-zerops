package io.zerops.examples;

import com.pulumi.Pulumi;
import com.pulumi.resources.CustomResourceOptions;
import io.zerops.zerops.Project;
import io.zerops.zerops.ProjectArgs;
import io.zerops.zerops.Service;
import io.zerops.zerops.ServiceArgs;
import io.zerops.zerops.ServiceDeploy;
import io.zerops.zerops.ServiceDeployArgs;
import io.zerops.zerops.ServiceEnv;
import io.zerops.zerops.ServiceEnvArgs;
import io.zerops.zerops.ServiceSubdomainAccess;
import io.zerops.zerops.ServiceSubdomainAccessArgs;
import java.io.IOException;
import java.io.UncheckedIOException;
import java.nio.file.Files;
import java.nio.file.Path;

public final class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var appName = config.require("appName");
            var message = config.get("message").orElse("Hello from Zerops");
            final String zeropsYaml;
            try {
                zeropsYaml = Files.readString(Path.of("app", "zerops.yml"));
            } catch (IOException error) {
                throw new UncheckedIOException("Cannot read app/zerops.yml", error);
            }

            var project = new Project("project", ProjectArgs.builder()
                    .name(appName).mode("LIGHT").build());
            var service = new Service("app", ServiceArgs.builder()
                    .projectId(project.id()).hostname("app").type("nodejs@22").build());
            var environment = new ServiceEnv("message", ServiceEnvArgs.builder()
                    .serviceId(service.id()).key("MESSAGE").content(message).build());
            var deployment = new ServiceDeploy("deploy", ServiceDeployArgs.builder()
                    .serviceId(service.id()).sourceDir("./app")
                    .zeropsYaml(zeropsYaml).setup("app").build(),
                    CustomResourceOptions.builder().dependsOn(environment).build());
            var access = new ServiceSubdomainAccess("access", ServiceSubdomainAccessArgs.builder()
                    .serviceId(service.id()).build(),
                    CustomResourceOptions.builder().dependsOn(deployment).build());

            ctx.export("projectId", project.id());
            ctx.export("serviceId", service.id());
            ctx.export("url", access.urls().applyValue(urls -> urls.get(0)));
        });
    }
}
