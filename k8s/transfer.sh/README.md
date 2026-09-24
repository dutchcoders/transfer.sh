# transfer.sh Helm Chart

A Helm chart to deploy [transfer.sh](https://github.com/dutchcoders/transfer.sh) on Kubernetes — easy and secure file sharing from the command line.

## Quick Start

```bash
helm install transfer-sh ./transfer.sh
```

Upload a file:

```bash
curl -u user:pass --upload-file ./file.txt https://transfer.example.com/file.txt
```

## Features

- **Multiple storage backends**: local, S3 (including MinIO), Storj, Google Drive
- **Upload protection**: HTTP Basic Auth, htpasswd (multi-user), IP whitelist
- **Network security**: IP whitelist/blacklist, rate limiting
- **Virus scanning**: optional ClamAV integration
- **Ingress**: standard Kubernetes Ingress and Gateway API (HTTPRoute) support
- **Auto-purge**: automatic file deletion after configurable retention period

## Storage Providers

### Local (default)

Files are stored on a PersistentVolumeClaim.

```yaml
transfersh:
  provider: "local"
  local:
    basedir: "/data"

persistence:
  enabled: true
  size: 10Gi
```

### S3 / MinIO

> **Note:** Tested with MinIO. Not tested directly with AWS S3.

Compatible with AWS S3 and any S3-compatible storage (MinIO, Ceph, etc.).

If no static credentials are configured, transfer.sh uses the AWS SDK default
credential chain. This supports EC2 instance profiles, ECS task roles, and EKS
IAM roles for service accounts (IRSA). For IRSA, annotate the chart's service
account with the role ARN and set `serviceAccount.automount: true`.

**Using a Kubernetes Secret (recommended):**

```bash
kubectl create secret generic s3-creds \
  --from-literal=AWS_ACCESS_KEY=<your-access-key> \
  --from-literal=AWS_SECRET_KEY=<your-secret-key> \
  -n <namespace>
```

```yaml
transfersh:
  provider: "s3"
  s3:
    bucket: "transfer"
    region: "eu-west-1"
    existingSecret: "s3-creds"
    # For S3-compatible storage (MinIO, etc.):
    # endpoint: "https://minio.example.com"
    # pathStyle: true

persistence:
  enabled: false
```

**Using plain values:**

The credentials are stored in a Secret managed by the chart (never in the ConfigMap), but they still live in your values file.

```yaml
transfersh:
  provider: "s3"
  s3:
    bucket: "transfer"
    region: "eu-west-1"
    accessKey: "minioadmin"
    secretKey: "minioadmin"
    endpoint: "https://minio.example.com"
    pathStyle: true

persistence:
  enabled: false
```

### Storj

> **Note:** Not tested.

**Using a Kubernetes Secret (recommended):**

```bash
kubectl create secret generic storj-creds \
  --from-literal=STORJ_ACCESS=<your-access-grant> \
  -n <namespace>
```

```yaml
transfersh:
  provider: "storj"
  storj:
    bucket: "transfer"
    existingSecret: "storj-creds"
```

**Using plain values:**

The access grant is stored in a Secret managed by the chart.

```yaml
transfersh:
  provider: "storj"
  storj:
    bucket: "transfer"
    access: "<your-access-grant>"
```

### Google Drive

> **Note:** Not tested.

Google Drive requires an OAuth `client.json` file obtained from the [Google Cloud Console](https://console.cloud.google.com/) (free, no billing required): enable the **Google Drive API**, then create an **OAuth client ID** (Desktop app) and download the JSON file.

**Using a Kubernetes Secret (recommended):**

```bash
kubectl create secret generic gdrive-client-json \
  --from-file=client.json=/path/to/your/client.json \
  -n <namespace>
```

```yaml
transfersh:
  provider: "gdrive"
  gdrive:
    basedir: "transfer.sh"   # folder name created in Google Drive
    existingSecret: "gdrive-client-json"

persistence:
  enabled: true
  size: 1Gi
```

When `existingSecret` is set, the chart automatically mounts the `client.json` file at the path defined by `clientJsonFilepath` (default: `/config/client.json`).

`basedir` is the name of the folder transfer.sh creates in Google Drive, not a local path. Uploaded files go to Google Drive; the only local state is `token.json` and `root_id.conf`, written to `localConfigPath` (default: `/config/gdrive`). The chart mounts the `data` volume there so the directory stays writable despite the read-only root filesystem. Enable `persistence` so this state survives pod restarts.

#### Initial OAuth authentication

Google Drive requires a one-time OAuth consent flow via a browser. This cannot be done inside a Kubernetes pod directly. The recommended approach is to run transfer.sh locally first to complete the authorization:

```bash
docker run -it -p 8080:8080 \
  -v /path/to/client.json:/config/client.json \
  -v /path/to/gdrive-config:/config/gdrive \
  dutchcoders/transfer.sh \
  --provider gdrive \
  --basedir transfer.sh \
  --gdrive-client-json-filepath /config/client.json \
  --gdrive-local-config-path /config/gdrive
```

Follow the URL printed in the logs and paste the authorization code. The generated `token.json` is written to `/path/to/gdrive-config`. Add it to the same Secret as `client.json` and enable `tokenFromSecret`:

```bash
kubectl create secret generic gdrive-client-json \
  --from-file=client.json=/path/to/your/client.json \
  --from-file=token.json=/path/to/gdrive-config/token.json \
  -n <namespace>
```

```yaml
transfersh:
  provider: "gdrive"
  gdrive:
    existingSecret: "gdrive-client-json"
    tokenFromSecret: true

persistence:
  enabled: true
```

`token.json` is mounted read-only. transfer.sh only writes it when missing, and refreshed access tokens are kept in memory. Keep `persistence` enabled: `root_id.conf` stores the ID of the Drive folder, and without it a new folder is created at every restart.

## Security

### Upload Authentication

Protects uploads only — downloads remain accessible via the generated link.

```yaml
transfersh:
  httpAuth:
    enabled: true
    user: "admin"
    pass: "changeme"          # stored in a chart-managed Secret
    # Or use an existing Secret:
    # existingSecret: "transfer-auth"  # must contain HTTP_AUTH_USER + HTTP_AUTH_PASS
    # Or use htpasswd for multi-user:
    # htpasswd: "/etc/htpasswd"
```

Rendering fails if `enabled` is `true` without `htpasswd`, `existingSecret` or both `user` and `pass`: transfer.sh silently disables authentication when credentials are empty.

### Network Access Control

```yaml
transfersh:
  security:
    rateLimit: 60              # requests per minute (0 = unlimited)
    ipWhitelist: "10.0.0.0/8"  # only these IPs can access (empty = all)
    ipBlacklist: "1.2.3.4"     # block specific IPs
```

> **Note:** The IP filter applies to every request, including kubelet probes. The chart uses TCP probes by default for this reason; keep them TCP if you set `ipWhitelist`.

### Kubernetes NetworkPolicy

Restrict ingress traffic to transfer.sh pods at the Kubernetes network level. When enabled, all inbound connections are denied by default except those explicitly allowed.

**Disabled by default** — enabling without configuration allows all ingress (safe, no disruption):

```yaml
networkPolicy:
  enabled: true
```

**Restrict to your ingress controller namespace** (recommended):

```yaml
networkPolicy:
  enabled: true
  ingressNamespace: "traefik"   # or "ingress-nginx", "kube-system", etc.
```

Only pods from the specified namespace can reach transfer.sh on port 8080. All other namespaces are blocked.

**Custom rules** — allow additional traffic (e.g. monitoring scrapes):

```yaml
networkPolicy:
  enabled: true
  ingressNamespace: "traefik"
  additionalIngressRules:
    - from:
        - podSelector:
            matchLabels:
              app: prometheus
      ports:
        - port: 8080
```

> **Note:** Egress (outbound) traffic is not restricted by this NetworkPolicy. transfer.sh can still reach external storage backends (S3, Storj, etc.).

> **Note:** Log collection via Promtail/Fluent Bit is not affected — those tools read logs from the node filesystem, not via network connections to the pod.

### ClamAV Virus Scanning

```yaml
transfersh:
  clamav:
    host: "clamav.clamav.svc.cluster.local:3310" # <service>.<namespace>.svc.cluster.local:PORT
    prescan: false
```

## Ingress

### Standard Kubernetes Ingress

```yaml
ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: transfer.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: transfer-tls
      hosts:
        - transfer.example.com
```

### Gateway API (HTTPRoute)

```yaml
gatewayApi:
  enabled: true
  parentRefs:
    - name: my-gateway
      namespace: default
  hostnames:
    - transfer.example.com
```

## Values Reference

| Key | Default | Description |
|-----|---------|-------------|
| `replicaCount` | `1` | Number of replicas |
| `image.repository` | `dutchcoders/transfer.sh` | Container image repository |
| `image.tag` | `latest-noroot` | Container image tag - chart is optimize for non root env |
| `fullnameOverride` | `""` | Override the resource names (defaults to `<release>-transfer-sh`) |
| `transfersh.provider` | `local` | Storage backend: `local`, `s3`, `storj`, `gdrive` |
| `transfersh.purgeDays` | `7` | Auto-delete files after N days (0 = disabled) |
| `transfersh.purgeInterval` | `1` | Purge check interval in hours |
| `transfersh.maxUploadSize` | `3145728` | Max upload size in KB (0 = unlimited) |
| `transfersh.randomTokenLength` | `6` | Length of the random token in file URLs |
| `transfersh.local.basedir` | `/data` | Base directory for local storage |
| `transfersh.s3.bucket` | `""` | S3 bucket name |
| `transfersh.s3.region` | `eu-west-1` | S3 region |
| `transfersh.s3.endpoint` | `""` | Custom S3 endpoint (MinIO, etc.) |
| `transfersh.s3.pathStyle` | `false` | Force path-style URLs (required for MinIO) |
| `transfersh.s3.noMultipart` | `false` | Disable multipart uploads |
| `transfersh.s3.existingSecret` | `""` | Secret with `AWS_ACCESS_KEY` + `AWS_SECRET_KEY` |
| `transfersh.s3.accessKey` | `""` | S3 access key (stored in a chart-managed Secret, ignored if `existingSecret` is set) |
| `transfersh.s3.secretKey` | `""` | S3 secret key (stored in a chart-managed Secret, ignored if `existingSecret` is set) |
| `transfersh.storj.bucket` | `""` | Storj bucket name |
| `transfersh.storj.existingSecret` | `""` | Secret with `STORJ_ACCESS` key |
| `transfersh.storj.access` | `""` | Storj access grant (stored in a chart-managed Secret, ignored if `existingSecret` is set) |
| `transfersh.gdrive.basedir` | `transfer.sh` | Name of the folder created in Google Drive |
| `transfersh.gdrive.clientJsonFilepath` | `/config/client.json` | Mount path for the OAuth client JSON file |
| `transfersh.gdrive.localConfigPath` | `/config/gdrive` | Writable directory for `token.json` / `root_id.conf` (backed by the `data` volume) |
| `transfersh.gdrive.existingSecret` | `""` | Secret containing the `client.json` key (auto-mounted) |
| `transfersh.gdrive.tokenFromSecret` | `false` | Also mount the `token.json` key of `existingSecret` into `localConfigPath` |
| `transfersh.httpAuth.enabled` | `false` | Enable HTTP Basic Auth on uploads |
| `transfersh.httpAuth.existingSecret` | `""` | Secret with `HTTP_AUTH_USER` + `HTTP_AUTH_PASS` |
| `transfersh.httpAuth.user` | `""` | Basic auth username (stored in a chart-managed Secret) |
| `transfersh.httpAuth.pass` | `""` | Basic auth password (stored in a chart-managed Secret) |
| `transfersh.httpAuth.htpasswd` | `""` | htpasswd file path (takes precedence over user/pass) |
| `transfersh.httpAuth.ipWhitelist` | `""` | IPs allowed to upload without auth |
| `transfersh.security.rateLimit` | `30` | Rate limit in requests/min (0 = unlimited) |
| `transfersh.security.ipWhitelist` | `""` | Allowed IPs (comma-separated) |
| `transfersh.security.ipBlacklist` | `""` | Blocked IPs (comma-separated) |
| `transfersh.clamav.host` | `""` | ClamAV daemon address |
| `transfersh.clamav.prescan` | `false` | Enable ClamAV prescan |
| `transfersh.extraEnv` | `{}` | Extra environment variables |
| `persistence.enabled` | `false` | Enable PVC (local: uploaded files, gdrive: OAuth token) |
| `persistence.size` | `10Gi` | PVC size |
| `persistence.accessMode` | `ReadWriteOnce` | PVC access mode |
| `ingress.enabled` | `false` | Enable Kubernetes Ingress |
| `gatewayApi.enabled` | `false` | Enable Gateway API HTTPRoute |
| `networkPolicy.enabled` | `false` | Enable Kubernetes NetworkPolicy to restrict ingress traffic |
| `networkPolicy.ingressNamespace` | `""` | Namespace of the ingress controller allowed to reach the app (empty = allow all) |
| `networkPolicy.additionalIngressRules` | `[]` | Additional raw NetworkPolicy ingress rules |
