#!/bin/bash

# Create A Self-Signed CA Certificate
openssl genrsa -out ca.key 1024
openssl req -new -subj \
       '/C=US/ST=CA/L=Mountain View/O=LINK0022/OU=Test/CN=AP' -key ca.key -out ca.csr
openssl x509 -days 1095 -extfile ca.ext -signkey ca.key -in ca.csr -req -out ca.pem

# Create A Server Certificate
openssl genrsa -out server.key 1024
echo -ne '01' > ca.serial
openssl req -subj \
       '/C=US/ST=CA/L=Mountain View/O=LINK0022/OU=Test/CN=AP' -new -key server.key -out server.csr
openssl x509 -days 730 -extfile server.ext -CA ca.pem -CAkey ca.key -CAserial ca.serial -in server.csr -req -out server.pem

# Create dh.
openssl dhparam -out dh 1024
