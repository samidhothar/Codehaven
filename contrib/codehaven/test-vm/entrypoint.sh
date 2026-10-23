#!/bin/sh
set -e

# the shared key volume is written by the keygen service before this container starts
install -d -m 700 -o deploy -g deploy /home/deploy/.ssh
install -m 600 -o deploy -g deploy /run/codehaven-keys/id_ed25519.pub /home/deploy/.ssh/authorized_keys
ssh-keygen -A >/dev/null

httpd -p 8081 -h /srv/app
exec /usr/sbin/sshd -D -e
