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
GWIP=10.11.11.1

sudo ip netns add ${NS}
sudo ip link set dev ${INTF} netns ${NS}
sudo ip netns exec ${NS} ip addr add ${GWIP}/24 dev ${INTF}
sudo ip netns exec ${NS} ip link set dev ${INTF} up

# Start DHCP
sudo ip netns exec ${NS} dnsmasq --no-ping -p 0 -k \
 -F set:s0,10.11.11.2,10.11.11.10 \
 -O tag:s0,3,10.11.11.1 -l /tmp/link022.leases -8 /tmp/link022.dhcp.log -i ${INTF} --conf-file= &
