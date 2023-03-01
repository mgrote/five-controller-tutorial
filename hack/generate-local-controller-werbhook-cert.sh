#!/usr/bin/env bash

set -e

mkdir -p /tmp/k8s-webhook-server/serving-certs

cat <<EOF > /tmp/k8s-webhook-server/serving-certs/tls_config.txt
[ req ]
default_bits       = 2048
default_md         = sha512
default_keyfile    = ca.key
prompt             = no
encrypt_key        = yes

# base request
distinguished_name = req_distinguished_name

# distinguished_name
[ req_distinguished_name ]
countryName            = "DE"                     # C=
stateOrProvinceName    = "Berlin"                # ST=
localityName           = "Berlin"                # L=
organizationName       = "frup.org"             # O=
organizationalUnitName = "berlin"       # OU=
commonName             = "berlin.frup.org"          # CN=
emailAddress           = "no-reply@berlin.frup.org" # CN/emailAddress=
EOF

# CREATE THE PRIVATE KEY FOR OUR CUSTOM CA
openssl genrsa -out /tmp/k8s-webhook-server/serving-certs/tls.key 2048

# GENERATE A CA CERT WITH THE PRIVATE KEY
openssl req -new -x509 -key /tmp/k8s-webhook-server/serving-certs/tls.key -out /tmp/k8s-webhook-server/serving-certs/tls.crt -config /tmp/k8s-webhook-server/serving-certs/tls_config.txt
