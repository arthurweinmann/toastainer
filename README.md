# Toastainer

Toastainer is a self - hosted platform to run and auto-scale serverless code instances. An instance is started by the first request and can then be joined by multiple other requests. HTTP, Websocket and SSH requests are supported. Joining the same running code instance with different kind of requests is supported. You can set the timeout and maximum number of joiners for each instance in order to autoscale them naturally. You can also make a request force join a particular instance with its ID. You can setup your own OS image to support any programming language.

# Motivation

At Toastate, we need a way to quickly deploy and run autoscaled code instances. We need those instance to be reachable by multiple request of distinct types. We could not vendor lock clients from our web agency. This is why we developed Toastainer> We believe it can also be useful for other projects and this is why we release it to the community. We also provide a hosted version at toastainer.toastate.com.

# Installation

## Requirements

TODO: checkout those requirements references for other non apt systems
With APT package manager: autoconf bison flex gcc g++ git libprotobuf-dev libnl-route-3-dev libtool make pkg-config btrfs-progs protobuf-compiler uidmap

## Ubuntu 20.04

- provide link to an image already done

## Standalone

## Multi - Nodes and Multi - Cloud

### IP Addresses

To maintain security, IP addresses of nodes must be in a private CIDR (10.0.0.0/8 or 172.16.0.0/12 or 192.168.0.0/16). This means they must be in the same private network or VPN. Toatscloud will throw an error if it is not the case.

## Installation from source

### Runner network setup

```bash

# if you do not have a non root user
groupadd -g 1000 toastainer --system
adduser --system --uid 1000 --gid 1000 --disabled-password --disabled-login toastainer
sudo -su toastainer
cd ~
mkdir toastainer

sudo -s

apt update && apt install git libprotobuf-dev libnl-route-3-dev libtool btrfs-progs protobuf-compiler uidmap

# The name tveth1 that will be used by the runner can be set in the configuration file
# It will be cloned and put inside Toaster's net linux namespace
ip link add dev tveth0 type veth peer name tveth1
ip link set dev tveth0 up
ip link set dev tveth1 up

# 10.0.0.0/16 is by convention a block of private addresses that we will use to attribute ip addresses to Toasters
# You may use another one as long as it is a conventional private address space as defined in http://www.faqs.org/rfcs/rfc1918.html
ip addr add 10.166.0.1/16 broadcast 10.166.255.255 dev tveth0

# If there are no forbidden ip addresses
iptables -t nat -A POSTROUTING -s 10.166.0.0/16 -j MASQUERADE

# Here we forbid toasters from connecting to an AWS VPC private ip addresses
iptables -t nat -A POSTROUTING -s 10.166.0.0/16 ! -d 172.16.0.0/12 -j MASQUERADE

# Here is another way to forbid some address spaces using a blackhole redirection
# See https://superuser.com/questions/1436913/what-is-ip-address-0-0-0-1-for-and-how-to-use-it/1436941 for address 0.0.0.1
# 169.254.169.254 is a special AWS IP used to retrieve metadata about the current EC2 instance - we forbid all 169.254. link local addresses for this reason>
iptables -t nat -N BLACKHOLE
iptables -t nat -A PREROUTING -s 10.166.0.0/16 -d 169.254.0.0/16,172.16.0.0/12,10.0.0.0/8,192.168.0.0/16,$LOCAL_SERVER_IP/32 -j BLACKHOLE -j BLACKHOLE
iptables -t nat -A BLACKHOLE -j DNAT --to-destination 0.0.0.1 # -j does not work anymore here with newer versions of iptables
iptables -t nat -A POSTROUTING -s 10.166.0.0/16 -j MASQUERADE

echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf

sysctl -p /etc/sysctl.conf

echo 1 > /proc/sys/net/ipv4/ip_forward

# we need an explicit name server ip address in resolv.conf for toasters to have access to internet
echo "nameserver 8.8.8.8" >> /etc/resolv.conf
```

Also the home folder and all its ancestors should be owned by toastainer user, for example ubuntu or uid/gid 1000

# Usage

# Full Configuration example