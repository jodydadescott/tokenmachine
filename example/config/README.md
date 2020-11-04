# Configuration

The config is provided by one or more files that may be JSON, YAML, or Rego. The config may be located on the local disk or on a remote http(s)server such as GitHub. The config can be in a single file or multiple files. When the config is located in multiple files the config will be processed in order. It is possible to use both local and remote configs (such as https://...).

Each entity requires a seed. This MUST remain secret as the SharedSecret secret and Keytab principal password are derived from this. The configuration file should be set with restrictive access permissions or the config should be broken into parts with non-sensitive data in one and sensitive data in the other and the file containg sensitive data should have restrictive read access.

Configuration files can be combined with the command

```bash
tokenmachine --config {list of comma seperated config files such as json, yaml or rego} config make 
```

For example the file combined.yaml in this directory is created with the command

```bash
tokenmachine --config opa.rego,main.yaml config make > combined.yaml
```

This assumes that tokenmachine is in your path.