#!/bin/sh

mkdir certs/
openssl ecparam -name secp521r1 -genkey -noout -out certs/ca.key
openssl req -new -x509 -days 3650 -key certs/ca.key -out certs/ca.crt -subj "/CN=proxy"
