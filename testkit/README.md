# Test Kit

This directory contains a tool to test the gNMI functionality of an AP device.

## Get Started

Follow steps below to set up a testing environment.

### Prerequisites
Install golang 1.7+ (get it from: https://golang.org/doc/install#install)
Install dependencies:
```
go get github.com/golang/glog
go get github.com/google/gnxi/utils/credentials
```

### 1. Download Link022 repository
Download the entire [repository](../).

### 2. Compile the test kit
```
cd testkit
go build test_kit.go
```

### 3. Start the test kit
Run the test kit binary. It takes three categories of input parameters:
1. gNMI client certs config:
    * ca: CA certificate file
    * cert: Certificate file
    * key: Private key file
2. gNMI target config:
    * targetAddr: the target address in the format of host:port
    * targetName: the target name for verifing the hostname returned by TLS handshake
3. target test cases:
    * test_file: the file containing gNMI test
    Note: To run mulitple tests in one single run, specifying multiple "-test_file" parameters in input.
4. mode selection:
    * pause_mode: if enabled, the test kit pauses after each test case. Disabled by default.
    * insecure: if enabled, gNMI client skips TLS validation. (May also need to [disable TLS validation](https://golang.org/pkg/crypto/tls/#ClientAuthType) on target side.)

Here is one example:
```
./test_kit -alsologtostderr \
-ca ../demo/cert/client/ca.crt \
-cert ../demo/cert/client/client.crt \
-key ../demo/cert/client/client.key \
-target_name www.example.com \
-target_addr 127.0.0.1:8080 \
-test_file=testdata/simple_test.json

insecure mode on:

./test_kit \
-alsologtostderr \
-insecure \
-target_name www.example.com \
-target_addr 127.0.0.1:8080 \
-test_file=testdata/simple_test.json
```

Note: The default location of test kit log file is "/tmp/test_kit.INFO"

## Run test on Link022 emulator

The test kit is compatible with any gNMI target with [ap-manager](https://github.com/openconfig/public/blob/master/release/models/wifi/ap-manager/openconfig-ap-manager.yang) and [access-points](https://github.com/openconfig/public/blob/master/release/models/wifi/access-points/openconfig-access-points.yang) supported.

The simplist way to check whether the test kit works is running it against an emulated Link022 AP.

### 1. Start the Link022 emulator
Follow the [detailed instruction](../emulator/README.md#start-emulator) to run a Link022 emulator.

Note, return here after you've completed Step 4 (verify the setup), but before Step 5.

An emulated Link022 AP should run inside mininet node "target".
* management interface IP: 10.0.0.1
* gNMI port: 8080

Note: The default log file location:
* link022 AP: "/tmp/link022_agent.INFO"
* emulator: "/tmp/link022_emulator.log"

### 2. Start the test kit
Run the test kit with sample gNMI test on mininet node "ctrlr", setting the target to the emulated Link022 AP.

Note, if testing agianst the Link022 emulator, be sure the hostname of the machine running the emulator matches the hostname in 'tests/ap_config.json'. For example, "link022-pi-ap".
```
mininet> ctrlr {path to testkit binary} -logtostderr \
-ca ../demo/cert/client/ca.crt \
-cert ../demo/cert/client/client.crt \
-key ../demo/cert/client/client.key \
-target_name www.example.com \
-target_addr 10.0.0.1:8080 \
-test_file=../testkit/testdata/simple_test.json
```

The output should be similar to:
```
=Test results=
--------------------
[PASS] Simple Test
|-[PASS] Push entire config
|-[PASS] Update Radio Config
|-[PASS] Fetch SSID Config
--------------------
[PASS] [Simple Test] Passed - 3, Failed - 0
```

## Create custom test cases
One test file contains one JSON blob. It represents a single gNMI test, with following properties.
* name: test name
* description: detail description of this test
* model: OpenConfig model used in this test case
* test_cases: a list of gNMI test cases to run in this test

### Test Case
"test_cases" is a list of gNMI test cases. It is defined as the following:
```
// TestCase describes a gNMI test cases.
type TestCase struct {
    // Name is the test case name.
    Name string `json:"name"`
    // Description is the detail description of this test case.
    Description string `json:"description"`
    // Model is used to construct the UseModels property in gNMI requests.
    // If not specified, all gNMI requests are sent without UseModels.
    Model *ModelData `json:"model"`
    // OPs contains a list of operations need to be processed in this test case.
    // All operations are processed in one single gNMI message.
    OPs []*Operation `json:"ops"`
}

// ModelData describes the OpenConfig model used in this test case.
type ModelData struct {
    Name         string `json:"name"`
    Organization string `json:"organization"`
    Version      string `json:"version"`
}

// Operation represents a gNMI operation.
type Operation struct {
    // Type is the gNMI operation type.
    Type OPType `json:"type"`
    // Path is the xPath of the target field/branch.
    Path string `json:"path"`
    // StatePath is the xPath of the corresponding state field/branch.
    // If specified, testkit will verify the state update.
    StatePath string `json:"state_path"`
    // Val is the string format of the desired value.
    // Val should be unset for gNMI delete operation.
    // Supported types:
    //     Integer: "1", "2"
    //     Float: "1.5", "2.4"
    //     String: "abc", "defg"
    //     Boolean: "true", "false"
    //     IETF JSON from file: "@ap_config.json"
    Val string `json:"val"`
}
```

There are three categories of test cases, based on the operations specified:
1. config test: A test case that contains only gNMI SET operation (replace, update, delete).
2. state fetching test: A test case that contains only gNMI GET operation.
3. telemetry streaming test: A test case that contains only gNMI SUBSCRIBE operation.

Note: Test cases with combined operation types are invalid.

When running one test case, the test kit sends one gNMI message to target (containing all given operations), and verifies the received response.
For config-related test cases, it also checks whether the updated configuraiton presents on target (through gNMI Get method).
A test case fails if any error detected while executing.

### Sample test file

One sample test file:
```
{
  "name":"Simple Test",
  "description":"This is an example of gNMI test.",
  "test_cases":[
    {
      "name":"Push entire config",
      "description":"Push the entire configuration to AP device.",
      "model":{
        "name": "openconfig-access-points",
        "organization": "OpenConfig working group",
        "version": "0.1.0"
      },
      "ops":[
        {
          "type":"replace",
          "path":"/",
          "val":"@../tests/ap_config.json"
        }
      ]
    },
    {
      "name":"Update Radio Config",
      "ops":[
        {
          "type":"update",
          "path":"/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/config/channel",
          "state_path":"/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/state/channel",
          "val": "6"
        },
        {
          "type":"update",
          "path":"/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/config/channel-width",
          "state_path":"/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/state/channel-width",
          "val": "20"
        }
      ]
    },
    {
      "name":"Fetch SSID Config",
      "ops":[
        {
          "type":"get",
          "path":"/access-points/access-point[hostname=link022-pi-ap]/ssids/ssid[name=Auth-Link022]/config/vlan-id",
          "val": "300"
        },
        {
          "type":"get",
          "path":"/access-points/access-point[hostname=link022-pi-ap]/ssids/ssid[name=Guest-Link022]/config/vlan-id",
          "val": "200"
        }
      ]
    }
  ]
}
```
For more examples, see [testdata directory](./testdata)
