# Copyright 2017 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/bin/bash
# A utility script to start services on another machine, from which the operator
# can manage the link022 device.
# A typical setup is to create a peer to peer ethernet link between link022 and
# the service machine.
# Interface connected to the link022 host
INTF=eth0
NS=lk22
GWIP=192.168.11.1

sudo ip netns add ${NS}
sudo ip link set dev ${INTF} netns ${NS}
sudo ip netns exec ${NS} ip addr add ${GWIP}/24 dev ${INTF}
sudo ip netns exec ${NS} ip link set dev ${INTF} up

# Start DHCP
sudo ip netns exec ${NS} dnsmasq --no-ping -p 0 -k \
 -F set:s0,192.168.11.2,192.168.11.10 \
 -O tag:s0,3,192.168.11.1 -O option:dns-server,8.8.8.8  -I lo -z \
 -l /tmp/link022.leases -8 /tmp/link022.dhcp.log -i ${INTF} -a ${GWIP} --conf-file= &

# Get Internet access for link022
TO_DEF=to_def
TO_NS=to_${NS}
OUT_INTF=wlan0

# enable forwarding
sudo sysctl net.ipv4.ip_forward=1
sudo ip netns exec ${NS} sysctl net.ipv4.ip_forward=1

# create veth pair
sudo ip link add name ${TO_NS} type veth peer name ${TO_DEF} netns ${NS}
# configure interfaces and routes
sudo ip addr add 192.168.22.1/30 dev ${TO_NS}
sudo ip link set ${TO_NS} up
# sudo ip route add 192.168.22.0/30 dev ${TO_NS}
sudo ip netns exec ${NS} ip addr add 192.168.22.2/30 dev ${TO_DEF}
sudo ip netns exec ${NS} ip link set ${TO_DEF} up
sudo ip netns exec ${NS} ip route add default via 192.168.22.1
# NAT in LK22
sudo ip netns exec ${NS} iptables -t nat -F
sudo ip netns exec ${NS} iptables -t nat -A POSTROUTING -o ${TO_DEF} -j MASQUERADE
# NAT in default
sudo iptables -P FORWARD DROP
sudo iptables -F FORWARD
# Assuming the host does not have other NAT rules.
sudo iptables -t nat -F
sudo iptables -t nat -A POSTROUTING -s 192.168.22.0/30 -o ${OUT_INTF} -j MASQUERADE
sudo iptables -A FORWARD -i ${OUT_INTF} -o ${TO_NS} -j ACCEPT
sudo iptables -A FORWARD -i ${TO_NS} -o ${OUT_INTF} -j ACCEPT

########### Adding vlans
function add_vlan {
	vlan_name=$1
	vlan_id=$2
	vlan_net=$3
	vlan_gw=${vlan_net}.1
	sudo ip netns exec ${NS} ip link add link ${INTF} name ${vlan_name} type vlan id ${vlan_id}
	sudo ip netns exec ${NS} ip addr add ${vlan_gw}/24 dev ${vlan_name}
	sudo ip netns exec ${NS} ip link set dev ${vlan_name} up

	# Start DHCP
	sudo ip netns exec ${NS} dnsmasq --no-ping -p 0 -k \
	 -F set:s0,${vlan_net}.2,${vlan_net}.100 \
	 -O tag:s0,3,${vlan_gw} -O option:dns-server,8.8.8.8  -I lo -z \
	 -l /tmp/link022.${vlan_name}.leases -8 /tmp/link022.${vlan_name}.dhcp.log -i ${vlan_name} -a ${vlan_gw} --conf-file= &
}
add_vlan guest 200 192.168.33
