#!/bin/bash

# Bootstraps a juju environment (from source) and deploys a few
# services in LXC containers. Uses 2 machines with different series.
# Waits for the deployment then collects all logs and relevant info
# from the remote machines and destroys the environment.

# Call with an environment name, its default series and another
# series as arguments. e.g. ./test-lxc.sh maas-hw trusty precise

# NOTE: All remote logs collected will be placed in the current
# directory.

JUJU_ENV="$1"
SERIES1="$2"
SERIES2="$3"

echo "Bootstrapping $JUJU_ENV with default-series: $SERIES1 using LXC containers"
# instance-type=m3.medium added to overcome the limitation of 4 private IPs per ENI on the default m1.small instance type.
# Once the template container is not getting a static IP (and thus wasting an IP effectively) remove this.
juju bootstrap -e $JUJU_ENV --upload-tools --debug --constraints "root-disk=20G instance-type=m3.medium"

echo "Deploying services"
juju deploy -e $JUJU_ENV cs:${SERIES1}/wordpress --to lxc:0
juju deploy -e $JUJU_ENV cs:${SERIES1}/mysql --to lxc:0
juju deploy -e $JUJU_ENV cs:${SERIES1}/ubuntu ubuntu1 --to lxc:0
juju deploy -e $JUJU_ENV cs:${SERIES2}/ubuntu ubuntu2 --to lxc:0

echo "Adding a second machine --series $SERIES2"
juju add-machine -e $JUJU_ENV --series $SERIES2

echo "Adding more units to the second machine"
juju add-unit -e $JUJU_ENV ubuntu1 --to lxc:1
juju add-unit -e $JUJU_ENV ubuntu2 --to lxc:1

echo "Adding relations"
juju add-relation -e $JUJU_ENV wordpress mysql

echo "Waiting for all machines to start..."
watch -n 5 juju status -e $JUJU_ENV --format tabular

echo "Getting logs from machine 0"
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/log/juju/all-machines.log" | tee -a all-machines.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/log/juju/machine-0.log" | tee -a machine-0.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-0.cloud-init-output.log
juju ssh -e $JUJU_ENV 0 -- "sudo iptables-save" | tee -a machine-0.iptables-save
juju ssh -e $JUJU_ENV 0 -- "sudo ip route list" | tee -a machine-0.ip-route-list
juju ssh -e $JUJU_ENV 0 -- "sudo ip addr list" | tee -a machine-0.ip-addr-list
juju ssh -e $JUJU_ENV 0 -- "sudo ip link show" | tee -a machine-0.ip-link-show

echo "Getting LXC template containers logs from machine 0"
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-${SERIES1}-lxc-template/console.log" | tee -a machine-0-${SERIES1}-lxc-template.console.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-${SERIES1}-lxc-template/container.log" | tee -a machine-0-${SERIES1}-lxc-template.container.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-${SERIES1}-lxc-template/cloud-init" | tee -a machine-0-${SERIES1}-lxc-template.cloud-init
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-${SERIES1}-lxc-template/lxc.conf" | tee -a machine-0-${SERIES1}-lxc-template.lxc.conf
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/lxc/juju-${SERIES1}-lxc-template/rootfs/var/log/cloud-init-output.log" | tee -a machine-0-${SERIES1}-lxc-template.cloud-init-output.log

juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-${SERIES2}-lxc-template/console.log" | tee -a machine-0-${SERIES2}-lxc-template.console.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-${SERIES2}-lxc-template/container.log" | tee -a machine-0-${SERIES2}-lxc-template.container.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-${SERIES2}-lxc-template/cloud-init" | tee -a machine-0-${SERIES2}-lxc-template.cloud-init
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-${SERIES2}-lxc-template/lxc.conf" | tee -a machine-0-${SERIES2}-lxc-template.lxc.conf
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/lxc/juju-${SERIES2}-lxc-template/rootfs/var/log/cloud-init-output.log" | tee -a machine-0-${SERIES2}-lxc-template.cloud-init-output.log

echo "Getting LXC remaining containers logs from machine 0"
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-0/console.log" | tee -a machine-0-lxc-0.console.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-0/container.log" | tee -a machine-0-lxc-0.container.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-0/cloud-init" | tee -a machine-0-lxc-0.cloud-init
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-0/lxc.conf" | tee -a machine-0-lxc-0.lxc.conf
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/lxc/juju-machine-0-lxc-0/rootfs/var/log/cloud-init-output.log" | tee -a machine-0-lxc-0.cloud-init-output.log

juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-1/console.log" | tee -a machine-0-lxc-1.console.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-1/container.log" | tee -a machine-0-lxc-1.container.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-1/cloud-init" | tee -a machine-0-lxc-1.cloud-init
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-1/lxc.conf" | tee -a machine-0-lxc-1.lxc.conf
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/lxc/juju-machine-0-lxc-1/rootfs/var/log/cloud-init-output.log" | tee -a machine-0-lxc-1.cloud-init-output.log

juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-2/console.log" | tee -a machine-0-lxc-2.console.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-2/container.log" | tee -a machine-0-lxc-2.container.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-2/cloud-init" | tee -a machine-0-lxc-2.cloud-init
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-2/lxc.conf" | tee -a machine-0-lxc-2.lxc.conf
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/lxc/juju-machine-0-lxc-2/rootfs/var/log/cloud-init-output.log" | tee -a machine-0-lxc-2.cloud-init-output.log

juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-3/console.log" | tee -a machine-0-lxc-3.console.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-3/container.log" | tee -a machine-0-lxc-3.container.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-3/cloud-init" | tee -a machine-0-lxc-3.cloud-init
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-lxc-3/lxc.conf" | tee -a machine-0-lxc-3.lxc.conf
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/lxc/juju-machine-0-lxc-3/rootfs/var/log/cloud-init-output.log" | tee -a machine-0-lxc-3.cloud-init-output.log

echo "Getting logs from machine 1"
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/log/juju/machine-1.log" | tee -a machine-1.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-1.cloud-init-output.log
juju ssh -e $JUJU_ENV 1 -- "sudo iptables-save" | tee -a machine-1.iptables-save
juju ssh -e $JUJU_ENV 1 -- "sudo ip route list" | tee -a machine-1.ip-route-list
juju ssh -e $JUJU_ENV 1 -- "sudo ip addr list" | tee -a machine-1.ip-addr-list
juju ssh -e $JUJU_ENV 1 -- "sudo ip link show" | tee -a machine-1.ip-link-show

echo "Getting LXC template containers logs from machine 1"
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-${SERIES1}-lxc-template/console.log" | tee -a machine-1-${SERIES1}-lxc-template.console.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-${SERIES1}-lxc-template/container.log" | tee -a machine-1-${SERIES1}-lxc-template.container.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-${SERIES1}-lxc-template/cloud-init" | tee -a machine-1-${SERIES1}-lxc-template.cloud-init
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-${SERIES1}-lxc-template/lxc.conf" | tee -a machine-1-${SERIES1}-lxc-template.lxc.conf
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/lxc/juju-${SERIES1}-lxc-template/rootfs/var/log/cloud-init-output.log" | tee -a machine-1-${SERIES1}-lxc-template.cloud-init-output.log

juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-${SERIES2}-lxc-template/console.log" | tee -a machine-1-${SERIES2}-lxc-template.console.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-${SERIES2}-lxc-template/container.log" | tee -a machine-1-${SERIES2}-lxc-template.container.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-${SERIES2}-lxc-template/cloud-init" | tee -a machine-1-${SERIES2}-lxc-template.cloud-init
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-${SERIES2}-lxc-template/lxc.conf" | tee -a machine-1-${SERIES2}-lxc-template.lxc.conf
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/lxc/juju-${SERIES2}-lxc-template/rootfs/var/log/cloud-init-output.log" | tee -a machine-1-${SERIES2}-lxc-template.cloud-init-output.log

echo "Getting LXC remaining containers logs from machine 1"
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-machine-1-lxc-0/console.log" | tee -a machine-1-lxc-0.console.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-machine-1-lxc-0/container.log" | tee -a machine-1-lxc-0.container.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-machine-1-lxc-0/cloud-init" | tee -a machine-1-lxc-0.cloud-init
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-machine-1-lxc-0/lxc.conf" | tee -a machine-1-lxc-0.lxc.conf
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/lxc/juju-machine-1-lxc-0/rootfs/var/log/cloud-init-output.log" | tee -a machine-1-lxc-0.cloud-init-output.log

juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-machine-1-lxc-1/console.log" | tee -a machine-1-lxc-1.console.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-machine-1-lxc-1/container.log" | tee -a machine-1-lxc-1.container.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-machine-1-lxc-1/cloud-init" | tee -a machine-1-lxc-1.cloud-init
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-machine-1-lxc-1/lxc.conf" | tee -a machine-1-lxc-1.lxc.conf
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/lxc/juju-machine-1-lxc-1/rootfs/var/log/cloud-init-output.log" | tee -a machine-1-lxc-1.cloud-init-output.log

echo "Destroying environment $JUJU_ENV"
juju destroy-environment $JUJU_ENV -y --force
