#!/bin/bash
#
#aml-jens
#
#(C) 2023 Deutsche Telekom AG
#
#Deutsche Telekom AG and all other contributors /
#copyright owners license this file to you under the Apache
#License, Version 2.0 (the "License"); you may not use this
#file except in compliance with the License.
#You may obtain a copy of the License at
#
#http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing,
#software distributed under the License is distributed on an
#"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
#KIND, either express or implied.  See the License for the
#specific language governing permissions and limitations
#under the License.
#
ssh_target="jens.tether"
ssh="ssh $ssh_target "
scp="scp $ssh_target:"
docker="${ssh} docker exec "
container_id=""
database="l4s_measure "
getid()
{
    container_id=`${ssh} docker ps --filter status=running | grep edgeaml | awk '{print $1}'`
    if [ -z $container_id ]; then
        echo "Could not get container_id"
        exit 255
    fi
    docker="$docker $container_id"
}

create_sql_dump()
{
    if [ -z ${1+x} ]; then 
        FILE_NAME="dump_"$(date +"%Y%m%d_%H%M%S");
    else 
        FILE_NAME=$1; 
    fi
    OUTPUT_FILE="${FILE_NAME}.d.tar.gz"
    $ssh docker exec -u postgres $container_id pg_dump $database -j 8 --format d -f /tmp/${FILE_NAME}/
    echo "Compressing dump (may take a few minutes)"
    $docker tar -zcf $OUTPUT_FILE -C /tmp/ ${FILE_NAME} > /dev/null
    $ssh docker cp $container_id:/usr/share/grafana/$OUTPUT_FILE ./
    if [ ! -z $container_id ]; then
    `${scp}$OUTPUT_FILE ./`
    fi
    echo "Dump located @ ./$OUTPUT_FILE"
}

create_dump()
{
    getid
    echo "Working in Container: $container_id"
    echo "Creating dump file (may take several minutes)"
    create_sql_dump
}
create_dump
set -x
origin_path="/home/merren/src/aml/aml-jens/dump_20230420_145320.d.tar.gz"
folder_name="20230420_145320"
file_name="${folder_name}.tar.gz"
path_tar="/tmp/${file_name}"
#if [ ! -z "$ssh" ]; then
#    scp ${origin_path} ${ssh_target}:$path_tar;
#     else
#    cp $origin_path $path_tar
#    fi
#
$ssh docker cp $path_tar $container_id:$path_tar
$docker tar -zxvf $path_tar -C "/tmp/"
$ssh docker exec -u pg_restore -v -cC -j 4 /tmp/dump_20230420_145320/ -d l4s_measure  2>&1 >/dev/null | grep -E 'error|finished'
set +x