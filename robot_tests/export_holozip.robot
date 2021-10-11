*** Settings ***
Library  OperatingSystem
Library  supporting.py
Resource  resources.robot
Suite Setup  Export setup
Suite Teardown  Export teardown

*** Keywords ***
Export setup
  Remove Directory  tmp/developer  True
  Remove Directory  tmp/guest  True
  Remove Directory  tmp/standalone  True
  Set Environment Variable  ROBOCORP_HOME  tmp/developer

Export teardown
  Set Environment Variable  ROBOCORP_HOME  tmp/robocorp
  Remove Directory  tmp/developer  True
  Remove Directory  tmp/guest  True
  Remove Directory  tmp/standalone  True

*** Test cases ***

Goal: Create extended robot into tmp/standalone folder using force.
  Step        build/rcc robot init --controller citests -t extended -d tmp/standalone -f
  Use STDERR
  Must Have   OK.

Goal: Create environment for standalone robot
  Step        build/rcc ht vars -s author --controller citests -r tmp/standalone/robot.yaml
  Must Have   RCC_ENVIRONMENT_HASH=
  Must Have   RCC_INSTALLATION_ID=
  Must Have   4e67cd8d4_fcb4b859
  Use STDERR
  Must Have   Downloading micromamba
  Must Have   Progress: 01/13
  Must Have   Progress: 13/13

Goal: Must have author space visible
  Step        build/rcc ht ls
  Use STDERR
  Must Have   4e67cd8d4_fcb4b859
  Must Have   rcc.citests
  Must Have   author
  Must Have   55aacd3b136421fd
  Wont Have   guest

Goal: Show exportable environment list
  Step        build/rcc ht export
  Use STDERR
  Must Have   Selectable catalogs
  Must Have   - 55aacd3b136421fd
  Must Have   OK.

Goal: Export environment for standalone robot
  Step        build/rcc ht export -z tmp/standalone/hololib.zip 55aacd3b136421fd
  Use STDERR
  Wont Have   Selectable catalogs
  Must Have   OK.

Goal: Wrap the robot
  Step        build/rcc robot wrap -z tmp/full.zip -d tmp/standalone/
  Use STDERR
  Must Have   OK.

Goal: See contents of that robot
  Step        unzip -v tmp/full.zip
  Must Have   robot.yaml
  Must Have   conda.yaml
  Must Have   hololib.zip

Goal: Can delete author space
  Step        build/rcc ht delete 4e67cd8d4_fcb4b859
  Step        build/rcc ht ls
  Use STDERR
  Wont Have   4e67cd8d4_fcb4b859
  Wont Have   rcc.citests
  Wont Have   author
  Wont Have   55aacd3b136421fd
  Wont Have   guest

Goal: Can run as guest
  Set Environment Variable  ROBOCORP_HOME  tmp/guest
  Step        build/rcc task run --controller citests -s guest -r tmp/standalone/robot.yaml -t 'run example task'
  Use STDERR
  Wont Have   Downloading micromamba
  Must Have   OK.

Goal: No spaces created under guest
  Set Environment Variable  ROBOCORP_HOME  tmp/guest
  Step        build/rcc ht ls
  Use STDERR
  Wont Have   4e67cd8d4_fcb4b859
  Wont Have   rcc.citests
  Wont Have   author
  Wont Have   55aacd3b136421fd
  Wont Have   4e67cd8d4_559e19be
  Wont Have   guest

Goal: Space created under author for guest
  Set Environment Variable  ROBOCORP_HOME  tmp/developer
  Step        build/rcc ht ls
  Use STDERR
  Wont Have   4e67cd8d4_fcb4b859
  Wont Have   author
  Must Have   rcc.citests
  Must Have   55aacd3b136421fd
  Must Have   4e67cd8d4_aacf1552
  Must Have   guest
