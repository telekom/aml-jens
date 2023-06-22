#! /bin/bash
#
# aml-jens
#
# (C) 2023 Deutsche Telekom AG
#
# Deutsche Telekom AG and all other contributors /
# copyright owners license this file to you under the Apache
# License, Version 2.0 (the "License"); you may not use this
# file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#
ARGS=("$@") 
logfile="example_callback_log"

pre_benchmark()
{
  BENCHMARK_ID=$1
  BENCHMARK_NAME=$2
  BENCHMARK_TAG=$3
  echo "=======PRE_BENCHMARK=======">> ${logfile}
  echo "BENCHMARK_ID=   "$BENCHMARK_ID >> ${logfile}
  echo "BENCHMARK_NAME= "$BENCHMARK_NAME >> ${logfile}
  echo "BENCHMARK_TAG=  "$BENCHMARK_TAG >> ${logfile}
}
post_benchmark()
{
  BENCHMARK_ID=$1
  echo "=======POST_BENCHMARK=======">> ${logfile}
  echo "BENCHMARK_ID=   "$BENCHMARK_ID >> ${logfile}
}

pre_session()
{
  SESS_NBR=$1 # 0,1,2...
  echo "=======PRE_SESS=======">> ${logfile}
  echo "SESS_NBR=    "$SESS_NBR >> ${logfile}
}


post_session()
{
  SESS_SUCCESS=$1
  SESS_ID=$2
  SESS_START_T=$3
  SESS_END_T=$4
  echo "=======POST_SESS=======">> ${logfile}
  echo "SESS_ID=      "$SESS_ID >> ${logfile}
  echo "SESS_SUCCESS= "$SESS_SUCCESS >> ${logfile}
  echo "SESS_START_T= "$(date -Rd @$SESS_START_T) >> ${logfile}
  echo "SESS_END_T=   "$(date -Rd @$SESS_END_T) >> ${logfile}
}


case ${ARGS[0]} in

  '1')
    pre_benchmark ${ARGS[@]:1}
    ;;

  "4")
    pre_session ${ARGS[@]:1}
    ;;

  "7")
    post_session ${ARGS[@]:1}
    ;;

  "9")
    post_benchmark ${ARGS[@]:1}
    ;;

  *)
    echo "UNKNOWN">> ${logfile}
    ;;
esac

sleep 0.1