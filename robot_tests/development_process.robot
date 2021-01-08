*** Settings ***
Resource  resources.robot

*** Test cases ***

Has required changes in commit based on development process.

  Goal       See git changes for required files have been changed
  Step       git show --stat
  Must Have  docs/changelog.md
  Must Have  common/version.go

