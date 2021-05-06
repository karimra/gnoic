
# Global Flags

### address

The address flag `[-a | --address]` is used to specify the target's gNOI server address in address:port format, for e.g: `192.168.113.11:57400`

Multiple target addresses can be specified, either as comma separated values:

```bash
gnoic --address 192.168.113.11:57400,192.168.113.12:57400 
```

or by using the `--address` flag multiple times:

```bash
gnoic -a 192.168.113.11:57400 --address 192.168.113.12:57400
```

### debug

The debug flag `[-d | --debug]` enables the printing of extra information when sending/receiving an RPC

### gzip

The `[--gzip]` flag enables gRPC gzip compression.

### insecure

The insecure flag `[--insecure]` is used to indicate that the client wishes to establish an non-TLS enabled gRPC connection.

To disable certificate validation in a TLS-enabled connection use [`skip-verify`](#skip-verify) flag.

<!-- ### log
The `--log` flag enables log messages to appear on stderr output. By default logging is disabled. -->

### password

The password flag `[-p | --password]` is used to specify the target password as part of the user credentials. If omitted, the password input prompt is used to provide the password.

Note that in case multiple targets are used, all should use the same credentials.

### proxy-from-env

The proxy-from-env flag `[--proxy-from-env]` indicates that gNOIc should use the HTTP/HTTPS proxy addresses defined in the environment variables `http_proxy` and `https_proxy` to reach the targets specified using the `--address` flag.

### skip-verify

The skip verify flag `[--skip-verify]` indicates that the target should skip the signature verification steps, in case a secure connection is used.  

### timeout

The timeout flag `[--timeout]` specifies the gRPC timeout after which the connection attempt fails.

Valid formats: 10s, 1m30s, 1h.  Defaults to 10s

### tls-ca

The TLS CA flag `[--tls-ca]` specifies the root certificates for verifying server certificates encoded in PEM format.

### tls-cert

The tls cert flag `[--tls-cert]` specifies the public key for the client encoded in PEM format.

### tls-key

The tls key flag `[--tls-key]` specifies the private key for the client encoded in PEM format.

### tls-max-version

The tls max version flag `[--tls-max-version]` specifies the maximum supported TLS version supported by gNOIc when creating a secure gRPC connection.

### tls-min-version

The tls min version flag `[--tls-min-version]` specifies the minimum supported TLS version supported by gNOIc when creating a secure gRPC connection.

### tls-version

The tls version flag `[--tls-version]` specifies a single supported TLS version gNOIc when creating a secure gRPC connection.

This flag overwrites the previously listed flags `--tls-max-version` and `--tls-min-version`.

### username

The username flag `[-u | --username]` is used to specify the target username as part of the user credentials. If omitted, the input prompt is used to provide the username.
