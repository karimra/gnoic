# OS Activate

### Description

The `os activate` command sets the request OS version to be used at the next reboot. It reboots the target if the `--no-boot` flag is not set.

It uses the [gNOI OS Activate RPC](https://github.com/openconfig/gnoi/blob/master/os/os.proto#L116), which is a unary RPC.
The client sends an [`ActivateRequest`](https://github.com/openconfig/gnoi/blob/master/os/os.proto#L253) message stating the version to be activated, if it needs to be activated on the standby supervisor and if the target should be rebooted right after setting the new active SW version.

The `os activate` command supports 3 flags:

- `version`: The OS package version.
- `standby`: If present, the package is activated in the standby supervisor.
- `no-reboot`: If present, the target is not rebooted after setting the active SW version.

### Usage

`gnoic [global-flags] os install [local-flags]`

### Flags

#### version

The  `--version` flag sets the package version to be transferred.

#### standby

The `--standby` flag, if present, instructs the target to activate the package in the standby supervisor.

#### no-reboot

The `--no-reboot` flag, if present, instructs the target to not reboot after setting the SW version as active.

### Examples

```bash
gnoic -a 172.17.0.100:57400 --skip-verify -u admin -p admin \
      os activate \
      --version srlinux_22.6.0 \
      --no-reboot
```
