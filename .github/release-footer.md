## Verify

```bash
link=https://github.com/ccremer/znapzend-exporter/releases/download/{{ .Tag }}
wget -q https://raw.githubusercontent.com/ccremer/znapzend-exporter/{{ .Tag }}/signature.asc && \
wget -q $link/znapzend-exporter_linux_amd64 && \
wget -q $link/checksums.txt.sig && \
wget -q $link/checksums.txt && \
gpg --import signature.asc && gpg --verify checksums.txt.sig && \
grep "$(sha256sum znapzend-exporter_linux_amd64)" checksums.txt
```

## Helm chart

```bash
helm repo add ccremer https://ccremer.github.io/charts
helm install znapzend ccremer/znapzend --set metrics.image.tag={{ .Tag }}
```
