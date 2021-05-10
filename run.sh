#!/usr/bin/dumb-init /bin/bash

DEFAULT_HFM_LOGLEVEL="info"
DEFAULT_HFM_PORT="8844"
DEFAULT_HFM_CONFIG="./config.yaml"

. /env.sh

for var in ${!DEFAULT_HFM*}; do
  t=${var/DEFAULT_/}
  if [ -z ${!t} ]; then
    echo "Using default for ${t}:${!var}"
    eval ${t}=${!var}
    export "${t}"
  else
    echo "Using override value for ${t}"
  fi
done


exec "${@}"

