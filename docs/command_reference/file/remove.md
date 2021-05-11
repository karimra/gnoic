
# File Remove

### Description

The `file remove` command uses the [gNOI File Remove RPC](https://github.com/openconfig/gnoi/blob/master/file/file.proto#L62) to remove a file from one or multiple targets.

It supports one flag `--path` that points to a remote file or directory to be removed.
In case the remote path is a directory, `gnoic` will walk the path recursively and remove contained files one by one.

### Usage

`gnoic [global-flags] file remove [local-flags]`

### Flags

#### path

The `--path` specifies the file name or directory name to removed from the target(s).
Can be set multiple times.

### Examples

```bash
gnoic -a 172.17.0.100:57400 --insecure -u admin -p admin file remove --path cf3:\license.lic
```

```bash
INFO[0000] "172.17.0.100:57400" file "cf3:license.lic" removed successfully
```
