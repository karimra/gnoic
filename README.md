`gnoic` is a gNOI CLI client that provides support for the below gNOI Services:
- [Certificate Managment](https://github.com/openconfig/gnoi/blob/master/cert/cert.proto)
- [Factory Reset](https://github.com/openconfig/gnoi/blob/main/factory_reset/factory_reset.proto)
- [File](https://github.com/openconfig/gnoi/blob/master/file/file.proto)
- [Healthz](https://github.com/openconfig/gnoi/blob/main/healthz/healthz.proto)
- [OS](https://github.com/openconfig/gnoi/blob/main/os/os.proto)
- [System](https://github.com/openconfig/gnoi/blob/master/system/system.proto)

As well as a server implementation of the gNOI [File](https://github.com/openconfig/gnoi/blob/master/file/file.proto) Service.

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
 ├─── completion
 │    ├─── bash
 │    ├─── fish
 │    ├─── powershell
 │    └─── zsh
 ├─── factory-reset
 │    └─── start
 ├─── file
 │    ├─── get
 │    ├─── put
 │    ├─── remove
 │    ├─── stat
 │    └─── transfer
 ├─── healthz
 │    ├─── ack
 │    ├─── artifact
 │    ├─── check
 │    ├─── get
 │    └─── list
 ├─── help [command]
 ├─── os
 │    ├─── activate
 │    ├─── install
 │    └─── verify
 ├─── server
 ├─── system
 │    ├─── cancel-reboot
 │    ├─── kill-process
 │    ├─── ping
 │    ├─── reboot
 │    ├─── reboot-status
 │    ├─── set-package
 │    ├─── switch-control-processor
 │    ├─── time
 │    └─── traceroute
 ├─── tree
 └─── version
      └─── upgrade
```
