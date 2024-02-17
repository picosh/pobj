# objx

`rsync`, `scp`, `sftp` for your object store. No extra front-end CLI tools
necessary, use what you already have installed.

# supported object stores

Currently support object stores:

- [filesystem](https://github.com/picosh/objx/blob/9e920bd907fca88ad90a300b02254464e3f598fb/storage/fs.go#L1)
- [minio](https://github.com/picosh/objx/blob/9e920bd907fca88ad90a300b02254464e3f598fb/storage/minio.go#L1)

We provide an
[interface](https://github.com/picosh/objx/blob/9e920bd907fca88ad90a300b02254464e3f598fb/storage/storage.go#L1)
to build your own.

We plan on slowly building more object storage interfaces but this is all we use
at [pico.sh](http://pico.sh).

# demo

```bash
make build
./build/authorized_keys
```

Separate terminal:

```bash
rsync -e "ssh -p 2222" -r ./files localhost:/
scp -P 2222 -r ./files localhost:/
sftp -P 2222 localhost
```

# info

By default, the user sent to the SSH server will be the bucket name and will be
created on-the-fly if it doesn't already exist.

You are free to change the bucket by providing whatever you want as the user:

```bash
scp -P 2222 -r ./files mybucket@localhost:/
```

# docker

```
ghcr.io/picosh/objx/objx:latest
```

We also have a
[docker compose file](https://github.com/picosh/objx/blob/305f252057d73d69a15b03cdf364596040f3735d/docker-compose.yml#L1)
which uses `minio`
