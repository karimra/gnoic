# Basic Usage

### Non RPC Commands

#### Tree

The tree command can be used to display `gNOIc's` commands tree.

```bash
gnoic tree
gnoic
 ├─── cert
 │    ├─── can-generate-csr
 │    ├─── create-ca
 │    ├─── generate-csr
 │    ├─── get-certs
 │    ├─── install
 │    ├─── load
 │    ├─── load-ca
 │    ├─── revoke
 │    └─── rotate
 ├─── file
 │    ├─── get
 │    ├─── put
 │    ├─── remove
 │    ├─── stat
 │    └─── transfer
 ├─── help [command]
 ├─── system
 │    ├─── cancel-reboot
 │    ├─── ping
 │    ├─── reboot
 │    ├─── reboot-status
 │    ├─── set-package
 │    ├─── switch-control-processor
 │    ├─── time
 │    └─── traceroute
 ├─── tree
 └─── version
```

```bash
gnoic tree --flat
gnoic
 gnoic cert
 gnoic cert can-generate-csr
 gnoic cert create-ca
 gnoic cert generate-csr
 gnoic cert get-certs
 gnoic cert install
 gnoic cert load
 gnoic cert load-ca
 gnoic cert revoke
 gnoic cert rotate
 gnoic file
 gnoic file get
 gnoic file put
 gnoic file remove
 gnoic file stat
 gnoic file transfer
 gnoic help [command]
 gnoic system
 gnoic system cancel-reboot
 gnoic system ping
 gnoic system reboot
 gnoic system reboot-status
 gnoic system set-package
 gnoic system switch-control-processor
 gnoic system time
 gnoic system traceroute
 gnoic tree
 gnoic version
```

```bash
gnoic tree --details
gnoic
 ├─── cert
 │    ├─── can-generate-csr [--cert-type] [--key-size] [--key-type]
 │    ├─── create-ca [--cert-out] [--common-name] [--country] [--email] [--key-out] [--key-size] [--locality] [--org] [--org-unit] [--postal-code] [--state] [--street-address] [--validity]
 │    ├─── generate-csr [--cert-type] [--city] [--common-name] [--country] [--email-id] [--id] [--ip-address] [--key-type] [--min-key-size] [--org] [--org-unit] [--state]
 │    ├─── get-certs [--details] [--id] [--save]
 │    ├─── install [--cert-type] [--city] [--common-name] [--country] [--email-id] [--id] [--ip-address] [--key-type] [--min-key-size] [--org] [--org-unit] [--print-csr] [--state] [--validity]
 │    ├─── load
 │    ├─── load-ca
 │    ├─── revoke [--all] [--id]
 │    └─── rotate [--cert-type] [--city] [--common-name] [--country] [--email-id] [--id] [--ip-address] [--key-type] [--min-key-size] [--org] [--org-unit] [--print-csr] [--state] [--validity]
 ├─── file
 │    ├─── get [--file] [--local-file] [--target-prefix]
 │    ├─── put [--chunk-size] [--file] [--hash-method] [--permission] [--remote-name]
 │    ├─── remove [--file]
 │    ├─── stat [--file] [--humanize]
 │    └─── transfer
 ├─── help [command]
 ├─── system
 │    ├─── cancel-reboot [--message] [--subcomponent]
 │    ├─── ping [--count] [--destination] [--do-not-fragment] [--do-not-resolve] [--interval] [--protocol] [--size] [--source] [--wait]
 │    ├─── reboot [--delay] [--force] [--message] [--method] [--subcomponent]
 │    ├─── reboot-status [--subcomponent]
 │    ├─── set-package
 │    ├─── switch-control-processor [--path]
 │    ├─── time
 │    └─── traceroute [--destination] [--do-not-fragment] [--do-not-resolve] [--initial-ttl] [--interval] [-3 | --l3protocol] [-4 | --l4protocol] [--max-ttl] [--size] [--source] [--wait]
 ├─── tree [--details] [--flat] [-h | --help]
 └─── version
```

#### Cert CreateCa

`[ cert create-ca ]` is a non RPC related command, it is a convenience command that can be used to generate a self signed Certificate Authority for quick testing.

It creates a CA key pair stored as `cert.pem` and `key.pem`

Check the [command reference](command_reference/cert/create-ca.md) for more flags options

