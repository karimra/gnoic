# File Stat

### Description

The `file stat` command uses the [gNOI File Stat RPC](https://github.com/openconfig/gnoi/blob/master/file/file.proto#L57) to retrieve metadata about a file or a directory.

It supports a couple of flags:

- `file`: file to get metadata about.
- `humanize`: makes some of the output data human readable, most notably dates and file sizes.
  
### Usage

`gnoic [global-flags] file stat [local-flags]`

### Flags

#### file

The `--file` sets the remote file name to get metadata about.

#### humanize

The `--humanize` flag makes some of the output data human readable, most notably dates and file sizes.

### Examples

```bash
gnoic -a r1,r2,r3 --insecure -u admin -p admin file stat --file cf3:\system-pki --humanize
```

```md
+-------------+---------------------------+--------------+------+-------+--------+
| Target Name |           Path            | LastModified | Perm | Umask |  Size  |
+-------------+---------------------------+--------------+------+-------+--------+
| r1:57400    | cf3:\system-pki\cert3.crt | 4 hours ago  | 0777 | 0777  | 980 B  |
|             | cf3:\system-pki\cert2.key | 4 hours ago  | 0777 | 0777  | 1.3 kB |
|             | cf3:\system-pki\cert2.crt | 4 hours ago  | 0777 | 0777  | 977 B  |
|             | cf3:\system-pki\cert1.key | 4 hours ago  | 0777 | 0777  | 1.3 kB |
|             | cf3:\system-pki\cert1.crt | 4 hours ago  | 0777 | 0777  | 977 B  |
|             | cf3:\system-pki\cert3.key | 4 hours ago  | 0777 | 0777  | 1.3 kB |
| r2:57400    | cf3:\system-pki\cert3.key | 4 hours ago  | 0777 | 0777  | 1.3 kB |
|             | cf3:\system-pki\cert1.crt | 4 hours ago  | 0777 | 0777  | 977 B  |
|             | cf3:\system-pki\cert1.key | 4 hours ago  | 0777 | 0777  | 1.3 kB |
|             | cf3:\system-pki\cert2.crt | 4 hours ago  | 0777 | 0777  | 977 B  |
|             | cf3:\system-pki\cert3.crt | 4 hours ago  | 0777 | 0777  | 980 B  |
|             | cf3:\system-pki\cert2.key | 4 hours ago  | 0777 | 0777  | 1.3 kB |
| r3:57400    | cf3:\system-pki\cert3.key | 4 hours ago  | 0777 | 0777  | 1.3 kB |
|             | cf3:\system-pki\cert3.crt | 4 hours ago  | 0777 | 0777  | 981 B  |
|             | cf3:\system-pki\cert2.key | 4 hours ago  | 0777 | 0777  | 1.3 kB |
|             | cf3:\system-pki\cert2.crt | 4 hours ago  | 0777 | 0777  | 977 B  |
|             | cf3:\system-pki\cert1.key | 4 hours ago  | 0777 | 0777  | 1.3 kB |
|             | cf3:\system-pki\cert1.crt | 4 hours ago  | 0777 | 0777  | 977 B  |
+-------------+---------------------------+--------------+------+-------+--------+
```
