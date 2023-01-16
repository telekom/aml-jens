# Build / Developer guide

## Building

### For Development-Use
Using `make` ( == `make go-build-dev` )

### For Release
Using `make go-build-release`

## Packaging
Using `make deb-package`.

The Version can be specified by setting VERSION: `VERSION=1.0 make deb-package`. It defaults to 99.99. 

## Developer Machine Setup

- Install Debian Bullseye without desktop on your machine.
- Install ssh server with root login enabled.
- Setup jens-cli with ansible from host machine connected (LAN) to jens-cli machine.
  - Install ansible, sshpass.
  - Adapt parameters ip, user and password in `inventory` to access the jens-cli machine.
  - In `ansible` folder execute `ansible-playbook -i inventory playbook_cli.yml`