```bash
gnoic cert create-ca
INFO[0000] CA certificate written to cert.pem           
INFO[0000] CA key written to key.pem  
```

## gNOI RPC Commands

The following examples demonstrate the basic usage of the `gnoic` in a scenario where the remote target runs insecure (not TLS enabled) gNOI server.
The `admin:admin` credentials are used to connect to the gNMI server running at `172.17.0.100:57400` address.

### Certificates Management gNOI service

#### Cert CanGenerateCSR

The `[ can-generate-csr | cgc ]` command acts as the client side of the [Cert CanGenerateCSR RPC](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L164)

Check the [command reference](command_reference/cert/can-generate-csr.md) for more flags options

```bash
gnoic -a 172.17.0.100:57400 --insecure -u admin -p admin cert can-generate-csr
```

```md
+--------------------+------------------+
|    Target Name     | Can Generate CSR |
+--------------------+------------------+
| 172.17.0.100:57400 | true             |
+--------------------+------------------+
```

#### Cert GenerateCSR

#### Cert GetCertificates

The `[ get | get-certs ]` command acts as the client side of the [Cert GetCertificates RPC](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L154)

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

#### Cert Install

The `[ install ]` command acts as the client side of the [Cert Install RPC](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L138)

```bash
gnoic -a 172.17.0.100:57400 --insecure -u admin -p admin \
      cert \
      --ca-cert cert.pem  --ca-key key.pem  \
      install \
      --ip-address 172.17.0.100 --common-name router1 --id cert2
```

```bash
INFO[0000] read local CA certs                          
INFO[0000] "172.17.0.100:57400" signing certificate "CN=router1" with the provided CA 
INFO[0000] "172.17.0.100:57400" installing certificate id=cert2 "CN=router1" 
INFO[0000] "172.17.0.100:57400" Install RPC successful  
```

#### Cert Load

Not implemented

#### Cert LoadCa

Not implemented

#### Cert Revoke

The `[ revoke ]` command acts as the client side of the [Cert RevokeCertificates RPC](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L160)

```bash
gnoic -a 172.17.0.100:57400 -u admin -p admin --insecure \
      cert revoke --id cert2
```

```bash
INFO[0000] "172.17.0.100:57400" certificateID=cert2 revoked successfully 
```

Multiple certificates can be revoked at once, check the [command reference](command_reference/cert/revoke.md) for more flags options

#### Cert Rotate

