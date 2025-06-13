# Semester Project

## Plugin implementation

The plugin implementation can be found in `plugins/namedpipe`. 

The type definitions and utilities used at runtime by the named pipe plugin can be found in `runtime/plugins/namedpipe`.

## E2e experiments with SockShop

### Wiring specs

The wiring spec for the named pipes deployment can be found in `examples/sockshop/wiring/specs/namedpipes.go` (what I refer to as NamedPipes deployment in the report). The `docker.go` wiring (**what I refer to gRPC deployment in the report**) has also been changed.

### Workload generator

The workload generator used for the e2e experiments can be found in `examples/sockshop/cmplx_workload`.

### Helper scripts to run the experiments and process the data

`run_experiments.sh` does the following:
- compiles the application and the workload generator (defined by `examples/sockshop/cmplx_workload/workloadgen/workload.go`),
- starts the application using the `examples/sockshop/docker.sh` helper script,
- waits until the services are running (by monitoring the logs),
- starts the workload generator using the `examples/sockshop/wlgen.sh` helper script and
- stops the application once the workload generator terminates.

`run_experiments_on_cluster.sh` can be used to start multiple experiments on a remote machine.

The data from the experiments can be found in folders:
- `experiments-final`: the data used for the plots in the report
- `experiments-message-sizes`: the data used for generating the files in folder `message-sizes`, which are then used by `calculate_msg_sizes.py` to calculate the average request and response sizes mentioned in the report

`process_stats.py` and `calculate_msg_sizes.py` can be used to process the logs and data collected during the experiments.

### Blueprint patches to make the experiments work

Compilation issues:
- `plugins/golang/gogen/workspacebuilder.go`: use newer version of Go in the workspaces of the generated boilerplates
- `plugins/goproc/linuxgen/dockerfile_buildcommands.go`: use Docker image that supports the newer workspace version

Runtime issues:
- `plugins/http/httpcodegen/clientgen.go`: hardcode 127.0.0.1 as the server IP in the HTTP client boilerplate. This is necessary because Blueprint does not seem to allow directly deploying wlgen in a container. Therefore its deployment is separate from the Docker compose deployment of the other services and cannot benefit from Docker compose's name resolution. Without hardcoding this, wlgen would try to send requests to `http://frontend:PORT` but would be unable to resolve frontend.
