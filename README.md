[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GoDoc](https://godoc.org/github.com/google/link022?status.svg)](https://godoc.org/github.com/google/link022)
[![Go Report Card](https://goreportcard.com/badge/github.com/google/link022)](https://goreportcard.com/report/github.com/google/link022)
[![Build Status](https://travis-ci.org/google/link022.svg?branch=master)](https://travis-ci.org/google/link022)

# Link022: an open WiFi access point
Link022 is an open reference implementation and experimental platform for an OpenConfig and gNMI
controlled WiFi access point.

The central part of Link022 is an gNMI agent that runs on a Linux host with WiFi capability. The
agent turns the host into an gNMI capable wireless access point which can be configured using
OpenConfig models.

*  See [gNMI Protocol documentation](https://github.com/openconfig/reference/tree/master/rpc/gnmi).
*  See [Openconfig documentation](http://www.openconfig.net/).

## Getting Started
The following instructions will get you a Link022 AP on a Raspberry Pi device.

### Prerequisites
Have a Raspberry Pi device set up. (Tested with Raspbian)

Install Golang on Raspberry Pi.
```
wget https://storage.googleapis.com/golang/go1.7.linux-armv6l.tar.gz
sudo tar -C /usr/local -xzf go1.7.linux-armv6l.tar.gz
export PATH=$PATH:/usr/local/go/bin
```
Install dependencies.
```
sudo apt-get install udhcpc bridge-utils hostapd
```

### Download Link022 agent
```
export GOPATH=$HOME/go
go get github.com/google/link022/agent
```

### Download certificates
Download sample certificates from github.com/google/link022/examples/cert/server/.
Or you can use your own cert.

### Configuring network interfaces of Pi
Editing the file /etc/network/interfaces on Pi.
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
Note: Make sure the chosen wireless device supports AP mode and has enough
capability.

## Disclaimer
*  This is not an official Google product.
*  See [how to contribute](CONTRIBUTING.md).
