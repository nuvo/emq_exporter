
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

You will also need to specify the authentication credentials (see below for additional details):

```bash
./emq_exporter --emq.uri="http://localhost:8080" --emq.username="admin" --emq.password="public"
```

The credentials can alternatively be specified in environment variables - `EMQ_USERNAME` and `EMQ_PASSWORD` respectively

### API Version

EMQ add a `v3` api version in `EMQX`. To specify the api version, use the `emq.api-version` flag:

```bash
./emq_exporter --emq.uri="http://localhost:8080" --emq.username="admin" --emq.password="public" --emq.api-version="v3"
```

The `emq_exporter` supports both `v2` and `v3` API versions seamlessly (mutually exclusive, pick either on start up)

### Authentication

The authentication method changed a bit in version `v3` of emq. If you're pulling the metrics through the dashboard port (default `18083`), you can use regular username and password. However, if you're using the API port (default `8080`), you'll need to set up application credentials: 
1. From the emq dashboard side bar -> applications
2. Select `New App` from the top 
3. Fill in the popup window with the relevant details and confirm
4. View the app details and use `AppID` as `--emq.username` and `AppSecret` as `--emq.password`

The default port `emq_exporter` uses is `18083`

See the docs for `v2` REST API [here](http://emqtt.io/docs/v2/rest.html) and for `v3` [here](http://emqtt.io/docs/v3/rest.html)

### Troubleshooting

If things aren't working as expected, try to start the exporter with `--log.level="debug"` flag. This will log additional details to the console and might help track down the problem. Fell free to raise an issue should you require additional help

### Docker

To run EMQ exporter as a Docker container, run:

```bash
docker run -p 9505:9505 nuvo/emq_exporter:v0.3.1 ---emq.uri="http://localhost:8080"
```

### Kubernetes

EMQ exporter was designed to run as a sidecar in the same pod as EMQ itself. See the examples folder for a `kubernetes` manifest that can serve as reference for implementation.

## Contributing

We welcome contributions!

Please see [CONTRIBUTING](https://github.com/nuvo/emq_exporter/blob/master/CONTRIBUTING.md) for guidelines on how to get involved.

## License
Apache License 2.0, see [LICENSE](https://github.com/nuvo/emq_exporter/blob/master/LICENSE).