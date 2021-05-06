`gnoic` is a gNOI CLI client that provides support for gNOI [Certificate Managment](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto), 
[File](https://github.com/openconfig/gnoi/blob/master/file/file.proto) and [System](https://github.com/openconfig/gnoi/blob/master/system/system.proto) Services.

Documentation available at [https://gnoic.kmrd.dev](https://gnoic.kmrd.dev)

## Quick start guide

### Installation
```
bash -c "$(curl -sL https://get-gnoic.kmrd.dev)"
```

### Tree Command

```bash
gnoic tree
```

```md
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
