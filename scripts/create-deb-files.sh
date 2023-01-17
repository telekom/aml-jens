#! /bin/bash

set -e

if [ -z $1 ]
then
  echo "please call this script this way: "
  echo "$0 <version-number>"
  exit 1
fi

#replace all slashes in version string by underscore
version=$(echo $1 | sed -e 's%/%.%g')
package_name="jens-cli"
package_name_v="${package_name}-${version}"
# copy template for new version
BINS="./bin"
BASE="./build/deb"
OUT= "./out"
package_base_dir="${BASE}/${package_name_v}"
cp -r "${BASE}/jens-cli-template" ${package_base_dir}

# replace version in control file
sed --in-place -e "s%TEMPLATE_VERSION%${version}%g" ${package_base_dir}/DEBIAN/control

# copy jens-cli util to /usr/bin/
mkdir -p ${package_base_dir}/usr/bin
cp "${BASE}/../drplay" ${package_base_dir}/usr/bin/drplay
cp "${BASE}/../drshow" ${package_base_dir}/usr/bin/drshow
cp "${BASE}/../drbenchmark" ${package_base_dir}/usr/bin/drbenchmark

# Gzip man page
gzip ${package_base_dir}/usr/share/man/man1/*
dpkg_cmd="dpkg-deb --build ${package_base_dir}"

if [ -x "$(command -v dpkg-deb )" ];
then
  echo "building package natively...."
  ${dpkg_cmd}
else
  echo "using docker to build package...."
  docker run -it --rm -v $(pwd):/root --workdir /root debian /bin/bash -c "${dpkg_cmd}"
fi

rm -rf $package_base_dir