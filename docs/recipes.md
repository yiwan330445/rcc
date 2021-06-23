# Tips, tricks, and recipies

## How to see dependency changes?

Since version 10.2.2, rcc can show dependency listings using
`rcc robot dependencies` command. Listing always have two sided, "Wanted"
which is content from dependencies.yaml file, and "Available" which is from
actual environment command was run against. Listing is also shown during
robot runs.

### Why is this important?

- as time passes and world moves forward, new version of used components
  (dependencies) are released, and this may cause "configuration drift" on
  your robots, and without tooling in place, this drift might go unnoticed
- if your dependencies are not fixed, there will be configuration drift and
  your robot may change behaviour (become buggy) when dependency changes and
  goes against implemented robot
- even if you fix your dependencies in `conda.yaml`, some of those components
  or their components might have floating dependencies and they change your
  robots behaviour
- if your execution environment is different from your development environment
  then there might be different versions available for different operating
  systems
- if dependency resolution algorithm changes (pip for example) then you might
  get different environment with same `conda.yaml`
- when you upgrade one of your dependencies (for example, rpaframework) to new
  version, dependency resolution will now change, and now listing helps you
  understand what has changed and how you need to change your robot
  implementation because of that

### Example of dependencies listing from holotree environment

```sh
# first list dependencies from execution environment
rcc robot dependencies --space user

# if everything looks good, copy it as wanted dependencies.yaml
rcc robot dependencies --space user --copy

# and verify that everything looks `Same`
rcc robot dependencies --space user
```

## How pass arguments to robot from CLI?

Since version 9.15.0, rcc supports passing arguments from CLI to underlying
robot. For that, you need to have task in `robot.yaml` that co-operates with
additional arguments appended at the end of given `shell` command.

### Example robot.yaml with scripting task

```yaml
tasks:
  Run all tasks:
    shell: python -m robot --report NONE --outputdir output --logtitle "Task log" tasks.robot

  scripting:
    shell: python -m robot --report NONE --outputdir output --logtitle "Scripting log"

condaConfigFile: conda.yaml
artifactsDir: output
PATH:
  - .
PYTHONPATH:
  - .
ignoreFiles:
  - .gitignore
```

### Run it with `--` separator.

```sh
rcc task run --interactive --task scripting -- --loglevel TRACE --variable answer:42 tasks.robot
```

## How to run any command inside robot environment?

Since version 9.20.0, rcc now supports running any command inside robot space
using `rcc task script` command.

### Some example commands

Run following commands in same direcotry where your `robot.yaml` is. Or
otherwise you have to provide `--robot path/to/robot.yaml` in commandline.

```sh
# what python version we are running
rcc task script --silent -- python --version

# get pip list from this environment
rcc task script --silent -- pip list

# start interactive ipython session
rcc task script --interactive -- ipython
```

## Where can I find updates for rcc?

https://downloads.robocorp.com/rcc/releases/index.html

That is rcc download site with two categories of:
- tested versions (these are ones we ship with our tools)
- latest 20 versions (which are not battle tested yet, but are bleeding edge)

## What has changed on rcc?

### See changelog from git repo ...

https://github.com/robocorp/rcc/blob/master/docs/changelog.md

### See that from your version of rcc directly ...

```sh
rcc docs changelog
```
