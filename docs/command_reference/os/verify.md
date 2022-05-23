# OS Verify

### Description

The `os verify` command allows to check the running SW version on the target(s).

It uses the [gNOI OS Verify RPC](https://github.com/openconfig/gnoi/blob/master/os/os.proto#L120), which is a unary RPC.

### Usage

`gnoic [global-flags] os verify [local-flags]`

### Examples

```bash
gnoic -a 172.17.0.100:57400 --skip-verify -u admin -p admin \
      os verify
```
