#!/bin/bash

set -x

pushd build/docker
set -a
source ../.local.env
set +a

# the following is a very dirty and risky manipulation of the docker compose file, to make sure that
# the catalogue waits for its db before initializing. Only works with the sockshop example...
sed -i '
/^  all_services_ctr:$/ {
  N
  /^  all_services_ctr:\n    build:$/ {
    c\
  all_services_ctr:\
    depends_on:\
      catalogue_db_ctr:\
        condition: service_healthy\
    build:
  }
}
' docker-compose.yml
docker-compose up --build
popd
