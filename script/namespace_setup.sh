#! /bin/bash

ip netns add host1
ip netns add host2

ip link add host1_veth0 type veth peer host2_veth0

ip link set host1_veth0 netns host1
ip link set host2_veth0 netns host2

ip netns exec host1 ip addr add 192.168.0.2/24 dev host1_veth0
ip netns exec host2 ip addr add 192.168.0.3/24 dev host2_veth0

ip netns exec host1 ip link set lo up
ip netns exec host2 ip link set lo up
ip netns exec host1 ip link set host1_veth0 up
ip netns exec host2 ip link set host2_veth0 up
