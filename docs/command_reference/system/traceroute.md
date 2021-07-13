# System Traceroute

## Description

Traceroute executes the traceroute command on the target and streams back the results.
Some targets may not stream any results until all results are in.  
If a hop count is not explicitly provided, 30 is used.

### Usage

`gnoic [global-flags] system traceroute [local-flags]`

### Flags

#### destination

The `--destination` flag sets the destination address of the traceroute command.

#### do-not-fragment

The `--do-not-fragment` flag sets the do not fragment bit. (IPv4 destinations)

#### do-not-resolve

The `--do-not-resolve` flag disables the resolution of the returned addresses.

#### initial-ttl

The `--initial-ttl` flag sets the initial TTL the packets start with, default=1

#### l3protocol

The `-3 | --l3protocol` flag sets the layer 3 protocol used to traceroute, possible values: `v4` or `v6`

#### l4protocol

The `-4 | --l4protocol` flag sets the layer 4 protocol used to traceroute, possible `ICMP`, `UDP` or `TCP`. Default `ICMP`

#### max-ttl

The `--max-ttl` flag sets the maximum number of hops. (default=30)

#### source

The `--source` flag sets the source address to traceroute from.

#### wait

The `--wait`  flag sets the duration to wait for a response.

### Examples

#### simple traceroute

```shell
gnoic -a clab-gnoi-ceos1:6030 --insecure -u admin -p admin \
    system traceroute \
    --destination 1.1.1.1
```

```shell
1 172-11-11-1.lightspeed.chrlnc.sbcglobal.net (172.11.11.1) 43µs
1 172-11-11-1.lightspeed.chrlnc.sbcglobal.net (172.11.11.1) 5µs
1 172-11-11-1.lightspeed.chrlnc.sbcglobal.net (172.11.11.1) 4µs
2 192.168.1.1 (192.168.1.1) 2.196ms
2 192.168.1.1 (192.168.1.1) 2.568ms
2 192.168.1.1 (192.168.1.1) 2.974ms
3 h254.s98.ts.hinet.net (168.95.98.254) [AS3462] 13.179ms
3 h254.s98.ts.hinet.net (168.95.98.254) [AS3462] 13.178ms
3 h254.s98.ts.hinet.net (168.95.98.254) [AS3462] 13.208ms
4 tpdb-3332.hinet.net (168.95.83.102) [AS3462] 3.707ms
4 tpdb-3332.hinet.net (168.95.83.102) [AS3462] 3.789ms
4 tpdb-3332.hinet.net (168.95.83.102) [AS3462] 3.874ms
5 tpdt-3032.hinet.net (220.128.4.142) [AS3462] <MPLS:L=24304,E=0,S=1,T=1> 3.53ms
5 tpdt-3032.hinet.net (220.128.4.142) [AS3462] <MPLS:L=24304,E=0,S=1,T=1> 3.528ms
5 tpdt-3032.hinet.net (220.128.4.142) [AS3462] <MPLS:L=24304,E=0,S=1,T=1> 3.6ms
6 tpdt-3012.hinet.net (220.128.25.89) [AS3462] <MPLS:L=24306,E=0,S=1,T=1> 3.987ms
6 tpdt-3012.hinet.net (220.128.25.89) [AS3462] <MPLS:L=24306,E=0,S=1,T=1> 5.5ms
6 tpdt-3012.hinet.net (220.128.25.89) [AS3462] <MPLS:L=24306,E=0,S=1,T=1> 5.588ms
7 220-128-4-193.HINET-IP.hinet.net (220.128.4.193) [AS3462] 2.566ms
7 220-128-4-193.HINET-IP.hinet.net (220.128.4.193) [AS3462] 3.791ms
7 220-128-4-193.HINET-IP.hinet.net (220.128.4.193) [AS3462] 3.752ms
8 210-242-214-45.HINET-IP.hinet.net (210.242.214.45) [AS3462] 4.749ms
8 210-242-214-45.HINET-IP.hinet.net (210.242.214.45) [AS3462] 4.853ms
8 210-242-214-45.HINET-IP.hinet.net (210.242.214.45) [AS3462] 4.949ms
9 one.one.one.one (1.1.1.1) [AS13335] 4.236ms
9 one.one.one.one (1.1.1.1) [AS13335] 4.47ms
9 one.one.one.one (1.1.1.1) [AS13335] 4.465ms
```

#### json formatted output

