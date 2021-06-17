# Tips, tricks, and recipies

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
