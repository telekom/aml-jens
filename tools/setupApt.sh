#! /bin/sh

echo 'deb [trusted=yes] https://jens.llcto.telekom-dienste.de/ ./' | tee -a /etc/apt/sources.list
echo '
Package: iproute2
Pin: origin jens.llcto.telekom-dienste.de
Pin-Priority: 1001
' > /etc/apt/preferences

echo '
machine https://jens.llcto.telekom-dienste.de/
login jens_fileserver
password 3+bEacQgweal0ruf7A6gt2FkoDK0mcNz9y03Lbl3Qkc=
' > /etc/apt/auth.conf

apt update
