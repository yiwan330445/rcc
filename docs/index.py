#!/usr/bin/env python3

import pathlib
import re
import subprocess

LIMIT = 20

HEADER='''<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<title>Robocorp `rcc` downloads</title>
<style>
body { background-color: black; color: #eee; font-family: sans-serif; }
h1, h2, h3 { color: gold; }
a { color: orange; }
</style>
</head>
<body>
<h1>Robocorp `rcc` downloads</h1>
'''.strip()

TESTED_HEADER='''
<h2>Tested versions</h2>
<p>Consider these as more stable.</p>
'''.strip()

LATEST_HEADER='''
<hr />
<h2>Latest %(limit)d versions</h2>
'''.strip()


ENTRY='''
<h3>%(version)s</h3>
<p>Release date: %(when)s</p>
<ul>
<li>Windows: <a href="%(windows)s">%(windows)s</a></li>
<li>MacOS: <a href="%(macos)s">%(macos)s</a></li>
<li>Linux: <a href="%(linux)s">%(linux)s</a></li>
</ul>
'''.strip()

FOOTER='''
<hr />
</body>
</html>
'''.strip()

VERSION_PATTERN = re.compile(r'^##\s*(v[0-9.]+)\D+([0-9.]+)\D{1,5}$')
TAG_PATTERN = re.compile(r'^(v[0-9.]+)\D*$')

DIRECTORY = pathlib.Path(__file__).parent.absolute()
CHANGELOG = DIRECTORY.joinpath('changelog.md')
REPO_ROOT = DIRECTORY.parent.absolute()

TAGLISTING = f"git -C {REPO_ROOT} tag --list --sort='-taggerdate'"

def sh(command):
    task = subprocess.Popen([command], shell=True, stderr=subprocess.STDOUT, stdout=subprocess.PIPE)
    out, _ = task.communicate()
    return task.returncode, out.decode()

def gittags_top(count):
    code, out = sh(TAGLISTING)
    if code == 0:
        for line in out.splitlines():
            if count == 0:
                break
            if found := TAG_PATTERN.match(line):
                yield(found.group(1))
                count -= 1

def changelog_top(count):
    with open(CHANGELOG) as source:
        for line in source:
            if count == 0:
                break
            if found := VERSION_PATTERN.match(line):
                yield(found.groups())
                count -= 1

def download(version, suffix):
    return 'https://downloads.robocorp.com/rcc/releases/%s/%s' % (version, suffix)

def process_versions():
    biglist = tuple(changelog_top(10000))
    limited = biglist[:LIMIT] if len(biglist) > LIMIT else biglist

    daymap = dict()
    for version, when in biglist:
        daymap[version] = when

    print(TESTED_HEADER)
    for version in gittags_top(3):
        details = dict(version=version, when=daymap.get(version, 'N/A'))
        details['windows'] = download(version, 'windows64/rcc.exe')
        details['linux'] = download(version, 'linux64/rcc')
        details['macos'] = download(version, 'macos64/rcc')
        print(ENTRY % details)

    print(LATEST_HEADER % dict(limit=LIMIT))
    for version, when in limited:
        details = dict(version=version, when=when)
        details['windows'] = download(version, 'windows64/rcc.exe')
        details['linux'] = download(version, 'linux64/rcc')
        details['macos'] = download(version, 'macos64/rcc')
        print(ENTRY % details)

def process():
    print(HEADER % dict(limit=LIMIT))
    process_versions()
    print(FOOTER)

if __name__ == '__main__':
    process()
