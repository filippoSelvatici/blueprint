#!/bin/bash

set -x

wiring="$1"

if [ -z "$wiring" ]; then
  echo "Usage: $0 <wiring>"
  exit 1
fi

if [ "$wiring" != "docker" ] && [ "$wiring" != "namedpipes" ]; then
  echo "Invalid wiring type. Use 'docker' or 'namedpipes'."
  exit 1
fi

# TODO update this path
pushd /home/fselvatici/blueprint/examples/sockshop

# to avoid grepping the previous content of the file
echo "" > logs.txt
# to ensure that the correct docker socket is used
set -a
# TODO update the user ID for the userspace docker socket
DOCKER_HOST="unix:///run/user/614046/docker.sock"
set +a
# run experiment
rm -rf build/ && go run wiring/main.go -o build -w "$wiring" && ./docker.sh > logs.txt 2>&1 &
PID_RUNNER_SUBPROCESS=$!

# wait until we see all processes running
while ! grep -q "catalogue_proc running" logs.txt; do
  echo "Waiting for catalogue to run..."
  sleep 1
done
while ! grep -q "order_proc running" logs.txt; do
  echo "Waiting for order to run..."
  sleep 1
done
while ! grep -q "shipping_proc running shipping_service" logs.txt; do
  echo "Waiting for shipping service to run..."
  sleep 1
done
while ! grep -q "shipping_proc running queue_master" logs.txt; do
  echo "Waiting for queue_master to run..."
  sleep 1
done
while ! grep -q "frontend_proc running" logs.txt; do
  echo "Waiting for frontend to run..."
  sleep 1
done
while ! grep -q "payment_proc running" logs.txt; do
  echo "Waiting for payment to run..."
  sleep 1
done
while ! grep -q "user_proc running" logs.txt; do
  echo "Waiting for user to run..."
  sleep 1
done
while ! grep -q "cart_proc running" logs.txt; do
  echo "Waiting for cart to run..."
  sleep 1
done
sleep 10 # wait a bit more

./wlgen.sh > wllogs.txt 2>&1

if [ "$wiring" == "namedpipes" ]; then
  docker logs docker-all_services_ctr-1 >> logs.txt
else 
  for id in $(docker ps --format "{{.ID}} {{.Names}}" | awk '{print $1}'); do
    echo "--- Logs for container: $id ---"
    docker logs "$id" >> logs.txt
  done
fi

# Cleanup
pushd build/docker
set -a
source ../.local.env
set +a
docker-compose down
popd

popd