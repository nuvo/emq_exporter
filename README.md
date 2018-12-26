
[![Release](https://img.shields.io/github/release/nuvo/emq_exporter.svg)](https://github.com/nuvo/emq_exporter/releases)
[![Travis branch](https://img.shields.io/travis/nuvo/emq_exporter/master.svg)](https://travis-ci.org/nuvo/emq_exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/nuvo/emq_exporter.svg)](https://hub.docker.com/r/nuvo/emq_exporter/)
[![Go Report Card](https://goreportcard.com/badge/github.com/nuvo/emq_exporter)](https://goreportcard.com/report/github.com/nuvo/emq_exporter)
[![license](https://img.shields.io/github/license/nuvo/emq_exporter.svg)](https://github.com/nuvo/emq_exporter/blob/master/LICENSE)

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

### Docker

To run EMQ exporter as a Docker container, run:

```bash
docker run -p 9505:9505 nuvo/emq_exporter:v0.1.0 ---emq.uri="http://localhost:8080"
```

### Kubernetes

EMQ exporter was designed to run as a sidecar in the same pod as EMQ itself. See the examples folder for a `kubernetes` manifest that can serve as reference for implementation.

## Contributing

We welcome contributions!

Please see [CONTRIBUTING](https://github.com/nuvo/emq_exporter/blob/master/CONTRIBUTING.md) for guidelines on how to get involved.

## License
Apache License 2.0, see [LICENSE](https://github.com/nuvo/emq_exporter/blob/master/LICENSE).