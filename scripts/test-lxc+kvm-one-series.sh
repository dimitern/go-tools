#!/bin/bash

# Bootstraps a juju environment (from source) and deploys a few
# services in KVM containers. Waits for the deployment then collects
# all logs and relevant info from the remote machines and destroys the
# environment.

# Call with an environment name as argument. e.g.
# ./test-lxc+kvm-one-series.sh maas-hw

# NOTE: All remote logs collected will be placed in the current
# directory.

JUJU_ENV="$1"

echo "Bootstrapping $JUJU_ENV; using LXC and KVM containers"
juju bootstrap -e $JUJU_ENV --upload-tools --debug --constraints root-disk=20G

echo "Deploying services in a KVM and LXC containers"
juju deploy -e $JUJU_ENV wordpress --to lxc:0
juju deploy -e $JUJU_ENV mysql --to kvm:0
juju deploy -e $JUJU_ENV ubuntu --to lxc:0

echo "Adding a second machine"
juju add-machine -e $JUJU_ENV

echo "Adding more units to the second machine"
juju add-unit -e $JUJU_ENV ubuntu --to lxc:1
juju add-unit -e $JUJU_ENV mysql --to lxc:1
juju add-unit -e $JUJU_ENV wordpress --to kvm:1

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

echo "Getting the KVM container cloud-init config from machine 0"
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-kvm-0/cloud-init" | tee -a machine-0-kvm-0.cloud-init

echo "Getting LXC template containers logs from machine 0"
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-*-lxc-template/console.log" | tee -a machine-0-lxc-template.console.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-*-lxc-template/container.log" | tee -a machine-0-lxc-template.container.log
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-*-lxc-template/cloud-init" | tee -a machine-0-lxc-template.cloud-init
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-*-lxc-template/lxc.conf" | tee -a machine-0-lxc-template.lxc.conf
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/lxc/juju-*-lxc-template/rootfs/var/log/cloud-init-output.log" | tee -a machine-0-lxc-template.cloud-init-output.log

echo "Getting the remaining LXC containers logs from machine 0"
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

echo "Getting logs from machine 0/kvm/0"
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo cat /var/log/juju/machine-0-kvm-0.log" | tee -a machine-0-kvm-0.log
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-0-kvm-0.cloud-init-output.log
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo iptables-save" | tee -a machine-0-kvm-0.iptables-save
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo ip route list" | tee -a machine-0-kvm-0.ip-route-list
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo ip addr list" | tee -a machine-0-kvm-0.ip-addr-list
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo ip link show" | tee -a machine-0-kvm-0.ip-link-show

echo "Getting logs from machine 1"
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/log/juju/machine-1.log" | tee -a machine-1.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-1.cloud-init-output.log
juju ssh -e $JUJU_ENV 1 -- "sudo iptables-save" | tee -a machine-1.iptables-save
juju ssh -e $JUJU_ENV 1 -- "sudo ip route list" | tee -a machine-1.ip-route-list
juju ssh -e $JUJU_ENV 1 -- "sudo ip addr list" | tee -a machine-1.ip-addr-list
juju ssh -e $JUJU_ENV 1 -- "sudo ip link show" | tee -a machine-1.ip-link-show

echo "Getting the KVM container cloud-init config from machine 1"
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-machine-1-kvm-0/cloud-init" | tee -a machine-1-kvm-0.cloud-init

echo "Getting LXC template containers logs from machine 1"
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-*-lxc-template/console.log" | tee -a machine-1-lxc-template.console.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-*-lxc-template/container.log" | tee -a machine-1-lxc-template.container.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-*-lxc-template/cloud-init" | tee -a machine-1-lxc-template.cloud-init
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-*-lxc-template/lxc.conf" | tee -a machine-1-lxc-template.lxc.conf
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/lxc/juju-*-lxc-template/rootfs/var/log/cloud-init-output.log" | tee -a machine-1-lxc-template.cloud-init-output.log

echo "Getting the remaining LXC containers logs from machine 1"
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

echo "Getting logs from machine 1/kvm/0"
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo cat /var/log/juju/machine-1-kvm-0.log" | tee -a machine-1-kvm-0.log
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-1-kvm-0.cloud-init-output.log
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo iptables-save" | tee -a machine-1-kvm-0.iptables-save
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo ip route list" | tee -a machine-1-kvm-0.ip-route-list
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo ip addr list" | tee -a machine-1-kvm-0.ip-addr-list
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo ip link show" | tee -a machine-1-kvm-0.ip-link-show

echo "Destroying environment $JUJU_ENV"
juju destroy-environment $JUJU_ENV -y --force
