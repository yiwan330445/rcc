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

  Comment     Verify micromamba is installed or download and install it.
  Step        build/rcc ht vars robot_tests/conda.yaml
  Must Exist  %{ROBOCORP_HOME}/bin/
  Must Exist  %{ROBOCORP_HOME}/base/
  Must Exist  %{ROBOCORP_HOME}/live/
  Must Exist  %{ROBOCORP_HOME}/wheels/
  Must Exist  %{ROBOCORP_HOME}/pipcache/

Step
  [Arguments]  ${command}  ${expected}=0
  ${code}  ${output}  ${error}=  Run and return code output error  ${command}
  Set Suite Variable  ${robot_stdout}  ${output}
  Set Suite Variable  ${robot_stderr}  ${error}
  Use Stdout
  Log  <b>STDOUT</b><pre>${output}</pre>  html=yes
  Log  <b>STDERR</b><pre>${error}</pre>  html=yes
  Should be equal as strings  ${expected}  ${code}
  Wont Have   Failure:

Use Stdout
  Set Suite Variable  ${robot_output}  ${robot_stdout}

Use Stderr
  Set Suite Variable  ${robot_output}  ${robot_stderr}

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
