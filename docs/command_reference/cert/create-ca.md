# Cert CreateCa

### Description

In oder to speed up testing, the `create-ca` command was added to generate a self-signed Certificate authority with the purpose of signing certificates that will be pushed to the targets using the `install` or `rotate` commands.

The `create-ca` commands takes up to 13 flags that have "smart" defaults designed to simplify usage.

### Usage

`gnoic [global-flags] cert create-ca [local-flags]`

### Flags

#### org

The `--org` sets the `OrganizationName` part of the certificate DN (Distinguished Name)

defaults to `gNOIc`

#### org-unit

The `--org-unit` sets the `OrganizationalUnit` part of the certificate DN (Distinguished Name)

defaults to `gNOIc Certs`

#### country

The `--country` sets the `Country` part of the certificate DN (Distinguished Name)

defaults to `OC`

#### state

The `--state` sets the `State` part of the certificate DN (Distinguished Name)

#### locality

The `--locality` sets the `Locality` part of the certificate DN (Distinguished Name)

#### street-address

The `--street-address` sets the `StreetAddress` part of the certificate DN (Distinguished Name)

#### postal-code

The `--postal-code` sets the `PostalCode` part of the certificate DN (Distinguished Name)

#### validity

The `--validity` sets the validity duration of the certificate, the expected format is Golang's duration format: 1s, 10m, 1h, 87600h.

defaults to `87600h` (10 years)

#### key-size

The `--key-size` sets the key size to be generated.

defaults to 2048

#### email

The `--email` sets the `Email` part of the certificate DN (Distinguished Name)

#### common-name

The `--common-name` sets the `CommonName` part of the certificate DN (Distinguished Name)

#### key-out

The `--key-out` defines the path where the generated key needs will stored.

defaults to `key.pem`

#### cert-out

The `--cert-out` defines the path where the generated certificate will be stored.

defaults to `cert.pem`

### Examples

```bash
gnoic cert create-ca
```

```md
INFO[0000] CA certificate written to cert.pem           
INFO[0000] CA key written to key.pem  
```
