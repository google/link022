# OpenConfig Telemetry Exporter (Prometheus)

This directory contains a exporter that periodically collects AP's info via gNMI and exposes converted metrics to a web page. Prometheus can monitor AP's status by scrape that web page.

## Get Started

Follow steps below to set up environment and start exposition server

### Prerequisites

Dolnload entire Link022 repository.  
Install golang 1.10+ (get it from: https://golang.org/doc/install#install)  
Install dependencies:  
Change directory in terminal to this folder, then run below command.

```
go get -t ./...
```

### Compile Exporter

```
go build exposition_server.go openconfig_ap_exporter.go
```

This command will generate binary file named exposition_server

### Start Exporter
Run the exporter binary. It takes two categories of input parameters:

1. gNMI client certs config:
    * ca: CA certificate file
    * cert: Certificate file
    * key: Private key file
2. gNMI target config:
    * target_addr: the target address in the format of host:port
    * target_name: the target name for verifing the hostname returned by TLS handshake
3. Exposition server config
    * listen_addr: the address for server to listen HTTP requests. The HTTP resource for scraping is /metrics.

Here is one example:

```
./exposition_server \
-ca ../../../demo/cert/client/ca.crt \
-cert ../../../demo/cert/client/client.crt \
-key ../../../demo/cert/client/client.key \
-target_name www.example.com \
-target_addr 10.0.0.1:8080 \
-listen_addr 127.0.0.1:8080
```

Note: The default location of exporter log file is "/tmp/exposition_server.INFO"