# foundershk-datasource-agent

The foundershk Private Datasource Connect Agent allows connecting private datasources with your grafana cloud instance.

## Installation

Follow installation and running instructions in the [Grafana Labs Documentation](https://grafana.com/docs/grafana-cloud/data-configuration/configure-private-datasource-connect/)

## Setting the ssh log level

Use the `-log.level` flag. Run the agent with the `-help` flag to see the possible values.

| go log level | ssh log level    |
| ------------ | ---------------- |
| `error`      | 0 (`-v` not set) |
| `warn`       | 0 (`-v` not set) |