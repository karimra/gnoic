# Cert GetCertificates

### Description

The certificates installed on a target can be queries with the [Cert GetCertificates RPC](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L154)

Using the `cert get-certs` command, the user can query all or some of the certificates installed on the target.

The command takes 3 flags:

- `id`: certificateID which can be repeated as many times as necessary to query a subset of the certificates.
- `details`: displays each certificate's details in a format similar to the openssl cli tool.
- `save`: save each certificate in the local file system, under a directory called after the target address.
  
### Usage

`gnoic [global-flags] cert get-certs [local-flags]`

### Flags

#### id

The `--id` takes one or multiple (comma-separated) certificate IDs

If not supplied, `gNOIc` displays all the available certificates.

#### details

The `--details` flag makes `gNOIc` displays each certificate's details in a format similar to the openssl cli tool.

#### save

When the `--save` flag is present, each certificate is saved in the local file system, under a directory called after the target address.

### Examples

```bash
gnoic -a 172.17.0.100:57400 --insecure -u admin -p admin cert get-certs
```

```md
+--------------------+-------+---------------------------+---------+---------+------------+----------------------+----------------------+--------------+
|    Target Name     |  ID   |     Modification Time     |  Type   | Version |  Subject   |      Valid From      |     Valid Until      |   IP Addrs   |
+--------------------+-------+---------------------------+---------+---------+------------+----------------------+----------------------+--------------+
| 172.17.0.100:57400 | cert3 | 2021-05-05T13:00:46+08:00 | CT_X509 | 3       | CN=router1 | 2021-05-05T04:00:46Z | 2031-05-03T05:00:46Z | 172.17.0.100 |
| 172.17.0.100:57400 | cert4 | 2021-05-05T13:00:46+08:00 | CT_X509 | 3       | CN=router1 | 2021-05-05T04:00:47Z | 2031-05-03T05:00:47Z | 172.17.0.100 |
| 172.17.0.100:57400 | cert5 | 2021-05-05T13:00:48+08:00 | CT_X509 | 3       | CN=router1 | 2021-05-05T04:00:48Z | 2031-05-03T05:00:48Z | 172.17.0.100 |
| 172.17.0.100:57400 | cert6 | 2021-05-05T13:00:48+08:00 | CT_X509 | 3       | CN=router1 | 2021-05-05T04:00:49Z | 2031-05-03T05:00:49Z | 172.17.0.100 |
+--------------------+-------+---------------------------+---------+---------+------------+----------------------+----------------------+--------------+
```