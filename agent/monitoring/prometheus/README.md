# OpenConfig Telemetry Exporter (Prometheus)

This directory contains an exporter that periodically collects AP's info via gNMI and exposes converted metrics to a web page. Prometheus can monitor AP's status by scraping that web page.

## Get Started

Follow steps below to set up environment and start exposition server

### Prerequisites

Download entire Link022 repository.  
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

This command will generate binary file named exposition_server.  

### Start Exporter
Run the exporter binary. It takes three categories of input parameters:  

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

## Monitoring In Prometheus

Follow steps below to set up Prometheus and monitoring AP status.  

### Prometheus Getting Started

Download [Prometheus](https://prometheus.io/download/). Follow official Prometheus [tutorial](https://prometheus.io/docs/prometheus/latest/getting_started/) to learn how to configure and start it.

### Configuring Prometheus to monitor AP

If exposition server is listening on 127.0.0.1:8080  
Save the following basic Prometheus configuration as a file named prometheus.yml

```
global:
  scrape_interval:     15s # Exposition server sends gNMI request every 15 seconds.

scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.
  - job_name: 'link022-pi-ap'
    static_configs:
      - targets: ['127.0.0.1:8080']
```

### Start Prometheus

Start Prometheus according to its official tutorial.  
By default, Prometheus admin page is localhost:9090.
You can see all exported AP status metrics in that page.