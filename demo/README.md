# Link022 Demo
This doc contains the steps to run a demo.

## Setup demo environment
Here is the structure of the demo setup.
![alt text](./Link022_diagrams-demo.png "Demo setup architecture")

The setup has two components, linked directly by an ethernet cable.
  - One Raspberry Pi device as the Link022 AP (gnmi target).
  - The other device (such as Raspberry Pi) as the gateway. It provides reqiured services (dhcp, radius and gnmi client).

### Setup Link022 AP
On the Link022 AP device, run the commands in this [instruction](../README.md)

### Setup Gateway
On the device for gateway follow the steps below.
1. Install Golang.
```
wget https://storage.googleapis.com/golang/go1.7.linux-armv6l.tar.gz
sudo tar -C /usr/local -xzf go1.7.linux-armv6l.tar.gz
export PATH=$PATH:/usr/local/go/bin
```
2. Install dependencies.
``` 
sudo apt-get install dnsmasq freeradius
```
3. Download GNMI clients.
```
export GOPATH=$HOME/go
go get github.com/google/gnxi/gnmi_set
go get github.com/google/gnxi/gnmi_get
```
4. Download the [demo folder](./).
5. Enter the demo folder.
```
cd <path to demo folder>
```
6. Setup the gateway.
```
./server.sh
```
This script creates a network namespace (lk22 by default), which has access to the Link022 AP device.

Note: Reboot the device or run 'cleanup_servers.sh' script to clean up the gateway.

## Demo
*All demo commands are executed inside the lk22 namspace on the gateway device.*
To enter the namespace, run
```
sudo ip netns exec lk22 bash
```

### Push the full configuration to AP
```
Add later
```
### Update AP configuration
```
Add later
```
### Fetch AP configuration
```
Add later
```
