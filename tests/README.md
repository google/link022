# Link022 Integration tests
This doc contains the steps to run the integration tests against the link022 agent.

## Setup testing environment
The setup needs Python 2.7 environment and some additional packages.
```
apt-get install python python-netaddr mininet
```

## Compiling Link022 binaries
Follow the [instructions](../README.md) to compile the Link022 binary and set up the certificates.

## Run the tests
  - Set up the required paths
    - GOPATH points to the Go directory for the compiled files of link022 and its dependencies.
    - LINK022 points to root directory of the LINK022 source code.
  - Run the tests
```
sudo python integration.py \
  --target_cmd "${GOPATH}/bin/agent -ca ${LINK022}/demo/cert/server/ca.crt \
  -cert ${LINK022}/demo/cert/server/server.crt \
  -key ${LINK022}/demo/cert/server/server.key -eth_intf_name target-eth1 \
  -wlan_intf_name target-eth2" \
  --gnmi_set "${GOPATH}/bin/gnmi_set \
  -ca=${LINK022}/demo/cert/client/ca.crt \
  -cert=${LINK022}/demo/cert/client/client.crt \
  -key=${LINK022}/demo/cert/client/client.key \
  -target_name=www.example.com \
  -target_addr=10.0.0.1:8080 \
  -replace=/:@${LINK022}/demo/ap_config.json"

```
