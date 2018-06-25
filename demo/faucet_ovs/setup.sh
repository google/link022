# Copyright 2018 Google Inc.
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

NS=gw
OUT_INTF=wlx00c0ca902f38  # Interface to connect to the Internet
CTRL_INTF=eth1  # Interface to connect to a controlling machine, ssh is enabled here.
GWIP=192.168.11.1


# Create veth pair, one side for ovs, the other side for the host itself
INTF_ON_OVS=ovs2${NS}
INTF=to_ovs # inside the host
sudo ip netns add ${NS}
sudo ip link add name ${INTF_ON_OVS} type veth peer name ${INTF} netns ${NS}

sudo ip link set dev ${INTF_ON_OVS} up
sudo ip netns exec ${NS} ip addr add ${GWIP}/24 dev ${INTF}
sudo ip netns exec ${NS} ip link set dev ${INTF} up
sudo ip netns exec ${NS} ip link set dev lo up

# Start DHCP
sudo ip netns exec ${NS} dnsmasq --no-ping -p 0 -k \
 -F set:s0,192.168.11.2,192.168.11.10 \
 -O tag:s0,3,192.168.11.1 -O option:dns-server,8.8.8.8  -I lo -z \
 -l /tmp/link022.leases -8 /tmp/link022.dhcp.log -i ${INTF} -a ${GWIP} --conf-file= &

########### Get Internet access for the NS
TO_DEF=to_def
TO_NS=def2${NS}

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
add_vlan auth 300 192.168.44

RADIUS_PATH=../radius/freeradius
sudo ip netns exec ${NS} freeradius -X -d ${RADIUS_PATH} > /tmp/${NS}_radius.log &

#############create ovs
sudo ovs-vsctl add-br br0 \
	-- set bridge br0 other-config:datapath-id=0000000000000001 \
	-- set bridge br0 other-config:disable-in-band=true \
	-- set bridge br0 fail_mode=secure \
	-- add-port br0 ${INTF_ON_OVS} -- set interface ${INTF_ON_OVS} ofport_request=1 \
	-- add-port br0 enp2s0 -- set interface enp2s0 ofport_request=2 \
	-- add-port br0 enp3s0 -- set interface enp3s0 ofport_request=3 \
	-- add-port br0 enp4s0 -- set interface enp4s0 ofport_request=4 \
	-- add-port br0 enp5s0 -- set interface enp5s0 ofport_request=5 \
	-- set-controller br0 tcp:127.0.0.1:6653 tcp:127.0.0.1:6654