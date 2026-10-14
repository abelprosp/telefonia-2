#!/bin/sh
set -eu

: "${MINIO_ROOT_USER:?}"
: "${MINIO_ROOT_PASSWORD:?}"
: "${OBJECT_STORAGE_ACCESS_KEY_ID:?}"
: "${OBJECT_STORAGE_SECRET_ACCESS_KEY:?}"
: "${PUBLIC_ORIGIN:?}"
bucket="${OBJECT_STORAGE_BUCKET_NAME:-luxus-connect}"
case "$bucket" in ''|*[!a-z0-9.-]*) echo 'Invalid storage bucket' >&2; exit 1;; esac
case "$PUBLIC_ORIGIN" in https://*) ;; *) echo 'HTTPS public origin required' >&2; exit 1;; esac
origin_host="${PUBLIC_ORIGIN#https://}"
case "$origin_host" in ''|*[!a-zA-Z0-9.:-]*) echo 'Invalid public origin' >&2; exit 1;; esac
if [ "$OBJECT_STORAGE_ACCESS_KEY_ID" = "$MINIO_ROOT_USER" ] || [ "$OBJECT_STORAGE_SECRET_ACCESS_KEY" = "$MINIO_ROOT_PASSWORD" ]; then
  echo 'Application credentials must differ from MinIO root credentials' >&2
  exit 1
fi
mc alias set local http://minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null
mc mb "local/$bucket" --ignore-existing
cat > /tmp/application-policy.json <<EOF
{"Version":"2012-10-17","Statement":[
 {"Effect":"Allow","Action":["s3:GetBucketLocation","s3:ListBucket"],"Resource":["arn:aws:s3:::$bucket"]},
 {"Effect":"Allow","Action":["s3:GetObject","s3:PutObject"],"Resource":["arn:aws:s3:::$bucket/*"]}
]}
EOF
if mc admin policy info local connect-storage >/dev/null 2>&1; then
  mc admin policy update local connect-storage /tmp/application-policy.json 2>/dev/null \
    || echo 'Policy connect-storage already exists'
else
  mc admin policy create local connect-storage /tmp/application-policy.json
fi
mc admin user add local "$OBJECT_STORAGE_ACCESS_KEY_ID" "$OBJECT_STORAGE_SECRET_ACCESS_KEY" >/dev/null
mc admin policy attach local connect-storage --user "$OBJECT_STORAGE_ACCESS_KEY_ID"
# CORS is configured on the MinIO server using MINIO_API_CORS_ALLOW_ORIGIN.
# The community server does not implement the per-bucket SetBucketCors API.