```shell
gnoic -a clab-gnoi-ceos1:6030 --insecure -u admin -p admin \
    system traceroute \
    --destination 1.1.1.1
    --format json
```

```json
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "destination_name": "1.1.1.1",
    "destination_address": "1.1.1.1",
    "hops": 30,
    "packet_size": 60
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 1,
    "address": "172.11.11.1",
    "name": "172-11-11-1.lightspeed.chrlnc.sbcglobal.net",
    "rtt": 56000
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 1,
    "address": "172.11.11.1",
    "name": "172-11-11-1.lightspeed.chrlnc.sbcglobal.net",
    "rtt": 5000
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 1,
    "address": "172.11.11.1",
    "name": "172-11-11-1.lightspeed.chrlnc.sbcglobal.net",
    "rtt": 4000
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 2,
    "address": "192.168.1.1",
    "name": "192.168.1.1",
    "rtt": 2146000
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 2,
    "address": "192.168.1.1",
    "name": "192.168.1.1",
    "rtt": 2515000
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 2,
    "address": "192.168.1.1",
    "name": "192.168.1.1",
    "rtt": 2885000
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 3,
    "address": "168.95.98.254",
    "name": "h254.s98.ts.hinet.net",
    "rtt": 6850000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 3,
    "address": "168.95.98.254",
    "name": "h254.s98.ts.hinet.net",
    "rtt": 6848000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 3,
    "address": "168.95.98.254",
    "name": "h254.s98.ts.hinet.net",
    "rtt": 7022000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 4,
    "address": "168.95.83.102",
    "name": "tpdb-3332.hinet.net",
    "rtt": 3365000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 4,
    "address": "168.95.83.102",
    "name": "tpdb-3332.hinet.net",
    "rtt": 3462000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 4,
    "address": "168.95.83.102",
    "name": "tpdb-3332.hinet.net",
    "rtt": 3561000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 5,
    "address": "220.128.4.142",
    "name": "tpdt-3032.hinet.net",
    "rtt": 3046000,
    "mpls": {
      "E": "0",
      "Label": "24304",
      "S": "1",
      "TTL": "1"
    },
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 5,
    "address": "220.128.4.142",
    "name": "tpdt-3032.hinet.net",
    "rtt": 3149000,
    "mpls": {
      "E": "0",
      "Label": "24304",
      "S": "1",
      "TTL": "1"
    },
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 5,
    "address": "220.128.4.142",
    "name": "tpdt-3032.hinet.net",
    "rtt": 3262000,
    "mpls": {
      "E": "0",
      "Label": "24304",
      "S": "1",
      "TTL": "1"
    },
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 6,
    "address": "220.128.25.89",
    "name": "tpdt-3012.hinet.net",
    "rtt": 4975000,
    "mpls": {
      "E": "0",
      "Label": "24306",
      "S": "1",
      "TTL": "1"
    },
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 6,
    "address": "220.128.25.89",
    "name": "tpdt-3012.hinet.net",
    "rtt": 3943000,
    "mpls": {
      "E": "0",
      "Label": "24306",
      "S": "1",
      "TTL": "1"
    },
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 6,
    "address": "220.128.25.89",
    "name": "tpdt-3012.hinet.net",
    "rtt": 7836000,
    "mpls": {
      "E": "0",
      "Label": "24306",
      "S": "1",
      "TTL": "1"
    },
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 7,
    "address": "220.128.4.193",
    "name": "220-128-4-193.HINET-IP.hinet.net",
    "rtt": 3936000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 7,
    "address": "220.128.4.193",
    "name": "220-128-4-193.HINET-IP.hinet.net",
    "rtt": 1772000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 7,
    "address": "220.128.4.193",
    "name": "220-128-4-193.HINET-IP.hinet.net",
    "rtt": 1752000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 8,
    "address": "210.242.214.45",
    "name": "210-242-214-45.HINET-IP.hinet.net",
    "rtt": 2907000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 8,
    "address": "210.242.214.45",
    "name": "210-242-214-45.HINET-IP.hinet.net",
    "rtt": 5204000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 8,
    "address": "210.242.214.45",
    "name": "210-242-214-45.HINET-IP.hinet.net",
    "rtt": 5285000,
    "as_path": [
      3462
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 9,
    "address": "1.1.1.1",
    "name": "one.one.one.one",
    "rtt": 4837000,
    "as_path": [
      13335
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 9,
    "address": "1.1.1.1",
    "name": "one.one.one.one",
    "rtt": 4834000,
    "as_path": [
      13335
    ]
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "hop": 9,
    "address": "1.1.1.1",
    "name": "one.one.one.one",
    "rtt": 4873000,
    "as_path": [
      13335
    ]
  }
}
```

