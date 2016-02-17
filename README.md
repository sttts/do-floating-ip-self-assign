# DigitalOcean Floating-IP Self-Assign

This tool assigns a given DigitalOcean floating IP to the current droplet. It get's the droplet id with the DigitalOcean metadata service.

## Usage

```bash
$ docker run -it sttts/do-floating-ip-self-assign --v=2 \
    -token 12936827563745928758923423424 -floating-ip 1.2.3.4 \
    --logtostderr=true
```

The token is a DigitalOcean API token which you can create at https://cloud.digitalocean.com/settings/api/tokens.

Usually, you will want to run `sttts/do-floating-ip-self-assign` as a side-kick container e.g. of a Kubernetes pods which offers a certain service on the floating ip.

It is also quite useful when setting up an highly available Kubernetes master using the podmaster utility (compare http://kubernetes.io/v1.1/docs/admin/high-availability.html). Just let the podmaster also control an instance of the `sttts/do-floating-ip-self-assign` container following that of the other master components. Here is the pod:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: do-floating-ip-self-assign
spec:
  containers:
  - name: do-floatingip-self-assign
    image: sttts/do-floating-ip-self-assign
    command: [
      "/do-floating-ip-self-assign",
      "-token={{.ApiToken}}",
      "-floating-ip={{.FloatingIP}}",
      "-v=4", "-logtostderr=true"
    ]
```

Here is the command line help:

```
Usage of /do-floating-ip-self-assign:
  -alsologtostderr
        log to standard error as well as files
  -backoff duration
        Initial backoff time after a failure. (default 1s)
  -backoff-factor float
        Backoff time multiplier after each failure. (default 1.2)
  -backoff-max duration
        Maximum backoff time after a failure. (default 30s)
  -floating-ip string
        The floating IP address to self-assign.
  -log_backtrace_at value
        when logging hits line file:N, emit a stack trace (default :0)
  -log_dir string
        If non-empty, write log files in this directory
  -logtostderr
        log to standard error instead of files
  -retries int
        The number of retries when self-assignment fails, negative values for forever. (default 5)
  -stderrthreshold value
        logs at or above this threshold go to stderr
  -token string
        A DigitalOcean API token.
  -token-file string
        A file path with a DigitalOcean API token.
  -update-period duration
        The time between floating IP update tries, 0 for only initial assignment. (default 1m0s)
  -v value
        log level for V logs
  -vmodule value
        comma-separated list of pattern=N settings for file-filtered logging
```

## Build

```bash
$ go get github.com/kubermatic/do-floating-ip-self-assign
$ cd src/github.com/kubermatic/do-floating-ip-self-assign
$ make docker
```
