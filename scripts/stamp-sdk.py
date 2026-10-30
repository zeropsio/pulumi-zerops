#!/usr/bin/env python3
"""Apply release identities not exposed by Pulumi's language generators."""
import json
from pathlib import Path
import sys

language, version = sys.argv[1:]
root = Path("sdk") / language
if language == "python":
    path = root / "pyproject.toml"
    text = path.read_text().replace('name = "pulumi_zerops"', 'name = "pulumi-zerops"')
    import re
    text = re.sub(r'(?m)^  version = ".*"$', f'  version = "{version}"', text)
    text = text.replace('  [project.license]\n    text = "MIT"', '  license = "MIT"')
    text = text.replace('"setuptools>=61.0"', '"setuptools>=77.0.3"')
    path.write_text(text)
    (root / "README.md").write_text(Path("README.md").read_text())
elif language == "nodejs":
    path = root / "package.json"
    data = json.loads(path.read_text())
    data["version"] = version
    path.write_text(json.dumps(data, indent=4) + "\n")
elif language == "go":
    path = root / "zerops/internal/pulumiVersion.go"
    text = path.read_text().replace("semver.Version{}", f'semver.MustParse("{version}")')
    text = text.replace('var pluginDownloadURL string = ""', 'var pluginDownloadURL string = "github://api.github.com/zeropsio/pulumi-zerops"')
    path.write_text(text)
elif language == "dotnet":
    path = root / "Zerops.Pulumi.csproj"
    text = path.read_text().replace("<Authors>", f"<PackageId>Zerops.Pulumi</PackageId>\n    <Version>{version}</Version>\n    <Authors>", 1)
    path.write_text(text)
    (root / "version.txt").write_text(version + "\n")
elif language == "java":
    path = root / "build.gradle"
    text = path.read_text()
    import re
    text = re.sub(r'\? "[0-9]+\.[0-9]+\.[0-9]+"', f'? "{version}"', text)
    text = text.replace('artifactId = "zerops"', 'artifactId = "pulumi"')
    text = text.replace('JavaLanguageVersion.of(11)', 'JavaLanguageVersion.of(21)')
    text += '\ntasks.withType(JavaCompile).configureEach { options.release = 11 }\n'
    text = text.replace('name = ""', 'name = "Zerops"', 1)
    text = text.replace('name = ""', 'name = "MIT License"', 1)
    text = text.replace('url = ""', 'url = "https://opensource.org/licenses/MIT"', 1)
    text = text.replace('id = ""', 'id = "zeropsio"')
    text = text.replace('name = ""', 'name = "Zerops"')
    text = text.replace('email = ""', 'email = "account@zerops.io"')
    path.write_text(text)
    (root / "version.txt").write_text(version + "\n")

# Every emitted plugin descriptor must agree with the plugin binary release.
for path in root.rglob("pulumi-plugin.json"):
    data = json.loads(path.read_text())
    data["version"] = version
    path.write_text(json.dumps(data, indent=4) + "\n")
