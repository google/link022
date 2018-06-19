#!/bin/bash

export PATH=$PATH:~/go/bin
sudo env PATH=$PATH ip netns exec lk22 gnmi_set \
-ca=../cert/client/ca.crt \
-cert=../cert/client/client.crt \
-key=../cert/client/client.key \
-target_name=www.example.com \
-target_addr=192.168.11.8:8080 \
-replace=/:@../ap_config_1r.json
