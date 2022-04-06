#!/bin/env python3

import glob
import re

NONCHAR_PATTERN = re.compile(r'[^.a-z-]+')
HEADING_PATTERN = re.compile(r'^\s*(#{1,3})\s+(.*?)\s*$')
CODE_PATTERN = re.compile(r'^\s*[`]{3}')

DOT = '.'
DASH = '-'
NEWLINE = '\n'

IGNORE_LIST = (
        'docs/changelog.md',
        'docs/toc.md',
        'docs/README.md',
        )

PRIORITY_LIST = (
        'docs/usecases.md',
        'docs/features.md',
        'docs/recipes.md',
        'docs/profile_configuration.md',
        'docs/environment-caching.md',
        )

def unify(value):
    low = str(value).lower()
    return DASH.join(filter(bool, NONCHAR_PATTERN.split(low))).replace('.', '')

class Toc:
    def __init__(self, title, baseurl):
        self.title = title
        self.baseurl = baseurl
        self.levels = [0]
        self.toc = [f'# {title}']

    def leveling(self, level):
        levelup = True
        while len(self.levels) > level:
            self.levels.pop()
        while len(self.levels) < level:
            self.levels.append(1)
            levelup = False
        if levelup:
            self.levels[-1] += 1

    def add(self, filename, level, title):
        self.leveling(level)
        numbering = DOT.join(map(str, self.levels))
        url = f'{self.baseurl}{filename}'
        prefix = '#' * level
        ref = unify(title)
        self.toc.append(f'#{prefix} {numbering} [{title}]({self.baseurl}{filename}#{ref})')

    def write(self, filename):
        with open(filename, 'w+') as sink:
            sink.write(NEWLINE.join(self.toc))

def headings(filename):
    inside = False
    with open(filename) as source:
        for line in source:
            if CODE_PATTERN.match(line):
                inside = not inside
            if inside:
                continue
            if found := HEADING_PATTERN.match(line):
                level, title = found.groups()
                yield filename, len(level), title

def process():
    toc = Toc("ToC for rcc documentation", "https://github.com/robocorp/rcc/blob/master/")
    documentation = list(glob.glob('docs/*.md'))
    for filename in PRIORITY_LIST:
        if filename in documentation:
            documentation.remove(filename)
        for filename, level, title in headings(filename):
            toc.add(filename, level, title)
    for filename in documentation:
        if filename in IGNORE_LIST:
            continue
        for filename, level, title in headings(filename):
            toc.add(filename, level, title)
    toc.write('docs/README.md')

if __name__ == '__main__':
    process()
