*** Settings ***
Library  OperatingSystem
Library  supporting.py

*** Keywords ***

Clean Local
  Remove Directory  tmp/robocorp  True

Prepare Local
  Remove Directory  tmp/fluffy  True
  Remove Directory  tmp/nodogs  True
  Remove Directory  tmp/robocorp  True
  Remove File  tmp/nodogs.zip
  Create Directory  tmp/robocorp
  Set Environment Variable  ROBOCORP_HOME  tmp/robocorp

  Goal        Verify miniconda is installed or download and install it.
  Step        build/rcc conda check -i
  Must Have   OK.
  Must Exist  %{ROBOCORP_HOME}/miniconda3/
  Wont Exist  %{ROBOCORP_HOME}/base/
  Wont Exist  %{ROBOCORP_HOME}/live/
  Wont Exist  %{ROBOCORP_HOME}/wheels/
  Wont Exist  %{ROBOCORP_HOME}/pipcache/

Goal
  [Arguments]  ${anything}
  Comment      ${anything}

Step
  [Arguments]  ${command}  ${expected}=0
  ${code}  ${output}=  Run and return rc and output  ${command}
  Set Suite Variable  ${robot_output}  ${output}
  Log  <pre>${output}</pre>  html=yes
  Should be equal as strings  ${expected}  ${code}
  Wont Have   Failure:

Must Be
  [Arguments]  ${content}
  Should Be Equal As Strings  ${robot_output}  ${content}

Wont Be
  [Arguments]  ${content}
  Should Not Be Equal As Strings  ${robot_output}  ${content}

Must Have
  [Arguments]  ${content}
  Should Contain  ${robot_output}  ${content}

Wont Have
  [Arguments]  ${content}
  Should Not Contain  ${robot_output}  ${content}

Must Exist
  [Arguments]  ${filepath}
  Should Exist  ${filepath}

Wont Exist
  [Arguments]  ${filepath}
  Should Not Exist  ${filepath}

Must Be Json Response
  Parse JSON  ${robot_output}
