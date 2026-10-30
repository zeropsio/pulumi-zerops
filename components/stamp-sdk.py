#!/usr/bin/env python3
"""Complete generated SDK release metadata and transitive primitive dependencies."""
import json
from pathlib import Path
import re
import shutil
import sys

language, version = sys.argv[1:]
root = Path('../sdk-app') / language
if language == 'go':
    target = root / 'zeropsapp'
    target.mkdir(exist_ok=True)
    for path in list(root.iterdir()):
        if path != target:
            if path.is_dir():
                shutil.copytree(path, target / path.name, dirs_exist_ok=True)
                shutil.rmtree(path)
            else:
                path.replace(target / path.name)
    (root.parent / 'go.mod').write_text(f'''module github.com/zeropsio/pulumi-zerops/sdk-app

go 1.26.6

require (
 github.com/pulumi/pulumi/sdk/v3 v3.273.0
 github.com/zeropsio/pulumi-zerops/sdk v{version}
)

replace github.com/zeropsio/pulumi-zerops/sdk => ../sdk
''')
    root = target
    path = root / 'internal/pulumiVersion.go'
    path.write_text(path.read_text().replace('semver.Version{}', f'semver.MustParse("{version}")'))
elif language == 'python':
    path = root / 'pyproject.toml'
    text = path.read_text().replace('name = "pulumi_zerops_app"', 'name = "pulumi-zerops-app"')
    text = re.sub(r'"pulumi[-_]zerops(?:[<>=!~][^"]*)?", ?', '', text)
    text = text.replace('dependencies = [', f'dependencies = ["pulumi-zerops=={version}", ')
    text = text.replace('  [project.license]\n    text = "MIT"', '  license = "MIT"')
    text = text.replace('"setuptools>=61.0"', '"setuptools>=77.0.3"')
    path.write_text(text)
    (root / "README.md").write_text((root / "pulumi_zerops_app/README.md").read_text())
elif language == 'nodejs':
    path = root / 'package.json'
    data = json.loads(path.read_text())
    data['dependencies'].pop('@pulumi/zerops', None)
    data['dependencies']['@zerops/pulumi'] = version
    path.write_text(json.dumps(data, indent=4) + '\n')
    for path in root.glob('*.ts'):
        path.write_text(path.read_text().replace('"@pulumi/zerops"', '"@zerops/pulumi"'))
elif language == 'dotnet':
    path = root / 'Zerops.Pulumi.App.csproj'
    text = path.read_text().replace('<Authors>', '<PackageId>Zerops.Pulumi.App</PackageId>\n    <Authors>', 1)
    text = re.sub(r'    <PackageReference Include="(?:Pulumi\.Zerops|Zerops\.Pulumi)"[^>]*/>\n', '', text)
    text = text.replace('<PackageReference Include="Pulumi"', f'<PackageReference Include="Zerops.Pulumi" Version="{version}" />\n    <PackageReference Include="Pulumi"')
    path.write_text(text)
    (root / 'version.txt').write_text(version + '\n')
elif language == 'java':
    path = root / 'build.gradle'
    text = path.read_text().replace('artifactId = "zerops-app"', 'artifactId = "pulumi-app"')
    text = text.replace('JavaLanguageVersion.of(11)', 'JavaLanguageVersion.of(21)')
    text += '\ntasks.withType(JavaCompile).configureEach { options.release = 11 }\n'
    text = re.sub(r'    implementation\("com\.pulumi:zerops:[^"]+"\)\n', '', text)
    text = text.replace('dependencies {', f'dependencies {{\n    api("io.zerops:pulumi:{version}")')
    text = text.replace('name = ""', 'name = "Zerops App"', 1)
    text = text.replace('name = ""', 'name = "MIT License"', 1)
    text = text.replace('url = ""', 'url = "https://opensource.org/licenses/MIT"', 1)
    text = text.replace('id = ""', 'id = "zeropsio"').replace('name = ""', 'name = "Zerops"').replace('email = ""', 'email = "account@zerops.io"')
    path.write_text(text)
    (root / 'version.txt').write_text(version + '\n')
for path in root.rglob('pulumi-plugin.json'):
    data = json.loads(path.read_text())
    data['version'] = version
    path.write_text(json.dumps(data, indent=4) + '\n')
