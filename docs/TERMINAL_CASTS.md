# Terminal Casts

No media recording tool was used for this pass. These are text snapshots of verified commands.

## Version

```text
$ ceo-packet --version
ceo-packet 0.1.0-dev commit=local
```

## Install

```text
$ PREFIX=/tmp/ceo-harness sh scripts/install-local.sh
ceo-packet 0.1.0-dev commit=local
installed /tmp/ceo-harness/bin/ceo-packet
```

## Release Checksum

```text
$ cd dist && shasum -a 256 -c checksums.txt
ceo-packet_0.1.0-dev_darwin_arm64.tar.gz: OK
ceo-packet_0.1.0-dev_linux_amd64.tar.gz: OK
ceo-packet_0.1.0-dev_linux_arm64.tar.gz: OK
```
