
[![Release](https://img.shields.io/github/release/nuvo/emq_exporter.svg)](https://github.com/nuvo/emq_exporter/releases)
[![Travis branch](https://img.shields.io/travis/nuvo/emq_exporter/master.svg)](https://travis-ci.org/nuvo/emq_exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/nuvo/emq_exporter.svg)](https://hub.docker.com/r/nuvo/emq_exporter/)
[![Go Report Card](https://goreportcard.com/badge/github.com/nuvo/emq_exporter)](https://goreportcard.com/report/github.com/nuvo/emq_exporter)
[![license](https://img.shields.io/github/license/nuvo/emq_exporter.svg)](https://github.com/nuvo/emq_exporter/blob/master/LICENSE)

# EMQ exporter for Prometheus

A simple server that scrapes EMQ metrics and exports them via HTTP for Prometheus consumption.

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

Specify EMQ's node uri and api port using the `--emq.uri` flag (short `-u`). For example,

```bash
./emq_exporter --emq.uri "http://localhost:8080"
```

Or to scrape a remote host:

```bash
./emq_exporter -u "https://emq.example.com:8080"
```

### Passing Credentials

EMQ requires that calls made to the API endpoints be authenticated. The exporter supports two ways to pass credentials:

1. Setting environment variables:
* `EMQ_USERNAME` for username
* `EMQ_PASSWORD` for the password

No need to pass anything to `emq_exporter` when using these vars, they will be searched for automatically on startup.

2. Using a file

The file should be json formatted and contain the following fields:

```json
{
  "username": "admin",
  "password": "public"
}
```

When staring `emq_exporter`, point it to the credentials file using `--emq.creds-file` flag (short `-f`):

```bash
./emq_exporter -u http://localhost:8080 -f /etc/emq_exporter/auth.json
```

The default path for credentials file is `$(CWD)/auth.json`. Note that `env vars` take precedence over using a file.

### API Version

EMQ add a `v3` api version in `EMQX`. To specify the api version, use the `emq.api-version` flag:

```bash
./emq_exporter -u http://localhost:8080 --emq.api-version v3
```

The `emq_exporter` supports both `v2` and `v3` API versions seamlessly (mutually exclusive, pick either on start up), default is `v2`.

### Authentication

The authentication method changed a bit in version `v3` of `emqx`. If you're pulling the metrics through the dashboard port (default `18083`), you can use regular username and password. However, if you're using the API port (default `8080`), you'll need to set up application credentials: 
1. From the emq dashboard side bar -> applications
2. Select `New App` from the top 
3. Fill in the popup window with the relevant details and confirm
4. View the app details and use `AppID` as `username` and `AppSecret` as `password` (as `creds-file` entries or `env vars`, see above)

The default port `emq_exporter` uses is `18083`

See the docs for `v2` REST API [here](http://emqtt.io/docs/v2/rest.html) and for `v3` [here](http://emqtt.io/docs/v3/rest.html)

### Troubleshooting

If things aren't working as expected, try to start the exporter with `--log.level debug` flag. This will log additional details to the console and might help track down the problem. Fell free to raise an issue should you require additional help.

### Docker

To run EMQ exporter as a Docker container, run:

```bash
docker run \
  -d \
  -p 9540:9540 \
  --name emq_exporter \
  -v /path/to/auth.json:/etc/emq/auth.json \
  nuvo/emq_exporter:v0.4.1 \
  --emq.uri "http://<emq-ip>:8080" \
  --emq.node "emqx@<emq-ip>" \
  --emq.api-version "v3" \
  --emq.creds-file "/etc/emq/auth.json"
```

Alternatively, One can also supply the credentials using `env vars`, replace the volume mount (`-v` flag) with `-e EMQ_USERNAME=<my-username> -e EMQ_PASSWORD=<super-secret>`

### Kubernetes

EMQ exporter was designed to run as a sidecar in the same pod as EMQ itself. 
See the examples folder for a `kubernetes` manifest that can serve as reference for implementation.

## Contributing

We welcome contributions!

Please see [CONTRIBUTING](https://github.com/nuvo/emq_exporter/blob/master/CONTRIBUTING.md) for guidelines on how to get involved.

## License
Apache License 2.0, see [LICENSE](https://github.com/nuvo/emq_exporter/blob/master/LICENSE).