# Installation

## Windows

For Windows, it is intended to run the server as a Windows service. The binary can install itself as a service and even start itself. Installation consists of placing the binary in a desired location, setting the config(s), installing the service, and finally starting the service. The service must be configured to run as a Directory Admin (this is required for the creation of keytabs). The config (covered in the configuration section) is one or more YAML, JSON, or OPA/Rego files or URLs.

Create a directory for the binary

```bash
mkdir "C:\Program Files\TokenMachine"
```

Change directory to "C:\Program Files\TokenMachine" and download the binary into this directory. Then install the windows service and create an example config with the following commands.

```bash
.\tokenmachine.exe service install
.\tokenmachine.exe config example > config.yaml
```

You will need to edit the file config.yaml. See Here.

Now set the config file location with the command

```bash
.\tokenmachine.exe service config set "C:\Program Files\TokenMachine\config.yaml"
```

Note that you may specify more than one file by separating the entries with a comma. For example

```bash
.\tokenmachine.exe service config set "C:\Program Files\TokenMachine\config.yaml,C:\other.yaml,https://github.com/myrepo/config.yaml"
```

It is a good idea to keep the file or files containing sensitive data such as seeds in a separate file and give it restrictive permissions.

When you are ready you can start the service with the command

```bash
.\tokenmachine.exe service start
```

## Linux and Darwin

For Linux and Darwin download and install the binary tokenmachine. Note that valid Keytabs will NOT be issued on Linux and Darwin.

Create an example config with the command

```bash
./tokenmachine config example > example.yaml
```

Run the server with the command

```bash
./tokenmachine --config example.yaml
```
