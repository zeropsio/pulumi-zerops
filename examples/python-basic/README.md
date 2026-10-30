# Python: deploy an app to Zerops

Prerequisites: Python 3.10+, Pulumi CLI, a Zerops account and API token.
The infrastructure program is Python; the application is a dependency-free Node.js
HTTP server deployed to a Node.js 22 service.

```sh
# From this directory. Use Pulumi Cloud, or choose a local file backend:
pulumi login                    # Pulumi Cloud
# pulumi login file://$HOME/pulumi-state
pulumi stack init dev
python3 -m venv venv
venv/bin/python -m pip install -r requirements.txt
pulumi config set --secret zerops:token YOUR_TOKEN
pulumi config set appName my-python-app
# Optional (default: Hello from Zerops):
pulumi config set message 'Hello from Python'
pulumi up
curl "$(pulumi stack output url)"
```

`appName` is the explicit Zerops project name. The program creates a LIGHT project,
an `app` runtime service, a `MESSAGE` environment variable, a source deployment,
and public subdomain access after deployment. It exports `projectId`, `serviceId`
and `url`. The deployment reads `app/zerops.yml` verbatim. Edit `app/server.js`
(for example, change `v1` to `v2`) and run `pulumi up` to redeploy changed sources.

When finished:

```sh
pulumi destroy
pulumi stack rm dev
```

## Local provider development

The requirements deliberately reference the published `pulumi-zerops==0.1.1`
package. For local SDK development, build the wheel from the repository root and
install it into this example's virtual environment:

```sh
make build_python
cd examples/python-basic
python3 -m venv venv
venv/bin/python -m pip install --find-links ../../sdk/python/dist -r requirements.txt
venv/bin/python -m pip install --force-reinstall --no-deps ../../sdk/python/dist/pulumi_zerops-0.1.1-py3-none-any.whl
venv/bin/python -m py_compile __main__.py
```

The credential-gated integration test reads `ZEROPS_TOKEN` and optional
`ZEROPS_ENDPOINT`/`ZEROPS_ORGANIZATION_ID` values from the environment. Run from
`examples/` with `go test -run '^TestPythonBasic$' -timeout 60m`.

