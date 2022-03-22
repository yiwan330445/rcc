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

# if everything looks good, export it as wanted dependencies.yaml
rcc robot dependencies --space user --export

# and verify that everything looks `Same`
rcc robot dependencies --space user
```

## How to freeze dependencies?

Starting from rcc 10.3.2, there is now possibility to freeze dependencies.
This is how you can experiment with it.

### Steps

- have your `conda.yaml` to contain only those dependencies that your robot
  needs, either with exact versions or floating ones
- run robot in your target environment at least once, so that environment
  there gets created
- from that run's artifact directory, you should find file that has name
  something like `environment_xxx_yyy_freeze.yaml`
- copy that file back into your robot, right beside existing `conda.yaml`
  file (but do not overwrite it, you need that later)
- edit your `robot.yaml` file to point `condaConfigFile` entry to your
  newly created `environment_xxx_yyy_freeze.yaml` file
- repackage your robot and now your environment should stay quite frozen

### Limitations

- this is new and experimental feature, and we don't know yet how well it
  works in all cases (but we love to get feedback)
- currently this freezing limits where robot can be run, since dependencies
  on different operating systems and architectures differ and freezing cannot
  be done in OS and architecture neutral way
- your robot will break, if some specific package is removed from pypi or
  conda repositories
- your robot might also break, if someone updates package (and it's dependencies)
  without changing its version number
- for better visibility on configuration drift, you should also have
  `dependencies.yaml` inside your robot (see other recipe for it)

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

## Is rcc limited to Python and Robot Framework?

Absolutely not! Here is something completely different for you to think about.

Lets assume, that you are in almost empty Linux machine, and you have to
quickly build new micromamba in that machine. Hey, there is `bash`, `$EDITOR`,
and `curl` here.  But there are no compilers, git, or even python installed.

> Pop quiz, hot shot! Who you gonna call? MacGyver!

### This is what we are going to do ...

Here is set of commands we are going to execute in our trusty shell

```sh
mkdir -p builder/bin
cd builder
$EDITOR robot.yaml
$EDITOR conda.yaml
$EDITOR bin/builder.sh
curl -o rcc https://downloads.robocorp.com/rcc/releases/v10.3.2/linux64/rcc
chmod 755 rcc
./rcc run -s MacGyver
```

### Write a robot.yaml

So, for this to be a robot, we need to write heart of our robot, which is
`robot.yaml` of course.

```yaml
tasks:
  µmamba:
    shell: builder.sh
condaConfigFile: conda.yaml
artifactsDir: output
PATH:
- bin
```

### Write a conda.yaml

Next, we need to define what our robot needs to be able to do our mighty task.
This goes into `conda.yaml` file.

```yaml
channels:
- conda-forge
dependencies:
- git
- gmock
- cli11
- cmake
- compilers
- cxx-compiler
- pybind11
- libsolv
- libarchive
- libcurl
- gtest
- nlohmann_json
- cpp-filesystem
- yaml-cpp
- reproc-cpp
- python=3.8
- pip=20.1
```

### Write a bin/builder.sh

And finally, what does our robot do. And this time, this goes to our directory
bin, which is on our PATH, and name for this "robot" is actually `builder.sh`
and it is a bash script.

```sh
#!/bin/bash -ex

rm -rf target output/micromamba*
git clone https://github.com/mamba-org/mamba.git target
pushd target
version=$(git tag -l --sort='-creatordate' | head -1)
git checkout $version
mkdir -p build
pushd build
cmake .. -DCMAKE_INSTALL_PREFIX=/tmp/mamba -DENABLE_TESTS=ON -DBUILD_EXE=ON -DBUILD_BINDINGS=OFF
make
popd
popd
mkdir -p output
cp target/build/micromamba output/micromamba-$version
```

## How to control holotree environments?

There is three controlling factors for where holotree spaces are created.

First is location of `ROBOCORP_HOME` at creation time of environment. This
decides general location for environment and it cannot be changed or relocated
afterwards.

Second controlling factor is given using `--controller` option and default for
this is value `user`. And when applications are calling rcc, they should
have their own "controller" identity, so that all spaces created for one
application are groupped together by prefix of their "space" identity name.

Third controlling factor is content of `--space` option and again default
value there is `user`. Here it is up to user or application to decide their
strategy of use of different names to separate environments to their logical
used partitions. If you choose to use just defaults (user/user) then there
is going to be only one real environment available.

But above three controls gives you good ways to control how you and your
applications manage their usage of different python environments for
different purposes. You can share environments if you want, but you can also
give dedicates space for thos things that need full control of their space.

So running following commands demonstrate different levels of control for
space creation.

```
export ROBOCORP_HOME=/tmp/rchome
rcc holotree variables simple.yaml
rcc holotree variables --space tips simple.yaml
rcc holotree variables --controller tricks --space tips simple.yaml
```

If you now run `rcc holotree list` it should list something like following.

```
Identity            Controller  Space  Blueprint         Full path
--------            ----------  -----  --------          ---------
5a1fac3c5_2daaa295  rcc.user    tips   c34ed96c2d8a459a  /tmp/rchome/holotree/5a1fac3c5_2daaa295
5a1fac3c5_9fcd2534  rcc.user    user   c34ed96c2d8a459a  /tmp/rchome/holotree/5a1fac3c5_9fcd2534
9e7018022_2daaa295  rcc.tricks  tips   c34ed96c2d8a459a  /tmp/rchome/holotree/9e7018022_2daaa295
```

### How to get understanding on holotree?

See: https://github.com/robocorp/rcc/blob/master/docs/environment-caching.md

### How to activate holotree environment?

On Linux/MacOSX:

```
source <(rcc holotree variables --space mine --robot path/to/robot.yaml)
```

On Windows

```
rcc holotree variables --space mine --robot path/to/robot.yaml > mine_activate.bat
call mine_activate.bat
```

You can also try

```
rcc task shell --robot path/to/robot.yaml
```

## What can be controlled using environment variables?

- `ROBOCORP_HOME` points to directory where rcc keeps most of Robocorp related
  files and directories are kept
- `ROBOCORP_OVERRIDE_SYSTEM_REQUIREMENTS` makes rcc more relaxed on system
  requirements (like long path support requirement on Windows) but it also
  means that if set, responsibility of resolving failures are on user side
- `RCC_VERBOSE_ENVIRONMENT_BUILDING` makes environment creation more verbose,
  so that failing environment creation can be seen with more details
- `RCC_CREDENTIALS_ID` is way to provide Control Room credentials using
  environment variables

## How to troubleshoot rcc setup and robots?

```sh
# to get generic setup diagnostics
rcc configure diagnostics

# to get robot and environment setup diagnostics
rcc configure diagnostics --robot path/to/robot.yaml

# to see how well rcc performs in your machine
rcc configure speedtest
```

### Additional debugging options

- generic flag `--debug` shows debug messages during execution
- generic flag `--trace` shows more verbose debugging messages during execution
- flag `--timeline` can be used to see execution timeline and where time was spent

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

## Can I see these tips as web page?

Sure. See following URL.

https://github.com/robocorp/rcc/blob/master/docs/recipes.md

