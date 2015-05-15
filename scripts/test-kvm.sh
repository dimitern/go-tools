#!/bin/bash

# Bootstraps a juju environment (from source) and deploys a few
# services in KVM containers. Uses 2 machines with different series.
# Waits for the deployment then collects all logs and relevant info
# from the remote machines and destroys the environment.

# Call with an environment name, its default series and another
# series as arguments. e.g. ./test-kvm.sh maas-hw trusty precise

# NOTE: All remote logs collected will be placed in the current
# directory.

JUJU_ENV="$1"
SERIES1="$2"
SERIES2="$3"

echo "Bootstrapping $JUJU_ENV with default-series: $SERIES1 using KVM containers"
juju bootstrap -e $JUJU_ENV --upload-tools --debug --constraints root-disk=20G

echo "Deploying services"
juju deploy -e $JUJU_ENV cs:${SERIES1}/wordpress --to kvm:0
juju deploy -e $JUJU_ENV cs:${SERIES1}/mysql --to kvm:0
juju deploy -e $JUJU_ENV cs:${SERIES1}/ubuntu ubuntu1 --to kvm:0
juju deploy -e $JUJU_ENV cs:${SERIES2}/ubuntu ubuntu2 --to kvm:0

echo "Adding a second machine --series $SERIES2"
juju add-machine -e $JUJU_ENV --series $SERIES2

echo "Adding more units to the second machine"
juju add-unit -e $JUJU_ENV ubuntu1 --to kvm:1
juju add-unit -e $JUJU_ENV ubuntu2 --to kvm:1

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

echo "Getting KVM containers cloud-init config from machine 0"
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-kvm-0/cloud-init" | tee -a machine-0-kvm-0.cloud-init
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-kvm-1/cloud-init" | tee -a machine-0-kvm-1.cloud-init
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-kvm-2/cloud-init" | tee -a machine-0-kvm-2.cloud-init
juju ssh -e $JUJU_ENV 0 -- "sudo cat /var/lib/juju/containers/juju-machine-0-kvm-3/cloud-init" | tee -a machine-0-kvm-3.cloud-init

echo "Getting logs from machine 0/kvm/0"
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo cat /var/log/juju/machine-0-kvm-0.log" | tee -a machine-0-kvm-0.log
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-0-kvm-0.cloud-init-output.log
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo iptables-save" | tee -a machine-0-kvm-0.iptables-save
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo ip route list" | tee -a machine-0-kvm-0.ip-route-list
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo ip addr list" | tee -a machine-0-kvm-0.ip-addr-list
juju ssh -e $JUJU_ENV 0/kvm/0 -- "sudo ip link show" | tee -a machine-0-kvm-0.ip-link-show

echo "Getting logs from machine 0/kvm/1"
juju ssh -e $JUJU_ENV 0/kvm/1 -- "sudo cat /var/log/juju/machine-0-kvm-1.log" | tee -a machine-0-kvm-1.log
juju ssh -e $JUJU_ENV 0/kvm/1 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-0-kvm-1.cloud-init-output.log
juju ssh -e $JUJU_ENV 0/kvm/1 -- "sudo iptables-save" | tee -a machine-0-kvm-1.iptables-save
juju ssh -e $JUJU_ENV 0/kvm/1 -- "sudo ip route list" | tee -a machine-0-kvm-1.ip-route-list
juju ssh -e $JUJU_ENV 0/kvm/1 -- "sudo ip addr list" | tee -a machine-0-kvm-1.ip-addr-list
juju ssh -e $JUJU_ENV 0/kvm/1 -- "sudo ip link show" | tee -a machine-0-kvm-1.ip-link-show

echo "Getting logs from machine 0/kvm/2"
juju ssh -e $JUJU_ENV 0/kvm/2 -- "sudo cat /var/log/juju/machine-0-kvm-2.log" | tee -a machine-0-kvm-2.log
juju ssh -e $JUJU_ENV 0/kvm/2 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-0-kvm-2.cloud-init-output.log
juju ssh -e $JUJU_ENV 0/kvm/2 -- "sudo iptables-save" | tee -a machine-0-kvm-2.iptables-save
juju ssh -e $JUJU_ENV 0/kvm/2 -- "sudo ip route list" | tee -a machine-0-kvm-2.ip-route-list
juju ssh -e $JUJU_ENV 0/kvm/2 -- "sudo ip addr list" | tee -a machine-0-kvm-2.ip-addr-list
juju ssh -e $JUJU_ENV 0/kvm/2 -- "sudo ip link show" | tee -a machine-0-kvm-2.ip-link-show

