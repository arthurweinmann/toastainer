# Tests

## Requirements

## Make toaster OS Image

### Base RootFS

You can fetch it directly from ubuntu here http://cdimage.ubuntu.com/ubuntu-base/releases/ or dump a docker image.

To dump a docker image, proceed as follow:

```
$ docker create {your image name}
553a07bcdb4e798f4083211dd3e7a0ec755f8bedcabbbc1e1b78892f3e0d8082

$ docker export 553a07bcdb4e798f4083211dd3e7a0ec755f8bedcabbbc1e1b78892f3e0d8082 > output.tar
```

### Setup with nsjail