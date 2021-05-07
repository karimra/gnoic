
# File Remove

### Description

The `file remove` command uses the [gNOI File Remove RPC](https://github.com/openconfig/gnoi/blob/master/file/file.proto#L62) to remove a file from one or multiple targets.

### Usage

`gnoic [global-flags] file remove [local-flags]`

### Flags

#### file

The `--file` specifies the file name to removed from the target(s).

### Examples

```bash
gnoic -a 172.17.0.100:57400 --insecure -u admin -p admin file remove --file cf3:\license.lic
```

```bash
INFO[0000] "172.17.0.100:57400" file "cf3:license.lic" removed successfully
```
