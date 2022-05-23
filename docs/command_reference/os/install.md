# OS Install

### Description

The `os install` command allows to transfer an OS package to the target(s).

It uses the [gNOI OS Install RPC](https://github.com/openconfig/gnoi/blob/master/os/os.proto#L110), which is a bidirectional streaming RPC.
The client starts by sending a TransferRequest stating the OS package version and whether the to perform the transfer with the standby supervisor as target.

If the Server already has the OS package it answers with [`Validated`](https://github.com/openconfig/gnoi/blob/master/os/os.proto#L206).
Otherwise it answers with [`TransferReady`](https://github.com/openconfig/gnoi/blob/master/os/os.proto#L185), at which point the client can start the actual transfer of the package by sending a stream of [`TransferContent`](https://github.com/openconfig/gnoi/blob/master/os/os.proto#L126) messages.

During the `TransferContent` phase, the server may send `TransferProgress` messages containing the number of bytes already received.

Once the client is done transferring the OS package file, it sends a [`TransferEnd`](https://github.com/openconfig/gnoi/blob/master/os/os.proto#L166) informing the target that the file transfer is done. The target Must perform a general health check to the OS package and return an [`InstallError`](https://github.com/openconfig/gnoi/blob/master/os/os.proto#L218) message with the appropriate error [type](https://github.com/openconfig/gnoi/blob/master/os/os.proto#L219)

If the health check succeeds, the target must response with a [`Validated`](https://github.com/openconfig/gnoi/blob/master/os/os.proto#L206) message.

The `os install` command supports 4 flags:

- `version`: The OS package version.
- `standby`: If present, the package is transferred towards the standby supervisor.
- `pkg`: Local path to the package file.
- `content-chunk-size`: the maximum number of bytes to transfer in a `TransferContent` message, defaults to 1MB

### Usage

`gnoic [global-flags] os install [local-flags]`

### Flags

#### version

The  `--version` flag sets the package version to be transferred.

#### standby

The `--standby` flag if present, instructs the target that the package should be transferred towards the standby supervisor.

#### pkg

The `--pkg` flag points to the OS package file to be transferred.

#### content-chunk-size

The `--content-chunk-size` flag sets the maximum number of bytes to transfer in a `TransferContent` message, defaults to 1MB.

### Examples

```bash
gnoic -a 172.17.0.100:57400 --skip-verify -u admin -p admin \
      os install \
      --version srlinux_22.6.0
      --pkg srlinux_22.6.0.zip
```
