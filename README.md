# transfer.sh on Railway, Cloudflare R2, and Cloudflare Workers

transfer.sh is a nice front-end for quickly sharing files, and can be easily self-hosted on https://railway.app.
However, having Railway proxy file downloads can be somewhat slow. So I used a Worker to intercept downloads and serve them straight from R2.

1. Deploy this repo to Railway with the following variables:

```
S3_NO_MULTIPART=true
LISTENER=8080
PURGE_DAYS=1
PROVIDER=s3
AWS_SECRET_KEY=xyz
RANDOM_TOKEN_LENGTH=12
S3_ENDPOINT=xyz.r2.cloudflarestorage.com
PORT=8080
AWS_ACCESS_KEY=xyz
PURGE_INTERVAL=1
BUCKET=my-bucket
S3_REGION=auto
```

2. Add a custom domain to the railway project that is orange-clouded in Cloudflare

3. In cf-worker, update wrangler.toml with the custom domain you used and your r2 bucket

4. Deploy cf-worker

```shell
cd cf-worker
npx wrangler publish
```
