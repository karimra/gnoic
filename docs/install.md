# Installation

`gnoic` is a single binary built for the Linux, Mac OS and Windows platforms distributed via [Github releases](https://github.com/karimra/gnoic/releases).

### Linux/Mac OS

#### Packages

### Docker

The `gnoic` container image can be pulled from GitHub container registries. The tag of the image corresponds to the release version and `latest` tag points to the latest available release:

```bash
# pull latest release from github registry
docker pull ghcr.io/karimra/gnoic:latest
# pull a specific release from github registry
docker pull ghcr.io/karimra/gnoic:0.0.2
```

Example running `gnoic tree` command using the docker image:

```bash
docker run \
       --network host \
       --rm ghcr.io/karimra/gnoic tree
```
