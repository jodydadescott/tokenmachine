# TokenMachine

## Overview

### What

Grants SharedSecrets, Kerberos Keytabs, and Nonces to bearers of authorized OAUTH tokens.

### Why

OAUTH toknes have become very popular in modern application development to authenticate and authorize API calls but many applications use Kerberos or shared secrets.

### How

Users must be able to obtain OAUTH tokens from their identity provider (IDP). Using the token they can request SharedSecrets or Keytabs from the TokenMachine. TokenMachine will authenticate and authorize the request by executing an OPA/Rego policy defined by the operator. Tokens will be validated by obtaining the public key from the issuer. A Nonce mechnacism has also been created to prevent replay attacks.

### More

When an application or users (client) desires a resource from the Tokenmachine they must first acquire an OAUTH compliant token (bearer token) from their Identity Provider (IDP). Using this bearer token the client should request a Nonce from the Tokenmachine. The Tokenmachine will authorize the Nonce request and if authroized return a time scoped Nonce. The Nonce will contain both a secret and an expiration time. The client should obtain a new token from their IDP with the Nonce secret encoded in the claims. The client should then use the bearer token to make one or more request to the Tokenmachine before the Nonce expires.

SharedSecrets contain a Secret and an expiration (exp) time in epoch seconds. If the SharedSecret has reached half life then the fields NextSecret and NextExp will be present. These fields represent the next period Secret and expiration. It is up to the user to make use of the SharedSecret and cordinate the period changes.

Keytabs contain a base64 Keytab file in the field Base64File and a Kerberos principal in the field Principal. It is not currently possible to provide the next Keytab as it is for a SharedSecret. Generally the user should obtain a Keytab when required and then discard it.

Authentication and authorization is performed by validating a bearers token signature and executing a pre-configured OPA/Rego policy. The signature validation works by following the tokens issuer and obtaining the public key. The issuer name must match the TLS certificate in the HTTPS request. The OPA/Rego policy may also be used to prevent replay attack by checking to see if a nonce pattern is inside the payload of the bearer token.

The Tokenmachine server works by being preconfigured with SharedSecrets and Keytabs. We will referfer to these as entities for short. Both entities must be configured with a seed. This seed should be kept secret as it will be used to generate the secret for the SharedSecret and the password for the Keytab principal. Each entity must also be configured with a Lifetime in seconds. Each entity will have a time period calculated by taking the seconds since epoch and dividing it by the Lifetime and removing the remainder. This will provide us with the current period and by adding the lifetime to this we can determine the next period.

When a request for a SharedSecret is made the current time period is determined and the seed is used to obtain a one time code. This is then appended to the seed and a SHA-256 hash is calculated. This will be converted into a alpha-numeric string.

Keytabs operate in a similiar method but there are signifigant differences. At the top of each period Keytabs are created using a password that is derived in a similiar method to how SharedSecrets are created. This is stored in a map that is read from when a request for a Keytab is made.

Resilency or redundancy can be achieved by running more then one instance of the Tokenmachine server. This should work without conflict as long as the seeds for each entity are the same. Once again it is important that the seeds remain secret.

### Operation

Operatioally the process works like this. Client is the user or machine that desires a SharedSecret or Keytab, Server is this (the TokenMachine) and Identity Provider (IDP) is the provider of tokens (outside our concern).

1. The Client obtains a token from their IDP
1. The Client uses the token to request a Nonce from the Server
1. The Server validates that the Client is authroized and returns a Nonce to the Client
1. The Client obtains a new token from their IDP with the Nonce value encoded somewhere in the token (such as audience)
1. The client request a SharedSecret or Keytab with the new token from the Server
1. The Server authorizes the request by checking to see if the Client is entitled to the requested entity and that a valid Nonce is present in the token

The authorization process for entitlement and nonce is done with an operator provided OPA/Rego policy.

## Installation

Tokenmachine is supported on Windows, Linux and Darwin. Keytabs can only be provided when running on a Windows server that is either a Domain Controller or is part of a Domain. For Linux and Darwin only dummy Keytabs will be provided.

### Windows

For Windows it is intended to run the server as a Windows service. The binary is able to install itself as a service and even start itself. Installation consist of placing the binary in a desired location, setting the config(s), installing the service and finally starting the service. It is importnat that the service be configured to run as a Directory Admin (this is required for the creation of keytabs). The config (covered in the configuration section) is one or more YAML, JSON or OPA/Rego files or URLs.

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

Note that you may specify more then one file my seperating the entries with a comma. For example

```bash
.\tokenmachine.exe service config set "C:\Program Files\TokenMachine\config.yaml,C:\other.yaml,https://github.com/myrepo/config.yaml"
```

It is a good idea to keep the file or files containing sensitive data such as seeds in a seperate file and give it restrictive permissions.

WHen you are ready you can start the service with the command

```bash
.\tokenmachine.exe service start
```

### Linux and Darwin

For Linux and Darwin download and install the binary tokenmachine. Note that valid Keytabs will NOT be issued on Linux and Darwin.

Create an example config with the command

```bash
./tokenmachine config example > example.yaml
```

Run the server with the command

```bash
./tokenmachine --config example.yaml
```

## Example

[Example client scripts](example)
