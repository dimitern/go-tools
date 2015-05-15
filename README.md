# go-tools
Collection of random tools I find useful, written (mostly) in Go.

## maas-utils
Using [gomaasapi](https://launchpad.net/gomaasapi), this command-line tool provides access to a running [MaaS](https://maas.ubuntu.com/) server. Supported sub-commands:
 - **list-ips** - display all statically allocated IP addresses.
 - **reserve-ip** - reserve a static IP address.
 - **release-ips** - release one or more statically allocated IP addresses.
 - **list-networks** - display all networks in MaaS.
 - **list-nics** - display all node group interfaces.

## juju test scripts
Exercising common networking scenarios.
