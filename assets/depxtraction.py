#!/bin/env python3

import pip, re, sys

from collections import namedtuple
from importlib import metadata

REJECTED = {'pkg_resources', 'pkgutil_resolve_name', 'pip', 'setuptools', 'wheel'}

NAMEFORM = re.compile(r'^([a-z0-9](?:[a-z0-9._-]*?[a-z0-9])?)([^a-z0-9._-].*)?$', re.I)
EXTRAFORM = re.compile(r'\bextra\s*=')

Metadata = namedtuple('Metadata', 'key name version needs')

def normalize(name):
    return re.sub(r'[-_.]+', '-', name).lower()

def unify(name):
    return normalize(str(name).strip())

def list_modules():
    for candidate in metadata.distributions():
        yield candidate

def parselet(text):
    head, *rest = map(str.strip, text.split(';'))
    name, *ignore = map(str.strip, filter(bool, NAMEFORM.match(head).groups()))
    extra = any(map(EXTRAFORM.match, rest))
    return extra, unify(name)

def conda_yaml(resolved):
    print('channels:\n- conda-forge\ndependencies:')
    python = sys.version_info
    print(f'- python={python.major}.{python.minor}.{python.micro}')
    print(f'- pip={pip.__version__}')
    if version := resolved.pop('robocorp-truststore', None):
        print(f'- robocorp-truststore={version}')
    if resolved:
        print(f'- pip:')
        for name, version in resolved.items():
            print(f'  - {name}=={version}')

def process():
    metadata = dict()
    for module in list_modules():
        name = module.metadata.get('name')
        key = unify(name)
        metadata[key] = Metadata(key, name, module.version, module.requires or tuple())

    cyclic = set()
    toplevel = set(metadata.keys())
    tuple(map(toplevel.discard, REJECTED))
    for package, needs in sorted(metadata.items()):
        for entry in needs.needs:
            rejected, name = parselet(entry)
            if (package, name) in cyclic:
                continue
            if not rejected:
                cyclic.add((name, package))
                toplevel.discard(name)

    resolved = dict([metadata[x].name, metadata[x].version] for x in sorted(toplevel))
    conda_yaml(resolved)

if __name__ == '__main__':
    process()
