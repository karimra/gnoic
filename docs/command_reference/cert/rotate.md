# Cert Rotate

## Description

There is sometimes a need to renew existing certificates on a target, the [gNOI Cert Rotate RPC](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L93) does exactly that.

`gNOIc` supports [*target generated CSR*](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L108)
as well as [*client side generated CSR*](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L124).

- If the flag `--gen-csr` is present, `gNOIc` generates the certificate locally instead of relying on the Target.

- In the opposite case, `gNOIc` will check if the target supports CSR generation, using the [CanGenerateCSR RPC](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L164).
If the target can generate a CSR, `gNOIc` will rely on the target to generate a CSR. Otherwise it generates the CSR and certificate locally.

The `--gen-csr` flag allows testing both message flows for a target that supports CSR generation.

The `rotate` command acts as the client side of the `gNOI Cert Rotate RPC` and effectively renews a previously installed certificate in 3 or 4 steps, depending on the CSR generation method:

- Target Generated CSR:
    - Start a bi-directional gRPC stream.
    - Request a CSR from the target.
    - Sign the Certificate using the provided CA.
    - Load the certificate into the target.

- Client Generated CSR:
    - Start a bi-directional gRPC stream.
    - Generate and Sign the Certificate using the provided CA.
    - Load the certificate into the target.

### Usage

`gnoic [global-flags] cert rotate [local-flags]`

### Flags

#### cert-type

The `--cert-type` flag sets the desired certificate type.

defaults to `CT_X509`

#### city

The `--city` sets the `City` part of the certificate DN (Distinguished Name)

#### common-name

The `--common-name` sets the `CommonName` part of the certificate DN (Distinguished Name)

#### country

The `--country` sets the `Country` part of the certificate DN (Distinguished Name)

#### email-id

The `--email-id` sets the `EmailID` part of the certificate DN (Distinguished Name)

#### gen-csr

The `--gen-csr` flag allows the running the rotate command with a locally generated certificate,

as opposed to using the [GenerateCSR](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L190)

to generate a CSR on the Target side.

#### ip-address

The `--ip-address` sets an IP address to be added to the certificate as a SAN.

#### id

The `--id` flag sets the desired certificate ID.

#### key-type

The `--key-type` flag sets the desired key type, defaults to `KT_RSA`

#### min-key-size

The `--min-key-size` flag sets the minimum desired key size, defaults to `1024`

#### org

The `--org` sets the `OrganizationName` part of the certificate DN (Distinguished Name)

#### print-csr

The `--print-csr` if set, `gNOIc` prints the CSR generated by the Target.

#### org-unit

The `--org-unit` sets the `OrganizationalUnit` part of the certificate DN (Distinguished Name)

#### state

The `--state` sets the `State` part of the certificate DN (Distinguished Name)

#### validity

The `--validity` sets the validity duration of the certificate, the expected format is Golang's duration format: 1s, 10m, 1h, 87600h.

defaults to `87600h` (10 years)

### Examples

```bash
gnoic -a 172.17.0.100:57400 -u admin -p admin --insecure \
      cert \
      --ca-cert cert.pem --ca-key key.pem \
      rotate --id cert2 \
      --common-name router1 \
      --org OrgInc --org-unit OrgUnit
```

```bash
INFO[0000] read local CA certs                          
INFO[0000] "172.17.0.100:57400" signing certificate "CN=router1,OU=OrgUnit,O=OrgInc" with the provided CA 
INFO[0000] "172.17.0.100:57400" rotating certificate id=cert2 "CN=router1,OU=OrgUnit,O=OrgInc" 
INFO[0000] "172.17.0.100:57400" Rotate RPC successful   
```
