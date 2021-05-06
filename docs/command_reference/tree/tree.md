# Tree

### Description

The `tree` command is a convinient method to display `gNOIc`'s command tree.

It takes a couple of flags:

- `flat`: display full commands instead of a tree
- `details`: display flags together with the command tree

Both flags cannot be combined.

### Usage

`gnoic tree [local-flags]`

### Flags

#### flat

If present, `gNOIc` displays a list of commands rather than a tree.

#### details

if present, `gNOIc` displays each command flags as part of the tree.

### Examples

```bash
gnoic tree
```

```md
gnoic
 ├─── cert
 │    ├─── can-generate-csr
 │    ├─── create-ca
 │    ├─── generate-csr
 │    ├─── get-certs
 │    ├─── install
 │    ├─── load
 │    ├─── load-ca
 │    ├─── revoke
 │    └─── rotate
 ├─── file
 │    ├─── get
 │    ├─── put
 │    ├─── remove
 │    ├─── stat
 │    └─── transfer
 ├─── help [command]
 ├─── system
 │    ├─── cancel-reboot
 │    ├─── ping
 │    ├─── reboot
 │    ├─── reboot-status
 │    ├─── set-package
 │    ├─── switch-control-processor
 │    ├─── time
 │    └─── traceroute
 ├─── tree
 └─── version
```

```bash
gnoic tree --flat
```

```md
gnoic
 gnoic cert
 gnoic cert can-generate-csr
 gnoic cert create-ca
 gnoic cert generate-csr
 gnoic cert get-certs
 gnoic cert install
 gnoic cert load
 gnoic cert load-ca
 gnoic cert revoke
 gnoic cert rotate
 gnoic file
 gnoic file get
 gnoic file put
 gnoic file remove
 gnoic file stat
 gnoic file transfer
 gnoic help [command]
 gnoic system
 gnoic system cancel-reboot
 gnoic system ping
 gnoic system reboot
 gnoic system reboot-status
 gnoic system set-package
 gnoic system switch-control-processor
 gnoic system time
 gnoic system traceroute
 gnoic tree
 gnoic version
```

```bash
gnoic tree --details
```

```md
gnoic
 ├─── cert
 │    ├─── can-generate-csr [--cert-type] [--key-size] [--key-type]
 │    ├─── create-ca [--cert-out] [--common-name] [--country] [--email] [--key-out] [--key-size] [--locality] [--org] [--org-unit] [--postal-code] [--state] [--street-address] [--validity]
 │    ├─── generate-csr [--cert-type] [--city] [--common-name] [--country] [--email-id] [--id] [--ip-address] [--key-type] [--min-key-size] [--org] [--org-unit] [--state]
 │    ├─── get-certs [--details] [--id] [--save]
 │    ├─── install [--cert-type] [--city] [--common-name] [--country] [--email-id] [--id] [--ip-address] [--key-type] [--min-key-size] [--org] [--org-unit] [--print-csr] [--state] [--validity]
 │    ├─── load
 │    ├─── load-ca
 │    ├─── revoke [--all] [--id]
 │    └─── rotate [--cert-type] [--city] [--common-name] [--country] [--email-id] [--id] [--ip-address] [--key-type] [--min-key-size] [--org] [--org-unit] [--print-csr] [--state] [--validity]
 ├─── file
 │    ├─── get [--file] [--local-file] [--target-prefix]
 │    ├─── put [--chunk-size] [--file] [--hash-method] [--permission] [--remote-name]
 │    ├─── remove [--file]
 │    ├─── stat [--file] [--humanize]
 │    └─── transfer
 ├─── help [command]
 ├─── system
 │    ├─── cancel-reboot [--message] [--subcomponent]
 │    ├─── ping [--count] [--destination] [--do-not-fragment] [--do-not-resolve] [--interval] [--protocol] [--size] [--source] [--wait]
 │    ├─── reboot [--delay] [--force] [--message] [--method] [--subcomponent]
 │    ├─── reboot-status [--subcomponent]
 │    ├─── set-package
 │    ├─── switch-control-processor [--path]
 │    ├─── time
 │    └─── traceroute [--destination] [--do-not-fragment] [--do-not-resolve] [--initial-ttl] [--interval] [-3 | --l3protocol] [-4 | --l4protocol] [--max-ttl] [--size] [--source] [--wait]
 ├─── tree [--details] [--flat] [-h | --help]
 └─── version
```
