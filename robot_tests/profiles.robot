*** Settings ***
Library  OperatingSystem
Library  supporting.py
Resource  resources.robot
Default Tags      WIP

*** Test cases ***

Goal: Initially there are no profiles
  Step        build/rcc configuration switch  2
  Use STDERR
  Must Have   No profiles found, you must first import some.

Goal: Can import profiles into rcc
  Step        build/rcc configuration import --filename robot_tests/profile_alpha.yaml
  Use STDERR
  Must Have   OK.

  Step        build/rcc configuration import --filename robot_tests/profile_beta.yaml
  Use STDERR
  Must Have   OK.

Goal: Can see imported profiles
  Step        build/rcc configuration switch
  Must Have   Alpha settings
  Must Have   Beta settings
  Must Have   Currently active profile is: default
  Use STDERR
  Must Have   OK.

Goal: Can see imported profiles as json
  Step        build/rcc configuration switch --json
  Must Be Json Response
  Must Have   "Alpha settings"
  Must Have   "Beta settings"

Goal: Can switch to Alpha profile
  Step        build/rcc configuration switch --profile alpha
  Use STDERR
  Must Have   OK.

  Step        build/rcc configuration switch
  Use STDOUT
  Must Have   Currently active profile is: Alpha
  Use STDERR
  Must Have   OK.

Goal: Can switch to Beta profile
  Step        build/rcc configuration switch --profile Beta
  Use STDERR
  Must Have   OK.

  Step        build/rcc configuration switch
  Use STDOUT
  Must Have   Currently active profile is: Beta
  Use STDERR
  Must Have   OK.

Goal: Can switch to no profile
  Step        build/rcc configuration switch --noprofile
  Use STDERR
  Must Have   OK.

  Step        build/rcc configuration switch
  Use STDOUT
  Must Have   Currently active profile is: default
  Use STDERR
  Must Have   OK.

Goal: Can export profile
  Step        build/rcc configuration export --profile Alpha --filename tmp/exported_alpha.yaml
  Use STDERR
  Must Have   OK.
