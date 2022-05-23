# File Put

### Description

The `file put` command allows to push a file from the local system to multiple Target(s).

It uses the [gNOI File Put RPC](https://github.com/openconfig/gnoi/blob/master/file/file.proto#L52), which is a client streaming RPC.

The client start with `open` message containing the remote file name (must be an absolute path) as well as the permissions intended for the file.

It continues by sending the file to the target(s) in a series of chunks of maximum 64KB and finishes with a message containing the Hash of the file.

The `file put` command supports 5 flags:

- `file`: file to put on the target(s).
- `dst`: file/directory remote name. If `--file` points to a directory or multiple `--file` flags are set, `--dst` represents the destination directory, otherwise it represents the remote file name.
- `chunk-size`: chunk write size in Bytes, default (64KB) is used if set to 0.
- `permission`: file permissions, in octal format. If set to 0, the local system file permissions are used. Default: 0777
- `hash-method`: one of MD5, SHA256 or SHA512. If another value is supplied MD5 is used.

### Usage

`gnoic [global-flags] file put [local-flags]`

### Flags

#### file

The `--file` defines the file remote name.

#### dst

The `dst` defines the remote destination. If `--file` points to a directory or multiple `--file` flags are set, `--dst` represents the destination directory, otherwise it represents the remote file name.

#### chunk-size

The `--chunk-size` defines the chunk size in Bytes, the default (64KB) is used if set to 0.

#### has-method

The `--hash-method` sets the method to be used to calculate the hash to be sent to the target(s)

### Examples

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
