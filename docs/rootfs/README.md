# RootFS

You can fetch an OS image directly from ubuntu here http://cdimage.ubuntu.com/ubuntu-base/releases/  or dump a docker image for example.

To dump a docker image, proceed as follow:

```shell
$ docker create {your image name}
553a07bcdb4e798f4083211dd3e7a0ec755f8bedcabbbc1e1b78892f3e0d8082

$ docker export 553a07bcdb4e798f4083211dd3e7a0ec755f8bedcabbbc1e1b78892f3e0d8082 > output.tar
```

You may also set it up manually as follow in a chrooted environment:

```shell
$ mkdir -p /home/arthur/toastcloud/build/images/base
$ cd /home/arthur/toastcloud/build/images/base

$ sudo tar -xf ubuntu-base-20.04.5-base-amd64.tar.gz
$ rm ubuntu-base-20.04.5-base-amd64.tar.gz

$ sudo chroot ./ /bin/bash

root@arthur:/# ls
bin  boot  dev  etc  home  lib  lib32  lib64  libx32  media  mnt  opt  proc  root  run  sbin  srv  sys  tmp  usr  var

root@arthur:/#  mknod -m 622 /dev/console c 5 1
root@arthur:/#  mknod -m 666 /dev/null c 1 3
root@arthur:/#  mknod -m 666 /dev/zero c 1 5
root@arthur:/#  mknod -m 666 /dev/ptmx c 5 2
root@arthur:/#  mknod -m 666 /dev/tty c 5 0
root@arthur:/#  mknod -m 444 /dev/random c 1 8
root@arthur:/#  mknod -m 444 /dev/urandom c 1 9

root@arthur:/# groupadd -g 1000 ubuntu --system

root@arthur:/# adduser --system --uid 1000 --gid 1000 --disabled-password --disabled-login ubuntu
Adding system user `ubuntu' (UID 1000) ...
Adding new user `ubuntu' (UID 1000) with group `ubuntu' ...
Creating home directory `/home/ubuntu' ...

root@arthur:/# chown -v ubuntu:tty /dev/console
changed ownership of '/dev/console' from root:root to ubuntu:tty
root@arthur:/# chown -v ubuntu:tty /dev/ptmx
changed ownership of '/dev/ptmx' from root:root to ubuntu:tty
root@arthur:/# chown -v ubuntu:tty /dev/tty
changed ownership of '/dev/tty' from root:root to ubuntu:tty
root@arthur:/# chown -v ubuntu:ubuntu /dev/null
changed ownership of '/dev/null' from root:root to ubuntu:ubuntu
root@arthur:/# chown -v ubuntu:ubuntu /dev/zero
changed ownership of '/dev/zero' from root:root to ubuntu:ubuntu
root@arthur:/# chown -v ubuntu:ubuntu /dev/random
changed ownership of '/dev/random' from root:root to ubuntu:ubuntu
root@arthur:/# chown -v ubuntu:ubuntu /dev/urandom
changed ownership of '/dev/urandom' from root:root to ubuntu:ubuntu

root@arthur:/# echo "nameserver 8.8.8.8" >> /etc/resolv.conf

root@arthur:/# apt-get update

root@arthur:/# apt-get install apt-utils wget iputils-ping git build-essential ca-certificates curl nano -y

root@arthur:/# cd tmp

root@arthur:/tmp# wget https://go.dev/dl/go1.19.1.linux-amd64.tar.gz

root@arthur:/tmp# tar -C /usr/local -xzf go1.19.1.linux-amd64.tar.gz

root@arthur:/tmp# rm go1.19.1.linux-amd64.tar.gz

root@arthur:/tmp# mkdir -p /home/ubuntu/go/src
root@arthur:/tmp# mkdir -p /home/ubuntu/go/pkg
root@arthur:/tmp# mkdir -p /home/ubuntu/go/bin

root@arthur:/# chown ubuntu:ubuntu -hR bin boot etc home lib lib32 lib64 libx32 mnt opt proc root run sbin srv sys tmp usr var media
root@arthur:/# chown ubuntu:ubuntu dev

root@arthur:/# passwd ubuntu
New password: toastcloud

root@arthur:/# usermod -aG sudo ubuntu

root@arthur:/# sudo -u ubuntu /bin/bash

ubuntu@arthur:/$ cd ~
ubuntu@arthur:~$ pwd
/home/ubuntu

ubuntu@arthur:~$ curl -sL https://deb.nodesource.com/setup_16.17 -o /tmp/nodesource_setup.sh
ubuntu@arthur:~$ curl -sL https://deb.nodesource.com/setup_16.x -o /tmp/nodesource_setup.sh

# check the content of the script
ubuntu@arthur:~$ nano /tmp/nodesource_setup.sh

ubuntu@arthur:~$ sudo bash /tmp/nodesource_setup.sh

ubuntu@arthur:~$ sudo apt install nodejs -y

ubuntu@arthur:~$ node -v
v16.17.0

# CTRL + A + D
ubuntu@arthur:~$ exit

# CTRL + A + D
root@arthur:/# exit

$ pwd
/home/arthur/toastcloud/build/images/base

$ sudo tar czf ubuntu-20.04-nodejs-golang.tar.gz ./

$ mv ubuntu-20.04-nodejs-golang.tar.gz ../
$ cd ../
$ sudo rm -rf base/
```