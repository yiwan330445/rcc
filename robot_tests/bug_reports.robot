*** Settings ***
Resource  resources.robot

*** Test cases ***

Github issue 7 about initial call with do-not-track
  [Setup]     Remove config    tmp/bug_7.yaml
  Wont Exist  tmp/bug_7.yaml

  Step        build/rcc configure identity --controller citests --do-not-track --config tmp/bug_7.yaml
  Must Have   anonymous health tracking is: disabled

Bug in virtual holotree with gzipped files
  [Tags]      WIP
  Step        build/rcc holotree blueprint --controller citests robot_tests/spellbug/conda.yaml
  Use STDERR
  Must Have   Blueprint "ef0163b57ff44cd5" is available: false

  Step        build/rcc run --liveonly --controller citests --robot robot_tests/spellbug/robot.yaml
  Use STDOUT
  Must Have   Bug fixed!

  Step        build/rcc holotree blueprint --controller citests robot_tests/spellbug/conda.yaml
  Use STDERR
  Must Have   Blueprint "ef0163b57ff44cd5" is available: false

  Step        build/rcc run --controller citests --robot robot_tests/spellbug/robot.yaml
  Use STDOUT
  Must Have   Bug fixed!

  Step        build/rcc holotree blueprint --controller citests robot_tests/spellbug/conda.yaml
  Use STDERR
  Must Have   Blueprint "ef0163b57ff44cd5" is available: true


*** Keywords ***

Remove Config
  [Arguments]  ${filename}
  Remove File  ${filename}

