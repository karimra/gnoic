# Cert Revoke
### Description

The [gNOI Revoke RPC](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L160) revokes any previously installed certificates.

The command `revoke` acts as the client side of the `gNOI Cert RevokeCertificates` RPC.

The Command takes 2 flags:

- `id`: a repeated flag that can be used to list (comma-separated) the certificate IDs to be revoked
- `all`: if present, `gNOI` revokes all installed certificates.

### Usage

`gnoic [global-flags] cert revoke [local-flags]`

### Flags

#### id

The `--id` flag specifies the certificate(s) to be removed. It can be repeated as many times as needed or supplied with a comma-separated list of certificateIDs

#### all

If present, `gNOIC` retrieves the list of certificates present on the target and revokes all of them with a single `Revoke` RPC.

### Examples

```bash
gnoic -a 172.17.0.100:57400 -u admin -p admin --insecure \
      cert revoke --id cert2
```

```bash
INFO[0000] "172.17.0.100:57400" certificateID=cert2 revoked successfully 
```
