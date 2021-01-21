*** Settings ***
Resource  resources.robot

*** Test cases ***
Can operate leased environments

  Goal        Create environment with lease
  Step        build/rcc env variables --lease "taker (1)" robot_tests/leasebot/conda.yaml
  Must Have   CONDA_DEFAULT_ENV=rcc
  Must Have   RCC_ENVIRONMENT_HASH=8f1d3dc95228edef

  Goal        Check listing for taker information
  Step        build/rcc env list
  Must Have   "taker (1)"

  Goal        Others can get same environment
  Step        build/rcc env variables robot_tests/leasebot/conda.yaml
  Must Have   CONDA_DEFAULT_ENV=rcc
  Must Have   RCC_ENVIRONMENT_HASH=8f1d3dc95228edef

  Goal        Can share environment, but wont own the lease
  Step        build/rcc env variables --lease "second (2)" robot_tests/leasebot/conda.yaml
  Must Have   CONDA_DEFAULT_ENV=rcc
  Must Have   RCC_ENVIRONMENT_HASH=8f1d3dc95228edef

  Goal        Check listing for taker information (still same)
  Step        build/rcc env list
  Must Have   "taker (1)"
  Wont Have   "second (2)"

  Goal        Lets corrupt the environment
  Step        build/rcc task run -r robot_tests/leasebot/robot.yaml
  Must Have   Successfully installed pytz

  Goal        Now others cannot get same environment anymore
  Step        build/rcc env variables robot_tests/leasebot/conda.yaml   1
  Must Have   Environment leased to "taker (1)" is dirty
  Wont Have   RCC_ENVIRONMENT_HASH=8f1d3dc95228edef
  Wont Have   CONDA_DEFAULT_ENV=rcc

  Goal        Cannot share environment, since it is dirty
  Step        build/rcc env variables --lease "second (2)" robot_tests/leasebot/conda.yaml  1
  Must Have   Cannot get environment "8f1d3dc95228edef" because it is dirty and leased by "taker (1)"
  Wont Have   RCC_ENVIRONMENT_HASH=8f1d3dc95228edef
  Wont Have   CONDA_DEFAULT_ENV=rcc

  Goal        Cannot unlease someone elses environment
  Step        build/rcc env unlease --lease "second (2)" --hash 8f1d3dc95228edef  1
  Must Have   Error:

  Goal        Cannot delete someone elses leased environment
  Step        build/rcc env delete 8f1d3dc95228edef  1
  Must Have   WARNING: "8f1d3dc95228edef" is leased by "taker (1)" and wont be deleted!

  Goal        Check listing for taker information (still same)
  Step        build/rcc env list
  Must Have   8f1d3dc95228edef
  Must Have   "taker (1)"
  Wont Have   "second (2)"

  Goal        Lease can be unleased
  Step        build/rcc env unlease --lease "taker (1)" --hash 8f1d3dc95228edef
  Must Have   OK.

  Goal        Others can now lease that environment
  Step        build/rcc env variables --lease "second (2)" robot_tests/leasebot/conda.yaml
  Must Have   CONDA_DEFAULT_ENV=rcc
  Must Have   RCC_ENVIRONMENT_HASH=8f1d3dc95228edef

  Goal        Check listing for taker information (still same)
  Step        build/rcc env list
  Must Have   "second (2)"
  Wont Have   "taker (1)"

  Goal        Lease can be unleased
  Step        build/rcc env unlease --lease "second (2)" --hash 8f1d3dc95228edef
  Must Have   OK.
