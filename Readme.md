# Kube Container Exec

## Summary

This is a small command line tool to basically make `kubectl get pods` and `kubectl exec` combined and usable in kubernetes.

One of the main purposes is using Kubernetes Jobs and CronJobs and use this tool to execute jobs in existings (Pet-) Pods.

Sometimes it is not feasible to run the Job as a new Pod, so this tool uses a very small Pod that talks to the Kubernetes API to run the actual Job as a new process in the specified container.

## Usage

The tool is configurable via both environment-variables and command line flags.

    ./exec-linux [flags] command -flag1 -flagFoo=bar

The first `flags` are described below. Everything which comes later is passed to the command(s).

The command is not run in a shell, as there are containers which do not have a shell.
If you want a shell you have to invoke it yourself, e.g.

    ./exec-linux /bin/bash -c 'ps aux && free -m && w'

### Environment Variables

- `KUBECONFIG`: specify the path to a Kubernetes config file. By default there should be sufficient self-configuration possibilities, including in-cluster self-configuration.
- `CONTAINER`: specify the container in which the process should be started.
- `FILTER`: specify which filter should be used to find a pod, e.g. `app=mytool`.

### Command Line Flags (excerpt)

- `-v`: Verbosity level, `4` includes exec-specific output, `6` includes basic kubernetes debug, `8` includes REST-api debugging.
- `-container`: Set the container, overrides the `CONTAINER` environment variable
- `-filter`: Set the filter, overrides the `FILTER` environment variable
- `-kubeconfig`: Set the path to kubeconfig. Overrides the `KUBECONFIG` variable.
