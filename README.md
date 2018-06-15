[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GoDoc](https://godoc.org/github.com/google/link022?status.svg)](https://godoc.org/github.com/google/link022)
[![Go Report Card](https://goreportcard.com/badge/github.com/google/link022)](https://goreportcard.com/report/github.com/google/link022)
[![Build Status](https://travis-ci.org/google/link022.svg?branch=master)](https://travis-ci.org/google/link022)
[![codecov](https://codecov.io/gh/google/link022/branch/master/graph/badge.svg)](https://codecov.io/gh/google/link022)

# Link022: an open WiFi access point
Link022 is an open reference implementation and experimental platform for an OpenConfig and gNMI
controlled WiFi access point.

The central part of Link022 is an gNMI agent that runs on a Linux host with WiFi capability. The
agent turns the host into an gNMI capable wireless access point which can be configured using
OpenConfig models.

*  See [gNMI Protocol documentation](https://github.com/openconfig/reference/tree/master/rpc/gnmi).
*  See [Openconfig documentation](http://www.openconfig.net/).

## Get Started
This repository contains following components.

### Link022 agent
A WiFi management component that runs on a Link022 AP, with OpenConfig and gNMI implemented.
It supports gNMI "SET" and "GET" opertions for AP configuration.

To run the agent on a Raspberry Pi device, see the [start guide](agent/README.md).

### Link022 demo
A demo for configuring Link022 AP though gNMI. [demo guide](demo/README.md)

### Link022 emulator
An emulator that runs Link022 agent inside a Linux namespace. [start guide](emulator/README.md)

### gNMI test kit.
A tool to test the gNMI functionality of an AP device. [start guide](testkit/README.md)

## Disclaimer
*  This is not an official Google product.
*  See [how to contribute](CONTRIBUTING.md).
