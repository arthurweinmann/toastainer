#!/usr/bin/env bash
set -euo pipefail

ip link add dev tveth0 type veth peer name tveth1
ip link set dev tveth0 up
ip link set dev tveth1 up

ip addr add 10.166.0.1/16 broadcast 10.166.255.255 dev tveth0

iptables -t nat -A POSTROUTING -s 10.166.0.0/16 -j MASQUERADE

echo 1 > /proc/sys/net/ipv4/ip_forward
echo "nameserver 8.8.8.8" >> /etc/resolv.conf