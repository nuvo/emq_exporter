# EMQ exporter for Prometheus

This is a simple server that scrapes EMQ metrics and exporters them via HTTP for
Prometheus consumption.

## Getting Started

To run it:

```bash
./emq_exporter [flags]
```

Help on flags:

```bash
./emq_exporter --help
```

## Usage

### EMQ URI

Specify EMQ's node uri and api port using the `--emq.uri` flag. For example,

```bash
./emq_exporter --emq.uri="http://localhost:8080"
```

Or to scrape a remote host:

```bash
./emq_exporter --emq.uri="https://emq.example.com:8080"
```

You will also need to specify the user credentials (user name and password):

```bash
./emq_exporter --emq.uri="http://localhost:8080" --emq.username="admin" --emq.password="public"
```

The username and password can alternatively be specified in environment variables - `EMQ_USERNAME` and `EMQ_PASSWORD` respectively

### API Version

EMQ add a `v3` api version in `EMQX`. To specify the api version, use the `emq.api-version` flag:

```bash
./emq_exporter --emq.uri="http://localhost:8080" --emq.username="admin" --emq.password="public" --emq.api-version="v3"
```

Note - currently only `v2` api version is supported (and set as the flag's default value). Passing any other values will cause the exporter to panic. We plan to add support for `v3` api in future releases of the exporter.

### Docker

To run EMQ exporter as a Docker container, run:

```bash
docker run -p 9505:9505 nuvo/emq_exporter:v0.1.0 ---emq.uri="http://localhost:8080"
```

### Kubernetes

EMQ exporter was designed to run as a sidecar in the same pod as EMQ itself. See the examples folder for a `kuberentes` manifest that can serve as reference for implementation.

## Contributing

We welcome contributions!

Please see [CONTRIBUTING](./CONTRIBUTING.md) for guidelines on how to get involved.

## License
Apache License 2.0, see [LICENSE](./LICENSE).