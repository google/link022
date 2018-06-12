# Link022 emulator
This doc contains steps to run a Link022 emulator.

The emulator builds a local testing environment with Link022 agent running inside a mininet node.

## Download Link022 repository
Download the entire [repository](../).

## Setup environment
The setup needs Python 2.7 environment and some additional packages.
```
apt-get install python python-netaddr mininet
```

## Compile Link022 agent
Run the [build script](../build.sh) to compile the Link022 agent.
It stores output binary file in the "binary" folder.
```
./build.sh
```

## Start emulator
Run the following command to start the emulator:
```
cd emulator

sudo env -u TERM python emulator.py \
  --target_cmd "../binary/link022_agent -ca ../demo/cert/server/ca.crt \
  -cert ../demo/cert/server/server.crt \
  -key ../demo/cert/server/server.key -eth_intf_name target-eth1 \
  -wlan_intf_name target-eth2"
```

The mininet CLI should appear after emulator started.
```
mininet>
```

## Verify the setup

### Check mininet nodes
```
mininet> nodes
available nodes are: 
ctrlr dummy target
```

There are three nodes in mininet.
* ctrlr: where gNMI client runs on.
* target: where Link022 agent runs on.
* dummy: a dummy host to contain the emulated eth and wlan interfaces.

### Check Link022 agent

The "target" node contains the following interfaces:
```
mininet> target ifconfig
lo: flags=73<UP,LOOPBACK,RUNNING>  mtu 65536
        inet 127.0.0.1  netmask 255.0.0.0
        inet6 ::1  prefixlen 128  scopeid 0x10<host>
        loop  txqueuelen 1  (Local Loopback)
        RX packets 0  bytes 0 (0.0 B)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 0  bytes 0 (0.0 B)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

target-eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 10.0.0.1  netmask 255.0.0.0  broadcast 10.255.255.255
        inet6 fe80::7cfb:bff:fe9c:7355  prefixlen 64  scopeid 0x20<link>
        ether 7e:fb:0b:9c:73:55  txqueuelen 1000  (Ethernet)
        RX packets 6  bytes 508 (508.0 B)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 5  bytes 418 (418.0 B)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

target-eth1: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet6 fe80::38b1:64ff:fee8:ef6c  prefixlen 64  scopeid 0x20<link>
        ether 3a:b1:64:e8:ef:6c  txqueuelen 1000  (Ethernet)
        RX packets 6  bytes 508 (508.0 B)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 6  bytes 508 (508.0 B)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

target-eth2: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet6 fe80::1490:40ff:fe65:1c4  prefixlen 64  scopeid 0x20<link>
        ether 16:90:40:65:01:c4  txqueuelen 1000  (Ethernet)
        RX packets 6  bytes 508 (508.0 B)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 6  bytes 508 (508.0 B)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
```
"target-eth0" is the management interface of the emulated AP, where gNMI target listens on (default port 8080).

The link022 agent process runs inside "target" node:
```
mininet> target ps aux | grep ../binary/link022_agent
root      28499  0.0  0.0 777732 14664 ?        Ssl  16:08   0:00 ../binary/link022_agent -ca ../demo/cert/server/ca.crt -cert ../demo/cert/server/server.crt -key ../demo/cert/server/server.key -eth_intf_name target-eth1 -wlan_intf_name target-eth2
```

The link022 log is "/tmp/link022_agent.INFO" by default.

## Config Link022 AP
All gNMI requests working on a physical Link022 AP should also work on the emulated one.
Run gNMI client in "ctrlr" node.

Here are some examples to start with:

1. Download GNMI clients.
* Download and compile [gNXI "SET" client](https://github.com/google/gnxi/tree/master/gnmi_set).
* Download and compile [gNXI "GET" client](https://github.com/google/gnxi/tree/master/gnmi_get).

2. Pushing the entire configuration to AP. It wipes out the existing configuration and applies the incoming one.
```
mininet> xterm ctrlr
{path to gnmi_set binary} \
-ca ../demo/cert/client/ca.crt \
-cert ../demo/cert/client/client.crt \
-key ../demo/cert/client/client.key \
-target_name www.example.com \
-target_addr 10.0.0.1:8080 \
-replace=/:@../tests/ap_config.json
```

3. Fetch AP configuration
```
mininet> xterm ctrlr
{path to gnmi_get binary} \
-ca ../demo/cert/client/ca.crt \
-cert ../demo/cert/client/client.crt \
-key ../demo/cert/client/client.key \
-target_name www.example.com \
-target_addr 10.0.0.1:8080 \
-xpath "/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/config/channel"
```

The output should be similar to:
```
== getResponse:
notification: <
  timestamp: 1521145574058185274
  update: <
    path: <
      elem: <
        name: "access-points"
      >
      elem: <
        name: "access-point"
        key: <
          key: "hostname"
          value: "link022-pi-ap"
        >
      >
      elem: <
        name: "radios"
      >
      elem: <
        name: "radio"
        key: <
          key: "id"
          value: "1"
        >
      >
      elem: <
        name: "config"
      >
      elem: <
        name: "channel"
      >
    >
    val: <
      uint_val: 8
    >
  >
>
```