echo "Getting logs from machine 0/kvm/3"
juju ssh -e $JUJU_ENV 0/kvm/3 -- "sudo cat /var/log/juju/machine-0-kvm-3.log" | tee -a machine-0-kvm-3.log
juju ssh -e $JUJU_ENV 0/kvm/3 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-0-kvm-3.cloud-init-output.log
juju ssh -e $JUJU_ENV 0/kvm/3 -- "sudo iptables-save" | tee -a machine-0-kvm-3.iptables-save
juju ssh -e $JUJU_ENV 0/kvm/3 -- "sudo ip route list" | tee -a machine-0-kvm-3.ip-route-list
juju ssh -e $JUJU_ENV 0/kvm/3 -- "sudo ip addr list" | tee -a machine-0-kvm-3.ip-addr-list
juju ssh -e $JUJU_ENV 0/kvm/3 -- "sudo ip link show" | tee -a machine-0-kvm-3.ip-link-show

echo "Getting logs from machine 1"
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-1.cloud-init-output.log
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/log/juju/machine-1.log" | tee -a machine-1.log
juju ssh -e $JUJU_ENV 1 -- "sudo iptables-save" | tee -a machine-1.iptables-save
juju ssh -e $JUJU_ENV 1 -- "sudo ip route list" | tee -a machine-1.ip-route-list
juju ssh -e $JUJU_ENV 1 -- "sudo ip addr list" | tee -a machine-1.ip-addr-list
juju ssh -e $JUJU_ENV 1 -- "sudo ip link show" | tee -a machine-1.ip-link-show

echo "Getting KVM containers cloud-init config from machine 1"
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-machine-1-kvm-0/cloud-init" | tee -a machine-1-kvm-0.cloud-init
juju ssh -e $JUJU_ENV 1 -- "sudo cat /var/lib/juju/containers/juju-machine-1-kvm-1/cloud-init" | tee -a machine-1-kvm-1.cloud-init

echo "Getting logs from machine 1/kvm/0"
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo cat /var/log/juju/machine-1-kvm-0.log" | tee -a machine-1-kvm-0.log
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-1-kvm-0.cloud-init-output.log
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo iptables-save" | tee -a machine-1-kvm-0.iptables-save
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo ip route list" | tee -a machine-1-kvm-0.ip-route-list
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo ip addr list" | tee -a machine-1-kvm-0.ip-addr-list
juju ssh -e $JUJU_ENV 1/kvm/0 -- "sudo ip link show" | tee -a machine-1-kvm-0.ip-link-show

echo "Getting logs from machine 1/kvm/1"
juju ssh -e $JUJU_ENV 1/kvm/1 -- "sudo cat /var/log/juju/machine-1-kvm-1.log" | tee -a machine-1-kvm-1.log
juju ssh -e $JUJU_ENV 1/kvm/1 -- "sudo cat /var/log/cloud-init-output.log" | tee -a machine-1-kvm-1.cloud-init-output.log
juju ssh -e $JUJU_ENV 1/kvm/1 -- "sudo iptables-save" | tee -a machine-1-kvm-1.iptables-save
juju ssh -e $JUJU_ENV 1/kvm/1 -- "sudo ip route list" | tee -a machine-1-kvm-1.ip-route-list
juju ssh -e $JUJU_ENV 1/kvm/1 -- "sudo ip addr list" | tee -a machine-1-kvm-1.ip-addr-list
juju ssh -e $JUJU_ENV 1/kvm/1 -- "sudo ip link show" | tee -a machine-1-kvm-1.ip-link-show

echo "Destroying environment $JUJU_ENV"
juju destroy-environment $JUJU_ENV -y --force
