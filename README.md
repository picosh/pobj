# pobj

`rsync`, `scp`, `sftp`, and `sshfs` for your object store. No extra front-end CLI tools
necessary, use what you already have installed.

All you need to get started is our golang binary and an object store.

# supported object stores

Currently support object stores:

- [filesystem](./storage/fs.go)
- [minio](./storage/minio.go)
- [s3](./storage/s3.go)

We provide an [interface](./storage/storage.go) to build your own.

We plan on slowly building more object storage interfaces but this is all we use
at [pico.sh](http://pico.sh).

# demo

```bash
go run ./cmd/authorized_keys
```

Separate terminal:

```bash
rsync -e "ssh -p 2222" -rv ./files localhost:/
scp -P 2222 -r ./files localhost:/
sftp -P 2222 localhost
sshfs -p 2222 localhost:/ ./objs
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
ghcr.io/picosh/pobj/pobj:latest
```

We also have a [docker compose file](./docker-compose.yml) which uses `minio`.

# inspiration

- [rclone](https://github.com/rclone/rclone)
