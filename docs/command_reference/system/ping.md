# System Ping

## Description

Ping executes the ping command on the target and streams back
the results.  Some targets may not stream any results until all
results are in.  If a packet count is not explicitly provided,
5 is used.

### Usage

`gnoic [global-flags] system ping [local-flags]`

### Flags

#### count

The `--count` flag sets the number of packets.

A count of 0 defaults to a vendor specified value, typically 5.
A count of -1 means continue until the RPC times out or is canceled.

#### destination

The `--destination` flag sets the destination address to ping. required.

#### do-not-fragment

The `--do-not-fragment` flag sets the do not fragment bit. (IPv4 destinations)

#### do-not-resolve

The `--do-not-resolve` flag disables the resolution of the returned addresses.

#### interval

The `--interval` flag sets the interval between requests, e.g 1s, 10ms

#### protocol

The `--protocol` flag sets the layer 3 protocol used to ping, possible values: `v4` or `v6`

#### size

The `--size` flag sets the size of request packet. (excluding ICMP header).

If the size is 0, the vendor default size will be used (typically 56 bytes).

#### source

The `--source` flag sets the source address to ping from.

#### wait

The `--wait`  flag sets the duration to wait for a response.

### Examples

#### simple ping

```shell
gnoic -a clab-gnoi-ceos1:6030 --insecure -u admin -p admin \
       system ping \
       --destination 1.1.1.1
```

```shell
64 bytes from 1.1.1.1: icmp_seq=1 ttl=56 time=2.43ms
64 bytes from 1.1.1.1: icmp_seq=2 ttl=56 time=3.79ms
64 bytes from 1.1.1.1: icmp_seq=3 ttl=56 time=3.78ms
64 bytes from 1.1.1.1: icmp_seq=4 ttl=56 time=3.47ms
64 bytes from 1.1.1.1: icmp_seq=5 ttl=56 time=3.49ms
--- 1.1.1.1 ping statistics ---
5 packets sent, 5 packets received, 0.00% packet loss
round-trip min/avg/max/stddev = 2.438/3.397/3.798/0.504 ms
```

#### json formatted output

```shell
gnoic -a clab-gnoi-ceos1:6030 --insecure -u admin -p admin \
       system ping \
       --destination 1.1.1.1
       --format json
```

```json
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "source": "1.1.1.1",
    "time": 3550000,
    "bytes": 64,
    "sequence": 1,
    "ttl": 56
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "source": "1.1.1.1",
    "time": 3240000,
    "bytes": 64,
    "sequence": 2,
    "ttl": 56
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "source": "1.1.1.1",
    "time": 4610000,
    "bytes": 64,
    "sequence": 3,
    "ttl": 56
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "source": "1.1.1.1",
    "time": 2450000,
    "bytes": 64,
    "sequence": 4,
    "ttl": 56
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "source": "1.1.1.1",
    "time": 4070000,
    "bytes": 64,
    "sequence": 5,
    "ttl": 56
  }
}
{
  "target": "clab-gnoi-ceos1:6030",
  "response": {
    "source": "1.1.1.1",
    "time": 4006000000,
    "sent": 5,
    "received": 5,
    "min_time": 2452000,
    "avg_time": 3587000,
    "max_time": 4611000,
    "std_dev": 734000
  }
}
```

#### multiple targets

```shell
gnoic -a clab-gnoi-ceos1:6030,clab-gnoi-ceos2:6030 --insecure -u admin -p admin \
       system ping \
       --destination 1.1.1.1
```

```shell
[clab-gnoi-ceos1:6030] 64 bytes from 1.1.1.1: icmp_seq=1 ttl=56 time=3.86ms
[clab-gnoi-ceos2:6030] 64 bytes from 1.1.1.1: icmp_seq=1 ttl=56 time=3.89ms
[clab-gnoi-ceos2:6030] 64 bytes from 1.1.1.1: icmp_seq=2 ttl=56 time=4.69ms
[clab-gnoi-ceos1:6030] 64 bytes from 1.1.1.1: icmp_seq=2 ttl=56 time=3.8ms
[clab-gnoi-ceos2:6030] 64 bytes from 1.1.1.1: icmp_seq=3 ttl=56 time=4.2ms
[clab-gnoi-ceos1:6030] 64 bytes from 1.1.1.1: icmp_seq=3 ttl=56 time=3.09ms
[clab-gnoi-ceos1:6030] 64 bytes from 1.1.1.1: icmp_seq=4 ttl=56 time=3.66ms
[clab-gnoi-ceos2:6030] 64 bytes from 1.1.1.1: icmp_seq=4 ttl=56 time=4.68ms
[clab-gnoi-ceos1:6030] 64 bytes from 1.1.1.1: icmp_seq=5 ttl=56 time=3.33ms
[clab-gnoi-ceos2:6030] 64 bytes from 1.1.1.1: icmp_seq=5 ttl=56 time=3.47ms
[clab-gnoi-ceos2:6030] --- 1.1.1.1 ping statistics ---
[clab-gnoi-ceos2:6030] 5 packets sent, 5 packets received, 0.00% packet loss
[clab-gnoi-ceos2:6030] round-trip min/avg/max/stddev = 3.471/4.188/4.693/0.469 ms
[clab-gnoi-ceos1:6030] --- 1.1.1.1 ping statistics ---
[clab-gnoi-ceos1:6030] 5 packets sent, 5 packets received, 0.00% packet loss
[clab-gnoi-ceos1:6030] round-trip min/avg/max/stddev = 3.090/3.551/3.861/0.296 ms
```
