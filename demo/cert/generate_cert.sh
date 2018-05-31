#!/bin/bash

mkdir -p tls_cert_key

mkdir -p tls_cert_key/server
openssl req -x509 -newkey rsa:4096 -keyout tls_cert_key/server/server.key -out tls_cert_key/server/server.crt -days 3650 -nodes -subj "/C=US/ST=CA/L=Sunnyvale/O=GNMI/OU=Org/CN=www.example.com" -sha256

mkdir -p tls_cert_key/client
openssl req -x509 -newkey rsa:4096 -keyout tls_cert_key/client/client.key -out tls_cert_key/client/client.crt -days 3650 -nodes -subj "/C=US/ST=CA/L=Sunnyvale/O=GNMI/OU=Test/CN=AP" -sha256

cp tls_cert_key/server/server.crt tls_cert_key/client/ca.crt
cp tls_cert_key/client/client.crt tls_cert_key/server/ca.crt