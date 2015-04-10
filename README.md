# go-tools
Collection of random tools I find useful, written in Go.

## maas-utils
Using [https://launchpad.net/gomaasapi](gomaasapi), this command provides access to a running [https://maas.ubuntu.com/](MaaS) server. Supported sub-commands:
 - **list-ips** - display all statically allocated IP addresses.
 - **reserve-ip** - reserve a static IP address.
 - **release-ips** - release one or more statically allocated IP addresses.
 - **list-networks** - display all networks in MaaS.
 - **list-nics** - display all node group interfaces.

