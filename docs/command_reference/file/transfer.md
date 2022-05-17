# File Transfer

### Description

The `file transfer` command uses the [gNOI File Transfer RPC](https://github.com/openconfig/gnoi/blob/master/file/file.proto#L41) to transfer a file from a target to a remote location.

It supports 2 flags:

- `local`: the file name local to the target.
- `remote`: the remote location to which the target should transfer the file.

### Usage

`gnoic [global-flags] file transfer [local-flags]`

### Flags

#### local

The `--local` sets the local file path to be transferred from the target to the remote location.

#### remote

The `--remote` flag is used to set the remote location to which the target should transfer the local file.

It should include the protocol used, any credentials (if needed), the remote address and path.

e.g:

- `scp://user:pass@server.com:/path/to/file`
- `scp:/192.168.1.1:/path/to/file`
- `sftp://user:pass@server.com:/path/to/file`
- `http://user:pass@server.com/path/to/file`
- `https://user:pass@server.com/path/to/file`
- `https://server.com/path/to/file`

#### source-address

The `--source-address` flag sets the source address used to initiate connections from the target.
It can be either an IPv4 address or an IPv6 address, depending on the connection's destination address.

### Examples

```bash
gnoic -a clab-gnoi-sr1 --insecure -u admin -p admin file transfer --path cf3:\bof.cfg --remote http://admin:admin@172.100.100.1:8000/bof.cfg
```

```
INFO[0000] sending file transfer request: local_path:"cf3:bof.cfg" remote_download:{path:"172.100.100.1:8000/bof.cfg" protocol:HTTP} to target "clab-gnoi-sr1:57400" 
+---------------------+-------------+----------------------------------+
|     Target Name     | Hash Method |               Hash               |
+---------------------+-------------+----------------------------------+
| clab-gnoi-sr1:57400 | MD5         | ce226201074f995db2c2a88a6948fb51 |
+---------------------+-------------+----------------------------------+
```