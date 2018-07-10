# Agent

This directory contains the WiFi management component that runs on the
Link022 Pi AP.

## Get Started
For documentation on installing a full demo system (with both a gNMI client and link022 access point)
please see the [full demo documentation](../demo/README.md) which explains a fully complete system.

The following instructions will get you a Link022 AP on a Linux-based device.

### Prerequisites
Have a device set up. (Tested with Raspian Stretch)

Install Golang.
```
If running on Raspberry Pi:
wget https://storage.googleapis.com/golang/go1.7.linux-armv6l.tar.gz
sudo tar -C /usr/local -xzf go1.7.linux-armv6l.tar.gz

For other systems:
Install golang 1.7+ (get it from: https://golang.org/doc/install#install)
```

Set up Path:
```
export PATH=$PATH:/usr/local/go/bin
```

Install dependencies.
```
sudo apt-get install udhcpc bridge-utils hostapd git
```

### Download Link022 agent
```
export GOPATH=$HOME/go
go get github.com/google/link022/agent
```

### Download certificates
Download sample certificates from [the demo directory](../demo/cert/server/).
Or you can use your own cert. Sample commands to generate certificates can be found [here](../demo/cert/generate_cert.sh).

### Configuring network interfaces of device
Editing the file /etc/network/interfaces on device.
```
auto lo
iface lo inet loopback

auto eth0
iface eth0 inet dhcp

# Disable all WLAN interfaces.
auto wlan0
iface wlan0 inet static
    address 0.0.0.0

# Repeat for other WLAN interfaces.
```
Note: Reboot the device to make change take effect.

### Running Link022 agent
```
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
sudo env PATH=$PATH agent -ca=<path to ca.crt> -cert=<path to server.crt> -key=<path to server.key> -eth_intf_name=<the eth interface> -wlan_intf_name=<the wlan interface for AP radio> -gnmi_port=<port number>
```

The default log file is "/tmp/agent.INFO". It can be modified by "-log_dir" option.

Note: Make sure the chosen wireless device supports AP mode and has enough
capability.