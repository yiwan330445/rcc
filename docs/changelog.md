# rcc change log

## v9.19.1 (date: 8.6.2021)

- added locking of holotree into new environment building and recording

## v9.19.0 (date: 8.6.2021)

- added event journaling support (no user visible yet)
- added first event "space-used" in holotree restore operations (this enables
  tracking of all places where environments are created)

## v9.18.0 (date: 3.6.2021)

- now using holotree location from catalog, so that catalog decides where
  holotree is created (defaults to `ROBOCORP_HOME` but can be different)
- if hololib.zip exist, then `--space` flag must be given or run fails
- hololib.zip is now reported in robot diagnostics
- environment difference print is now (mostly) behind `--trace` flag
- if rcc is not interactive, color toggling on Windows is skipped
- micromamba download is now done "on demand" only
- added robot tests for hololib.zip workflow

## v9.17.2 (date: 2.6.2021)

- fixing broken tests, and taking account changed specifications

## v9.17.1 (date: 2.6.2021) broken

- adding supporting structures for zip based holotree runs [experimental]

## v9.17.0 (date: 26.5.2021)

- added `export` command to holotree [experimental]

## v9.16.0 (date: 21.5.2021)

- catalog extension based on operating system, architecture and directory
  location

## v9.15.1 (date: 21.5.2021)

- added images as non-executable files
- run and testrun commands have new option `--no-outputs` which prevent
  capture of stderr/stdout into files