The `[ rotate ]` command acts as the client side of the [Cert Rotate RPC](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto#L93)

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

### File gNOI Service

#### File Get

```bash
gnoic -a 172.17.0.100:57400 --insecure -u admin -p admin \
      file get \
      --file cf3:\license.lic
```

```bash
INFO[0000] "172.17.0.100:57400" received 987 bytes      
INFO[0000] "172.17.0.100:57400" file "cf3:license.lic" saved 
```

#### File Put

```bash
gnoic -a 172.17.0.100:57400 --insecure -u admin -p admin \
      file put \
      --file license.lic --remote-name cf3:\license.lic
```

```bash
INFO[0000] writing 987 byte(s) to "172.17.0.100:57400"  
INFO[0000] sending file hash to "172.17.0.100:57400"    
INFO[0000] "172.17.0.100:57400" file "cf3:license.lic" written successfully 
```

#### File Remove

```bash
gnoic -a 172.17.0.100:57400 --insecure -u admin -p admin file remove --file cf3:\license.lic
```

```bash
INFO[0000] "172.17.0.100:57400" file "cf3:license.lic" removed successfully
```

#### File Stat

```bash
gnoic -a 172.17.0.100:57400 --insecure -u admin -p admin \
      file stat --file cf3:\\
```

```md
+--------------------+-----------------------+---------------------------+------+-------+------+
|    Target Name     |         Path          |       LastModified        | Perm | Umask | Size |
+--------------------+-----------------------+---------------------------+------+-------+------+
| 172.17.0.100:57400 | cf3:\.ssh\            | 2021-05-04T11:33:10+08:00 | 0777 | 0777  | 0    |
| 172.17.0.100:57400 | cf3:\CONFIG.CFG       | 2021-03-31T19:19:38+08:00 | 0777 | 0777  | 0    |
| 172.17.0.100:57400 | cf3:\NVRAM.DAT        | 2021-03-31T19:19:38+08:00 | 0777 | 0777  | 101  |
| 172.17.0.100:57400 | cf3:\SUPPORT\         | 2021-03-31T20:34:48+08:00 | 0777 | 0777  | 0    |
| 172.17.0.100:57400 | cf3:\SYSLINUX\        | 2021-03-31T19:19:38+08:00 | 0777 | 0777  | 0    |
| 172.17.0.100:57400 | cf3:\TIMOS\           | 2021-03-31T19:19:38+08:00 | 0777 | 0777  | 0    |
| 172.17.0.100:57400 | cf3:\bof.cfg          | 2021-05-04T11:34:16+08:00 | 0777 | 0777  | 636  |
| 172.17.0.100:57400 | cf3:\bof.cfg.1        | 2021-03-31T19:19:38+08:00 | 0777 | 0777  | 196  |
| 172.17.0.100:57400 | cf3:\bootlog.txt      | 2021-05-04T11:35:06+08:00 | 0777 | 0777  | 4298 |
| 172.17.0.100:57400 | cf3:\bootlog_prev.txt | 2021-05-04T11:33:10+08:00 | 0777 | 0777  | 4539 |
| 172.17.0.100:57400 | cf3:\license.lic      | 2021-05-04T11:33:58+08:00 | 0777 | 0777  | 987  |
| 172.17.0.100:57400 | cf3:\nvsys.info       | 2021-05-04T11:35:04+08:00 | 0777 | 0777  | 315  |
| 172.17.0.100:57400 | cf3:\restcntr.txt     | 2021-05-04T11:35:04+08:00 | 0777 | 0777  | 1    |
| 172.17.0.100:57400 | cf3:\system-pki\      | 2021-05-06T11:34:24+08:00 | 0777 | 0777  | 0    |
+--------------------+-----------------------+---------------------------+------+-------+------+
```

For a more human readable output use `--humanize` flag

```bash
gnoic -a 172.17.0.100:57400 --insecure -u admin -p admin \
      file stat --file cf3: --humanize
```

```md
+--------------------+-----------------------+----------------+------+-------+--------+
|    Target Name     |         Path          |  LastModified  | Perm | Umask |  Size  |
+--------------------+-----------------------+----------------+------+-------+--------+
| 172.17.0.100:57400 | cf3:\.ssh\            | 2 days ago     | 0777 | 0777  | 0 B    |
| 172.17.0.100:57400 | cf3:\CONFIG.CFG       | 1 month ago    | 0777 | 0777  | 0 B    |
| 172.17.0.100:57400 | cf3:\NVRAM.DAT        | 1 month ago    | 0777 | 0777  | 101 B  |
| 172.17.0.100:57400 | cf3:\SUPPORT\         | 1 month ago    | 0777 | 0777  | 0 B    |
| 172.17.0.100:57400 | cf3:\SYSLINUX\        | 1 month ago    | 0777 | 0777  | 0 B    |
| 172.17.0.100:57400 | cf3:\TIMOS\           | 1 month ago    | 0777 | 0777  | 0 B    |
| 172.17.0.100:57400 | cf3:\bof.cfg          | 2 days ago     | 0777 | 0777  | 636 B  |
| 172.17.0.100:57400 | cf3:\bof.cfg.1        | 1 month ago    | 0777 | 0777  | 196 B  |
| 172.17.0.100:57400 | cf3:\bootlog.txt      | 2 days ago     | 0777 | 0777  | 4.3 kB |
| 172.17.0.100:57400 | cf3:\bootlog_prev.txt | 2 days ago     | 0777 | 0777  | 4.5 kB |
| 172.17.0.100:57400 | cf3:\license.lic      | 2 days ago     | 0777 | 0777  | 987 B  |
| 172.17.0.100:57400 | cf3:\nvsys.info       | 2 days ago     | 0777 | 0777  | 315 B  |
| 172.17.0.100:57400 | cf3:\restcntr.txt     | 2 days ago     | 0777 | 0777  | 1 B    |
| 172.17.0.100:57400 | cf3:\system-pki\      | 47 minutes ago | 0777 | 0777  | 0 B    |
+--------------------+-----------------------+----------------+------+-------+--------+
```

#### File Transfer

Not implemented

### System gNOI Service

#### System CancelReboot

#### System Ping

#### System RebootStatus

#### System Reboot

#### System SetPackage

#### System SwitchControlProcessor

#### System Time

#### System Traceroute
