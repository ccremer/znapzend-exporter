## Verify

1. Set `tag=<release-version>`
2. Verify checksums and binary
```bash
link=https://github.com/ccremer/znapzend-exporter/releases/download/$tag
wget -q https://raw.githubusercontent.com/ccremer/znapzend-exporter/$tag/signature.asc && \
wget -q $link/znapzend-exporter_linux_amd64 && \
wget -q $link/checksums.txt.sig && \
wget -q $link/checksums.txt && \
gpg --import signature.asc && gpg --verify checksums.txt.sig && \
grep "$(sha256sum znapzend-exporter_linux_amd64)" checksums.txt
```
