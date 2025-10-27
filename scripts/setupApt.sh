#! /bin/sh

echo 'deb [trusted=yes] https://<<>>/ ./' | tee -a /etc/apt/sources.list
echo '
Package: iproute2
Pin: origin <<>>
Pin-Priority: 1001
' > /etc/apt/preferences

echo '
machine https://<<>>/
login <<>>
password <<>>
' > /etc/apt/auth.conf

apt update