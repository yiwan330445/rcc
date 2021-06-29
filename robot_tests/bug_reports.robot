*** Settings ***
Resource  resources.robot

*** Test cases ***

Github issue 7 about initial call with do-not-track
  [Setup]     Remove config    tmp/bug_7.yaml
  Wont Exist  tmp/bug_7.yaml

  Step        build/rcc configure identity --controller citests --do-not-track --config tmp/bug_7.yaml
  Must Have   anonymous health tracking is: disabled

*** Keywords ***

Remove Config
  [Arguments]  ${filename}
  Remove File  ${filename}