- separated `--trace` and `--debug` flags from `micromamba` and `pip` verbosity
  introduced in v9.12.0 (it is causing too much output and should be reserved
  only for `RCC_VERBOSE_ENVIRONMENT_BUILDING` variable

## v9.15.0 (date: 20.5.2021)

- for `task run` and `task testrun` there is now possibility to give additional
  arguments from commandline, by using `--` separator between normal rcc
  arguments and those intended for executed robot
- rcc now considers "http://127.0.0.1" as special case that does not require
  https

## v9.14.0 (date: 19.5.2021)

- added PYTHONPATH diagnostics validation
- added `--production` flag to diagnostics commands

## v9.13.0 (date: 18.5.2021)

- micromamba upgrade to version 0.13.1
- activation script fix for windows environment

## v9.12.1 (date: 18.5.2021)

- new environment variable `ROBOCORP_OVERRIDE_SYSTEM_REQUIREMENTS` to make
  skip those system requirements that some users are willing to try
- first such thing is "long path support" on some versions of Windows

## v9.12.0 (date: 18.5.2021)

- new environment variable `RCC_VERBOSE_ENVIRONMENT_BUILDING` to make
  environment building more verbose
- with above variable and `--trace` or `--debug` flags, both micromamba
  and pip are run with more verbosity

## v9.11.3 (date: 12.5.2021)

- adding error signaling on anywork background workers
- more work on improving slow parts of holotree
- fixed settings.yaml conda link (conda.anaconda.org reference)

## v9.11.2 (date: 11.5.2021)

- added query cache in front of slow "has blueprint" query (windows)
- more timeline entries added for timing purposes

## v9.11.1 (date: 7.5.2021)

- new get/robot capabilitySet added into rcc
- added User-Agent to rcc web requests

## v9.11.0 (date: 6.5.2021)

- started using new capabilitySet feature of cloud authorization
- added metric for run/robot authorization usage
- one minor typo fix with "terminal" word

## v9.10.2 (date: 5.5.2021)

- added metrics to see when there was catalog failure (pre-check related)
- added PYTHONDONTWRITEBYTECODE=x setting into rcc generated environments,
  since this will pollute the cache (every compilation produces different file)
  without much of benefits
- also added PYTHONPYCACHEPREFIX to point into temporary folder
- added `--space` flag to `rcc cloud prepare` command

## v9.10.1 (date: 5.5.2021)

- added check for all components owned by catalog, to verify that they all
  are actually there
- added debug level logging on environment restoration operations
- added possibility to have line numbers on rcc produced log output (stderr)
- rcc log output (stderr) is now synchronized thru a channel
- made holotree command tree visible on toplevel listing

## v9.10.0 (date: 4.5.2021)

- refactoring code so that runs can be converted to holotree
- added `--space` option to runs so that they can use holotree
- holotree blueprint should now be unified form (same hash everywhere)
- holotree now co-exists with old implementation in backward compatible way

## v9.9.21 (date: 4.5.2021)

- documentation fix for toplevel config flag, closes #18

## v9.9.20 (date: 3.5.2021)

- added blueprint subcommand to holotree hierarchy to query blueprint
  existence in hololib

## v9.9.19 (date: 29.4.2021)

- refactoring to enable virtual holotree for --liveonly functionality
- NOTE: leased environments functionality will go away when holotree
  goes mainstream (and plan for that is rcc series v10)

## v9.9.18 (date: 28.4.2021)

- some cleanup on code base
- changed autoupdate url for Robocorp Lab

## v9.9.17 (date: 20.4.2021)

- added environment, workspace, and robot support to holotree variables command
- also added some robot tests for holotree to verify functionality

## v9.9.16 (date: 20.4.2021)

- added support for deleting holotree controller spaces
- added holotree and hololib to full environment cleanup
- added required parameter to `rcc env delete` command also

## v9.9.15 (date: 19.4.2021)

- bugfix: locking while multiple rcc are doing parallel work should now
  work better, and not corrupt configuration (so much)

## v9.9.14 (date: 15.4.2021)

- environment variables conda.yaml ordering fix (from robot.yaml first)
- task shell does not need task specified anymore

## v9.9.13 (date: 15.4.2021)

- fixing environment variables bug from below

## v9.9.12 (date: 15.4.2021)

- updated rpaframework to version 9.5.0 in templates
- added more timeline entries around holotree
- minor performance related changes for holotree
- removed default PYTHONPATH settings from "taskless" environment
- known, remaining bug: on "env variables" command, with robot without default
  task and without task given in CLI, environment wont have PATH or PYTHONPATH
  or robot details setup correctly

## v9.9.11 (date: 13.4.2021)

- added support for listing holotree controller spaces

## v9.9.10 (date: 12.4.2021)

- removed index.py utility, since better place is on other repo, and it
  was mistake to put it here

## v9.9.9 (date: 9.4.2021)

- fixed index.py utility tool to work in correct repository

## v9.9.8 (date: 9.4.2021)

- skip environment bootstrap when there is no conda.yaml used
- added index.py utility tool for generating index.html for S3

## v9.9.7 (date: 8.4.2021)

- now `rcc holotree bootstrap` can only download templates with `--quick`
  flag, or otherwise also prepare environment based on that template

## v9.9.6 (date: 8.4.2021)

- holotree note: in this series 9, holotree will remain experimental and
  will not be used for production yet
- added separate `holotree` subtree in command structure (it is not internal
  anymore, but still hidden)
- partial implementations of holotree variables and bootstrap commands
- settings.yaml version 2021.04 update: now there is separate section
  for templates
- profiling option `--pprof` is now global level option
- improved error message when rcc is not configured yet

## v9.9.5 (date: 6.4.2021)

- micromamba upgrade to version 0.9.2

## v9.9.4 (date: 6.4.2021)

- fix for holotree change detection when switching blueprints

## v9.9.3 (date: 1.4.2021)

- added export/SET prefix to `rcc env variables` command
- updated README.md with patterns to version numbered releases
- known bug: holotree does not work correctly yet -- DO NOT USE

## v9.9.2 (date: 1.4.2021)

- more holotree integration work to get it more experimentable

## v9.9.1 (date: 31.3.2021)

- Github Actions upgrade to use Go 1.16 for rcc compilation

## v9.9.0 (date: 31.3.2021) broken

- added holotree as part of source code (but not as integrated part yet)
- added new internal command: holotree

## v9.8.11 (date: 30.3.2021)

- added Accept header to micromamba download command
- made some URL diagnostics optional, if they are left empty

## v9.8.10 (date: 30.3.2021)

- fix: no more panics when directly writing to settings.yaml

## v9.8.9 (date: 29.3.2021)

- added `cloud-ui` to settings.yaml

## v9.8.8 (date: 29.3.2021)

- mixed fixes and experiments edition
- ignoring empty variable names on environment dumps, closes #17
- added some missing content types to web requests
- added experimental ephemeral ECC implementation
- more common timeline markers added
- will not list pip dependencies on assistant runs
- will not ask cloud for runtime authorization (bug fix)

## v9.8.7 (date: 26.3.2021)

- more finalization of settings.yaml change
- made micromamba less quiet on environment building
- secrets now have write access enabled in rcc authorization requests
- if merged conda.yaml files do not have names, merge result wont have either

## v9.8.6 (date: 25.3.2021)

- settings.yaml cleanup
- fixed robot tests for 9.8.5 template changes

## v9.8.5 (date: 24.3.2021)

- Robot templates updated: Rpaframework updated to v9.1.0
- Robot templates updated: Improved task names
- Robot templates updated: Extended template has example of multiple tasks execution

## v9.8.4 (date: 24.3.2021)

- fix for pip made too silent on this v9.8.x series
- and also in failure cases, print out full installation plan

## v9.8.3 (date: 24.3.2021)

- can configure all rcc operations not to verify correct SSL certificate
  (please note, doing this is insecure and allows man-in-the-middle attacks)
- applied reviewed changes to what is actually in settings.yaml file

## v9.8.2 (date: 23.3.2021)

- ALPHA level pre-release (do not use, unless you know what you are doing)
- reorganizing some code to allow better use of settings.yaml
- more values from settings.yaml are now used

## v9.8.1 (date: 22.3.2021)

- ALPHA level pre-release (do not use, unless you know what you are doing)
- now some parts of settings are used from settings.yaml
- settings.yaml is now critical part of rcc, so diagnostics also contains it
- also from now, problems in settings.yaml may make rcc to fail
- changed ephemeral key size to 2048, which should be good enough

## v9.8.0 (date: 18.3.2021)

- ALPHA level pre-release with settings.yaml (do not use, unless you know
  what you are doing)
- started to moved some of hardcoded things into settings.yaml (not used yet)
- minor assistant upload fix, where one error case was not marked as error

## v9.7.4 (date: 17.3.2021)

- typo fix pull request from jaukia
- added micromamba --no-rc flag

## v9.7.3 (date: 16.3.2021)

- upgrading micromamba dependency to 0.8.2 version
- added .robot, .csv, .yaml, .yml, and .json in non-executable fileset
- also added "dot" files as non-executable
- added timestamp update to copyfile functionality
- added toplevel --tag option to allow semantic tagging for client
  applications to indicate meaning of rcc execution call

## v9.7.2 (date: 11.3.2021)

- adding visibility of installation plans in environment listing
- added --json support to environment listing including installation plan file
- added command `rcc env plan` to show installation plans for environment
- installation plan is now also part of robot diagnostics, if available

## v9.7.1 (date: 10.3.2021)

- fixes/improvements to activation and installation plan
- added missing content type to assistant requests
- micromamba upgrade to 0.8.0

## v9.7.0 (date: 10.3.2021)

- conda environments are now activated once on creation, and variables go
  with environment, as `rcc_activate.json`
- there is also now new "installation plan" file inside environment, called
  `rcc_plan.log` which contains events that lead to activation
- normal runs are now more silent, since details are moved into "plan" file

## v9.6.2 (date: 5.3.2021)

- fix for time formats used in timeline, some metrics, and stopwatch

## v9.6.1 (date: 3.3.2021)

- refactored code use common.When as consistent timestamp for current rcc run

## v9.6.0 (date: 3.3.2021)

- new command `rcc cloud prepare` to support installing assistants on
  local computer for faster startup time
- added more timeline entries on relevant parts

## v9.5.4 (date: 2.3.2021)

- Updated rpaframework to version 7.6.0 in templates

## v9.5.3 (date: 2.3.2021)

- added `--interactive` flag to `rcc task run` command, so that developers
  can use debuggers and other interactive tools while debugging

## v9.5.2 (date: 25.2.2021)

- bug fix: now cloning sources are not removed during --liveonly action,
  even when that source seems to be invalid
- changed timeline to use percent (not permilles anymore)
- minor fix on env diff printout

## v9.5.1 (date: 25.2.2021)

- now also printing environment differences when live is dirty and base
  is not, just before restoring live from base

## v9.5.0 (date: 25.2.2021)

- added support for detecting environment corruption
- now dirhash command can be used to compare environment content

## v9.4.4 (date: 24.2.2021)

- fix: added panic protection to telemetry sending, this closes #13
- added initial support for execution timeline tracking

## v9.4.3 (date: 23.2.2021)

- added generic reading and parsing diagnostics for JSON and YAML files

## v9.4.2 (date: 23.2.2021)

- fix: marked --report flag required in issue reporting
- added account-email to issue report, as backup contact information

## v9.4.1 (date: 17.2.2021)

- added conda.yaml diagnostics (initial take)
- made `rcc env variables` to be not silent anymore
- log level changes in environment creation
- env creation workflow has now 6 steps, added identity visibility

## v9.4.0 (date: 17.2.2021)

- added initial robot diagnostics (just robot.yaml for now)
- integrated robot diagnostics into configuration diagnostics (optional)
- integrated robot diagnostics to issue reporting (optional)
- fix: windows paths were wrong; "bin" to "usr" change

## v9.3.12 (date: 17.2.2021)

- introduced 48 hour delay to recycling temp folders (since clients depend on
  having temp around after rcc process is gone); this closes #12

## v9.3.11 (date: 15.2.2021)

- micromamba upgrade to 0.7.14
- made process fail early and visibly, if micromamba download fails

## v9.3.10 (date: 11.2.2021)

- Windows automation made environments dirty by generating comtypes/gen
  folder. Fix is to ignore that folder.
- Added some more diagnostics information.

## v9.3.9 (date: 8.2.2021)

- micromamba cleanup bug fix (got error if micromamba is missing)
- micromamba download bug fix (killed on MacOS)

## v9.3.8 (date: 4.2.2021)

- making started and finished subprocess PIDs visible in --debug level.

## v9.3.7 (date: 4.2.2021)

- micromamba version printout changed, so rcc now parses new format
- micromamba is 0.x, so it does not follow semantic versioning yet, so
  rcc will now "lockstep" versions, with micromamba locked to 0.7.12 now

## v9.3.6 (date: 3.2.2021)

- removing "defaults" channel from robot templates

## v9.3.5 (date: 2.2.2021)

- micromamba upgrade to 0.7.12
- REGRESSION: `rcc task shell` got broken when micromamba was introduced,
  and this version fixes that

## v9.3.4 (date: 1.2.2021)

- fix: removing environments now uses rename first and then delete,
  to get around windows locked files issue
- warning: on windows, if environment is somehow locked by some process,
  this will fail earlier in the process (which is good thing), so be aware
- minor change on cache statistics representation and calculation

## v9.3.3 (date: 1.2.2021)

- adding `--dryrun` option to issue reporting

## v9.3.2 (date: 29.1.2021)

- added environment variables for installation identity, opt-out status as
  `RCC_INSTALLATION_ID` and `RCC_TRACKING_ALLOWED`

## v9.3.1 (date: 29.1.2021)

- fix: when environment is leased, temporary folder is will not be recycled
- cleanup command now cleans also temporary folders based on day limit

## v9.3.0 (date: 28.1.2021)

- support for applications to submit issue reports thru rcc
- print "robot.yaml" to logs, to make it visible for support cases
- diagnostics can now print into a file, and that is used as part
  of issue reporting
- added links to diagnostic checks, for user guidance

## v9.2.0 (date: 25.1.2021)

- experiment: carrier PoC

## v9.1.0 (date: 25.1.2021)

- new command `rcc configure diagnostics` to help identify environment
  related issues
- also requiring new version of micromamba, 0.7.10

## v9.0.2 (date: 21.1.2021)

- fix: prevent direct deletion of leased environment

## v9.0.1 (date: 20.1.2021)

- BREAKING CHANGES
- removal of legacy "package.yaml" support

## v9.0.0 (date: 18.1.2021)

- BREAKING CHANGES
- new cli option `--lease` to request longer lasting environment (1 hour from
  lease request, and next requests refresh the lease)
- new environment variable: `RCC_ENVIRONMENT_HASH` for clients to use
- new command `rcc env unlease` to stop leasing environments
- this breaks contract of pristine environments in cases where one application
  has already requested long living lease, and other wants to use environment
  with exactly same specification (if pristine, it is shared, otherwise it is
  an error)

## v8.0.12 (date: 18.1.2021)
- Templates conda -channel ordering reverted pending conda-forge chagnes.

## v8.0.10 (date: 18.1.2021)

- fix: when there is no pip dependencies, do not try to run pip command

## v8.0.9 (date: 15.1.2021)

- fix: removing one verbosity flag from micromamba invocation

## v8.0.8 (date: 15.1.2021)

- now micromamba 0.7.8 is required
- repodata TTL is reduced to 16 hours, and in case of environment creation
  failure, fall back to 0 seconds TTL (immediate update)
- using new --retry-with-clean-cache option in micromamba

## v8.0.7 (date: 11.1.2021)

- Now rcc manages TEMP and TMP locations for its subprocesses

## v8.0.6 (date: 8.1.2021)

- Updated to robot templates
- conda channels in order for `--strict-channel-priority`
- library versions updated and strict as well (rpaframework v7.1.1)
- Added basic guides for what to do in conda.yaml for end-users.

## v8.0.5 (date: 8.1.2021)

- added robot test to validate required changes, which are common/version.go
  and docs/changelog.md

## v8.0.4 (date: 8.1.2021)

- now requires micromamba 0.7.7 at least, with version check added
- micromamba now brings --repodata-ttl, which rcc currently sets for 7 days
- and touching conda caches is gone because of repodata ttl
- can now also cleanup micromamba binary and with --all
- environment validation checks simplified (no more separate space check)

## v8.0.3 (date: 7.1.2021)

- adding path validation warnings, since they became problem (with pip) now
  that we moved to use micromamba instead of miniconda
- also validation pattern update, with added "~" and "-" as valid characters
- validation is now done on toplevel, so all commands could generate
  those warnings (but currently they don't break anything yet)

## v8.0.2 (date: 5.1.2021)

- fixing failed robot tests for progress indicators (just tests)

## v8.0.1 (date: 5.1.2021)

- added separate pip install phase progress step (just visualization)
- now `rcc env cleanup` has option to remove miniconda3 installation

## v8.0.0 (date: 5.1.2021)

- BREAKING CHANGES
- removed miniconda3 download and installing
- removed all conda commands (check, download, and install)
- environment variables `CONDA_EXE` and `CONDA_PYTHON_EXE` are not available
  anymore (since we don't have conda installation anymore)
- adding micromamba download, installation, and usage functionality
- dropping 32-bit support from windows and linux, this is breaking change,
  so that is why version series goes up to v8

## v7.1.5 (date: 4.1.2021)

- now command `rcc man changelog` shows changelog.md from build moment

## v7.1.4 (date: 4.1.2021)

- bug fix for background metrics not send when application ends too fast
- now all telemetry sending happens in background and synchronized at the end
- added this new changelog.md file

## Older versions

Versions 7.1.3 and older do not have change log entries. This changelog.md
file was started at 4.1.2021.
