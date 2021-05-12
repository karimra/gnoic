# File Stat

### Description

The `file stat` command uses the [gNOI File Stat RPC](https://github.com/openconfig/gnoi/blob/master/file/file.proto#L57) to retrieve metadata about a file or a directory.

It supports 3 flags:

- `path`: path(s) to get metadata about.
- `humanize`: makes some of the output data human readable, most notably dates and file sizes.
- `recursive`: recursively looks up the path subdirectories and runs a Stat RPC against them.
  
### Usage

`gnoic [global-flags] file stat [local-flags]`

### Flags

#### path

The `--path` sets the remote file/dir name to get metadata about.

#### humanize

The `--humanize` flag makes some of the output data human readable, most notably dates and file sizes.

#### recursive

The `--recursive` recursively looks up the path subdirectories and runs a Stat RPC against them.

### Examples

```bash
gnoic -a r1,r2,r3 --insecure -u admin -p admin file stat --path cf3: --humanize
```

```md
+-------------+---------------------------+----------------+------------+------------+--------+
| Target Name |           Path            |  LastModified  |    Perm    |   Umask    |  Size  |
+-------------+---------------------------+----------------+------------+------------+--------+
| r1:57400    | cf3:\system-pki\cert1.crt | 24 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 994 B  |
|             | cf3:\system-pki\cert1.key | 24 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 1.3 kB |
|             | cf3:\system-pki\cert2.crt | 22 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 994 B  |
|             | cf3:\system-pki\cert2.key | 22 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 1.3 kB |
|             | cf3:\system-pki\cert3.crt | 20 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 994 B  |
|             | cf3:\system-pki\cert3.key | 20 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 1.3 kB |
| r2:57400    | cf3:\system-pki\cert1.crt | 24 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 995 B  |
|             | cf3:\system-pki\cert1.key | 24 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 1.3 kB |
|             | cf3:\system-pki\cert2.crt | 22 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 995 B  |
|             | cf3:\system-pki\cert2.key | 22 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 1.3 kB |
|             | cf3:\system-pki\cert3.crt | 20 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 994 B  |
|             | cf3:\system-pki\cert3.key | 20 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 1.3 kB |
| r3:57400    | cf3:\system-pki\cert1.crt | 24 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 995 B  |
|             | cf3:\system-pki\cert1.key | 24 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 1.3 kB |
|             | cf3:\system-pki\cert2.crt | 22 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 995 B  |
|             | cf3:\system-pki\cert2.key | 22 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 1.3 kB |
|             | cf3:\system-pki\cert3.crt | 20 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 995 B  |
|             | cf3:\system-pki\cert3.key | 20 seconds ago | -rwxrwxrwx | -rwxrwxrwx | 1.3 kB |
+-------------+---------------------------+----------------+------------+------------+--------+
```
