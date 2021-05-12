# File Get

### Description

The `file get` command allows to retrieve a file from multiple Target(s) and store it in the local system.

It uses the [gNOI File Get RPC](https://github.com/openconfig/gnoi/blob/master/file/file.proto#L34), which is a server streaming gRPC. After receiving the request, the Target streams the file to the client in chunks of maximum 64KB and finishes with a message containing the Hash of the file.
At which point, the client should verify the Hash and save the file locally

The `file get` command supports 3 flags:

- `file`: the file(s)/dir(s) names to be retrieved.
- `dst`: local directory name, defaults to `$PWD`.
- `target-prefix`: if present, `gNOIc` prepends the target name to the retrieved file.

### Usage

`gnoic [global-flags] file get [local-flags]`

### Flags

#### file

The  `--file` flag sets the file(s)/dir(s) names to be retrieved.

#### dst

The `--dst` flag sets the local directory name, defaults to `$PWD`.

#### target-prefix

If present, the `--target-prefix` prepends the target name to the name the file will be stored under.

### Examples

```bash
gnoic -a 172.17.0.100:57400 --insecure -u admin -p admin \
      file get \
      --file cf3:\license.lic
```

```bash
INFO[0000] "172.17.0.100:57400" received 987 bytes      
INFO[0000] "172.17.0.100:57400" file "cf3:license.lic" saved 
```
