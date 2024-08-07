# rcc -- how to build it

## Tooling

Required tools are:

- golang for implementing the thing
- rake for automating building the thing
- robot for testing the thing
- zip to build template zipfiles

See also: developer/README.md and developer/setup.yaml

Internal requirements:

- can be seen from go.mod and go.sum files in toplevel directory

## Commands

- to see available tasks, use `rake -T`
- to build everything, use `rake build` command
- to run robot tests, use `rake robot` command
- note, that most of rake commands are build to be used in Github Actions

## Where to start reading code?

To get started with CLI, start from "cmd" directory, which contains commands
executed from CLI, each in separate file (plus additional support files).
From there, use your editors code navigation to get to actual underlying
functions.
