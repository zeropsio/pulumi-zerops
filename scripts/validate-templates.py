#!/usr/bin/env python3
"""Validate template manifests when the optional templates directory exists."""
from pathlib import Path
import yaml

for directory in sorted(Path('templates').glob('*/')):
    if directory.is_dir():
        manifest = directory / 'Pulumi.yaml'
        with manifest.open() as stream:
            data = yaml.safe_load(stream)
        if not isinstance(data, dict) or not data.get('name') or not data.get('runtime'):
            raise SystemExit(f'{manifest}: name and runtime are required')
