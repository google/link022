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

Here is one example:
```
./test_kit -logtostderr \
-ca ../demo/cert/client/ca.crt \
-cert ../demo/cert/client/client.crt \
-key ../demo/cert/client/client.key \
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
--------------------
[PASS] [Simple Test] Passed - 2, Failed - 0
```

## Create custom test cases
One test file contains one JSON blob. It represents a single gNMI test, with following properties.
* name: test name
* description: detail description of this test
* config_tests: a list of config-related test cases to run in this test
* state_tests: a list of state-related test cases to run in this test (This is not supported yet.)

### config_tests
"config_tests" is a list of config-related test cases. It is defined as the following:
```
// TestCase describes a config-related gNMI test cases.
type TestCase struct {
    // Name is the test case name.
    Name string `json:"name"`
    // Description is the detail description of this test case.
    Description string `json:"description"`
    // OPs contains a list of operations need to be processed in this test case.
    // All operations are processed in one single gNMI SetRequest.
    OPs []*operation `json:"ops"`
}

type operation struct {
    // Type is the gNMI SET operation type.
    Type OPType `json:"type"`
    // Path is the xPath of the target field/branch.
    Path string `json:"path"`
    // Val is the string format of the desired value.
    // Supported types:
    //     Integer: "1", "2"
    //     Float: "1.5", "2.4"
    //     String: "abc", "defg"
    //     Boolean: "true", "false"
    //     IETF JSON from file: "@ap_config.json"
    Val string `json:"val"`
}
```

For each test case, the test kit sends one gNMI SetRequest to target, containing all specified operations. Then, it verifies the received SetResponse. If gNMI Set succeeds, it also checks whether the updated configuraiton presents on target (through gNMI Get method).
A test case fails if either the gNMI Set operation fails or the desired configuraiton is not present on target.

### state_tests
"state_tests" is a list of state-related test cases.
For each test case, the test kit fetches a branch/leaf from target (though gNMI Get/Subscribe method), comparing it with the provided value. A test case fails if the fetched value does not match the desired one.

Note: "state_tests" runs after "config_tests".

Note: "state_tests" is not implemented yet.


### Sample test file

One sample test file:
```
{
  "name":"Simple Test",
  "description":"This is an example of gNMI test.",
  "config_tests":[
    {
      "name":"Push entire config",
      "description":"Push the entire configuration to AP device.",
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
          "val": "6"
        },
        {
          "type":"update",
          "path":"/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=1]/config/channel-width",
          "val": "20"
        }
      ]
    }
  ]
}
```
For more examples, see [testdata directory](./testdata)
