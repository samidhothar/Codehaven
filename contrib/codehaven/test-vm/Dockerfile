# A stand-in for a real virtual machine: SSH for deployments, and HTTP to look at what was deployed
FROM alpine:3.22

RUN apk add --no-cache openssh-server tar busybox-extras \
 && adduser -D -s /bin/sh deploy \
 && echo 'deploy:*' | chpasswd -e \
 && mkdir -p /srv/app \
 && chown deploy /srv/app \
 && sed -i -e 's/^#\?PasswordAuthentication.*/PasswordAuthentication no/' -e 's/^#\?PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config

COPY entrypoint.sh /entrypoint.sh
EXPOSE 22 8081
ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]
