#!/usr/bin/env python3
"""Validate committed release versions and produce release notes."""
import json
from pathlib import Path
import re
import sys
import tomllib
import xml.etree.ElementTree as ET

root = Path(__file__).resolve().parent.parent
version = re.search(r'^VERSION\s*\?=\s*(\S+)', (root / 'Makefile').read_text(), re.M)[1]
if len(sys.argv) > 1 and sys.argv[1] != 'v' + version:
    raise SystemExit('Release tag does not match Makefile VERSION')
for sdk in ('sdk', 'sdk-app'):
    base = root / sdk
    versions = [json.loads((base / 'nodejs/package.json').read_text())['version'],
                tomllib.loads((base / 'python/pyproject.toml').read_text())['project']['version']]
    versions += [p.read_text().strip() for p in base.glob('*/version.txt')]
    versions += re.findall(r'\? "([0-9]+\.[0-9]+\.[0-9]+)"', (base / 'java/build.gradle').read_text())
    versions += [e.text for p in (base / 'dotnet').glob('*.csproj') for e in ET.parse(p).iter('Version')]
    versions += re.findall(r'semver.MustParse\("([^"]+)"\)', next((base / 'go').rglob('pulumiVersion.go')).read_text())
    versions += [json.loads(p.read_text())['version'] for p in base.rglob('pulumi-plugin.json')]
    if not versions or any(v != version for v in versions):
        raise SystemExit(f'{sdk}: committed versions disagree with {version}: {versions}')
module = (root / 'provider/go.mod').read_text()
def dependency(name):
    return re.search(r'^\s*' + re.escape(name) + r'\s+(\S+)', module, re.M)[1]
changelog = root / 'CHANGELOG.md'
notes = ''
if changelog.exists():
    sections = re.split(r'(?m)^## ', changelog.read_text())
    notes = next((s for s in sections if s.startswith(('[' + version + ']', version))), '')
print(f'# Zerops Pulumi {version}\n\n' + notes)
print('## Pinned toolchain\n')
for name in ('github.com/zeropsio/terraform-provider-zerops', 'github.com/pulumi/pulumi-terraform-bridge/v3'):
    print(f'- `{name}`: `{dependency(name)}`')
print('- Pulumi CLI: `3.273.0`')
print('\nMaven Central publishing uploads to the Central Portal compatibility staging API.\nA maintainer must review and publish each deployment in the Central Portal after upload.')
