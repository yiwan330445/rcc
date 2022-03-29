# Profile Configuration

## Quick start guide

```sh
# interactively create "Office" profile
rcc interactive configuration Office

# import that Office profile, so that it can be used
rcc configuration import --filename profile_office.yaml

# start using that Office profile
rcc configuration switch --profile Office

# verify that basic things work by doing diagnostics
rcc configuration diagnostics

# when basics work, see if full environment creation works
rcc configuration speedtest

# when you want to reset profile to "default" state
rcc configuration switch --noprofile

# if you want to export profile and deliver to others
rcc configuration export --profile Office --filename shared.yaml
```

## What is needed?

- you need rcc 11.9.7 or later
- your existing `settings.yaml` file (optional)
- your existing `micromambarc` file (optional)
- your existing `pip.ini` file (optional)
- your existing `cabundle.pem` file (optional)
- knowledge about your network proxies and certificate policies

## Discovery process

1. You must be inside that network that you are targetting the configuration.
2. Run interactive configuration and answer questions there.
3. Take created profile in use.
4. Run diagnostics and speed test to verify functionality.
5. Repeat these steps until everything works.
6. Export profile and share it with rest of your team/organization.
7. Create other profiles for different network locations (remote, VPN, ...)