#### multiple targets traceroute

```shell
gnoic -a clab-gnoi-ceos1:6030,clab-gnoi-ceos2:6030 --insecure -u admin -p admin \
    system traceroute \
    --destination 1.1.1.1
```

```shell
[clab-gnoi-ceos2:6030] Traceroute to 1.1.1.1 (1.1.1.1), 30 max hops, 60 byte packet size
[clab-gnoi-ceos1:6030] Traceroute to 1.1.1.1 (1.1.1.1), 30 max hops, 60 byte packet size
[clab-gnoi-ceos2:6030] 1 172-11-11-1.lightspeed.chrlnc.sbcglobal.net (172.11.11.1) 53µs
[clab-gnoi-ceos2:6030] 1 172-11-11-1.lightspeed.chrlnc.sbcglobal.net (172.11.11.1) 6µs
[clab-gnoi-ceos2:6030] 1 172-11-11-1.lightspeed.chrlnc.sbcglobal.net (172.11.11.1) 4µs
[clab-gnoi-ceos1:6030] 1 172-11-11-1.lightspeed.chrlnc.sbcglobal.net (172.11.11.1) 24µs
[clab-gnoi-ceos1:6030] 1 172-11-11-1.lightspeed.chrlnc.sbcglobal.net (172.11.11.1) 6µs
[clab-gnoi-ceos1:6030] 1 172-11-11-1.lightspeed.chrlnc.sbcglobal.net (172.11.11.1) 5µs
[clab-gnoi-ceos2:6030] 2 192.168.1.1 (192.168.1.1) 3.521ms
[clab-gnoi-ceos2:6030] 2 192.168.1.1 (192.168.1.1) 3.932ms
[clab-gnoi-ceos2:6030] 2 192.168.1.1 (192.168.1.1) 4.284ms
[clab-gnoi-ceos1:6030] 2 192.168.1.1 (192.168.1.1) 3.981ms
[clab-gnoi-ceos1:6030] 2 192.168.1.1 (192.168.1.1) 4.385ms
[clab-gnoi-ceos1:6030] 2 192.168.1.1 (192.168.1.1) 4.715ms
[clab-gnoi-ceos1:6030] 3 h254.s98.ts.hinet.net (168.95.98.254) [AS3462] 10.295ms
[clab-gnoi-ceos1:6030] 3 h254.s98.ts.hinet.net (168.95.98.254) [AS3462] 10.391ms
[clab-gnoi-ceos1:6030] 3 h254.s98.ts.hinet.net (168.95.98.254) [AS3462] 10.487ms
[clab-gnoi-ceos2:6030] 3 h254.s98.ts.hinet.net (168.95.98.254) [AS3462] 10.666ms
[clab-gnoi-ceos2:6030] 3 h254.s98.ts.hinet.net (168.95.98.254) [AS3462] 10.769ms
[clab-gnoi-ceos2:6030] 3 h254.s98.ts.hinet.net (168.95.98.254) [AS3462] 10.911ms
[clab-gnoi-ceos1:6030] 4 tpdb-3332.hinet.net (168.95.83.102) [AS3462] 6.396ms
[clab-gnoi-ceos1:6030] 4 tpdb-3332.hinet.net (168.95.83.102) [AS3462] 6.492ms
[clab-gnoi-ceos1:6030] 4 tpdb-3332.hinet.net (168.95.83.102) [AS3462] 6.59ms
[clab-gnoi-ceos2:6030] 4 tpdb-3332.hinet.net (168.95.83.102) [AS3462] 6.571ms
[clab-gnoi-ceos2:6030] 4 tpdb-3332.hinet.net (168.95.83.102) [AS3462] 6.751ms
[clab-gnoi-ceos2:6030] 4 tpdb-3332.hinet.net (168.95.83.102) [AS3462] 6.958ms
[clab-gnoi-ceos1:6030] 5 tpdt-3032.hinet.net (220.128.4.142) [AS3462] <MPLS:L=24304,E=0,S=1,T=1> 5.578ms
[clab-gnoi-ceos1:6030] 5 tpdt-3032.hinet.net (220.128.4.142) [AS3462] <MPLS:L=24304,E=0,S=1,T=1> 5.677ms
[clab-gnoi-ceos1:6030] 5 tpdt-3032.hinet.net (220.128.4.142) [AS3462] <MPLS:L=24304,E=0,S=1,T=1> 5.773ms
[clab-gnoi-ceos2:6030] 5 tpdt-3032.hinet.net (220.128.4.142) [AS3462] <MPLS:L=24304,E=0,S=1,T=1> 5.964ms
[clab-gnoi-ceos2:6030] 5 tpdt-3032.hinet.net (220.128.4.142) [AS3462] <MPLS:L=24304,E=0,S=1,T=1> 6.063ms
[clab-gnoi-ceos2:6030] 5 tpdt-3032.hinet.net (220.128.4.142) [AS3462] <MPLS:L=24304,E=0,S=1,T=1> 6.184ms
[clab-gnoi-ceos1:6030] 6 tpdt-3012.hinet.net (220.128.25.89) [AS3462] <MPLS:L=24306,E=0,S=1,T=1> 6.183ms
[clab-gnoi-ceos1:6030] 6 tpdt-3012.hinet.net (220.128.25.89) [AS3462] <MPLS:L=24306,E=0,S=1,T=1> 4.017ms
[clab-gnoi-ceos1:6030] 6 tpdt-3012.hinet.net (220.128.25.89) [AS3462] <MPLS:L=24306,E=0,S=1,T=1> 4.099ms
[clab-gnoi-ceos2:6030] 6 tpdt-3012.hinet.net (220.128.25.89) [AS3462] <MPLS:L=24306,E=0,S=1,T=1> 6.645ms
[clab-gnoi-ceos2:6030] 6 tpdt-3012.hinet.net (220.128.25.89) [AS3462] <MPLS:L=24306,E=0,S=1,T=1> 3.423ms
[clab-gnoi-ceos2:6030] 6 tpdt-3012.hinet.net (220.128.25.89) [AS3462] <MPLS:L=24306,E=0,S=1,T=1> 3.508ms
[clab-gnoi-ceos1:6030] 7 220-128-4-193.HINET-IP.hinet.net (220.128.4.193) [AS3462] 3.293ms
[clab-gnoi-ceos1:6030] 7 220-128-4-193.HINET-IP.hinet.net (220.128.4.193) [AS3462] 12.711ms
[clab-gnoi-ceos1:6030] 7 220-128-4-193.HINET-IP.hinet.net (220.128.4.193) [AS3462] 12.814ms
[clab-gnoi-ceos2:6030] 7 220-128-4-193.HINET-IP.hinet.net (220.128.4.193) [AS3462] 2.629ms
[clab-gnoi-ceos2:6030] 7 220-128-4-193.HINET-IP.hinet.net (220.128.4.193) [AS3462] 3.227ms
[clab-gnoi-ceos2:6030] 7 220-128-4-193.HINET-IP.hinet.net (220.128.4.193) [AS3462] 3.312ms
[clab-gnoi-ceos1:6030] 8 210-242-214-45.HINET-IP.hinet.net (210.242.214.45) [AS3462] 4.61ms
[clab-gnoi-ceos1:6030] 8 210-242-214-45.HINET-IP.hinet.net (210.242.214.45) [AS3462] 4.698ms
[clab-gnoi-ceos1:6030] 8 210-242-214-45.HINET-IP.hinet.net (210.242.214.45) [AS3462] 4.812ms
[clab-gnoi-ceos2:6030] 8 210-242-214-45.HINET-IP.hinet.net (210.242.214.45) [AS3462] 4.18ms
[clab-gnoi-ceos2:6030] 8 210-242-214-45.HINET-IP.hinet.net (210.242.214.45) [AS3462] 4.222ms
[clab-gnoi-ceos2:6030] 8 210-242-214-45.HINET-IP.hinet.net (210.242.214.45) [AS3462] 4.516ms
[clab-gnoi-ceos2:6030] 9 one.one.one.one (1.1.1.1) [AS13335] 3.619ms
[clab-gnoi-ceos2:6030] 9 one.one.one.one (1.1.1.1) [AS13335] 3.86ms
[clab-gnoi-ceos2:6030] 9 one.one.one.one (1.1.1.1) [AS13335] 3.935ms
[clab-gnoi-ceos1:6030] 9 one.one.one.one (1.1.1.1) [AS13335] 4.375ms
[clab-gnoi-ceos1:6030] 9 one.one.one.one (1.1.1.1) [AS13335] 4.259ms
[clab-gnoi-ceos1:6030] 9 No Response
```
