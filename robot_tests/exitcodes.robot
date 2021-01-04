*** Settings ***
Library  OperatingSystem
Test template  Verify exitcodes

*** Test cases ***           EXITCODE    COMMAND

General failure of rcc command      1    build/rcc crapiti -h --controller citests

General output for rcc command      0    build/rcc --controller citests

Help for rcc command                0    build/rcc -h

Help for rcc assistant subcommand   0    build/rcc assistant -h --controller citests
Help for rcc cloud subcommand       0    build/rcc cloud -h --controller citests
Help for rcc configure subcommand   0    build/rcc configure -h --controller citests
Help for rcc env subcommand         0    build/rcc env -h --controller citests
Help for rcc feedback subcommand    0    build/rcc feedback -h --controller citests
Help for rcc help subcommand        0    build/rcc help -h --controller citests
Help for rcc man subcommand         0    build/rcc man -h --controller citests
Help for rcc internal subcommand    0    build/rcc internal -h --controller citests
Help for rcc robot subcommand       0    build/rcc robot -h --controller citests
Help for rcc task subcommand        0    build/rcc task -h --controller citests
Help for rcc version subcommand     0    build/rcc version -h --controller citests

Help for rcc assistant list         0    build/rcc assistant list -h --controller citests
Help for rcc assistant run          0    build/rcc assistant run -h --controller citests

Help for rcc cloud authorize        0    build/rcc cloud authorize -h --controller citests
Help for rcc cloud download         0    build/rcc cloud download -h --controller citests
Help for rcc cloud new              0    build/rcc cloud new -h --controller citests
Help for rcc cloud pull             0    build/rcc cloud pull -h --controller citests
Help for rcc cloud push             0    build/rcc cloud push -h --controller citests
Help for rcc cloud upload           0    build/rcc cloud upload -h --controller citests
Help for rcc cloud userinfo         0    build/rcc cloud userinfo -h --controller citests
Help for rcc cloud workspace        0    build/rcc cloud workspace -h --controller citests

Help for rcc configure credentials  0    build/rcc configure credentials -h --controller citests

Help for rcc env delete             0    build/rcc env delete -h --controller citests
Help for rcc env list               0    build/rcc env list -h --controller citests
Help for rcc env new                0    build/rcc env new -h --controller citests
Help for rcc env variables          0    build/rcc env variables -h --controller citests

Help for rcc configure identity     0    build/rcc configure identity -h --controller citests
Help for rcc feedback metric        0    build/rcc feedback metric -h --controller citests

Help for rcc man license            0    build/rcc man license -h --controller citests

Help for rcc robot fix              0    build/rcc robot fix -h --controller citests
Help for rcc robot initialize       0    build/rcc robot initialize -h --controller citests
Help for rcc robot libs             0    build/rcc robot libs -h --controller citests
Help for rcc robot list             0    build/rcc robot list -h --controller citests
Help for rcc robot unwrap           0    build/rcc robot unwrap -h --controller citests
Help for rcc robot wrap             0    build/rcc robot wrap -h --controller citests

Help for rcc task run               0    build/rcc task run -h --controller citests
Help for rcc task shell             0    build/rcc task shell -h --controller citests
Help for rcc task testrun           0    build/rcc task testrun -h --controller citests

*** Keywords ***

Verify exitcodes
  [Arguments]  ${exitcode}  ${command}
  ${code}  ${output}=  Run and return rc and output  ${command}
  Log  <pre>${output}</pre>  html=yes
  Should be equal as strings  ${exitcode}  ${code}
