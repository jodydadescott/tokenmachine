# Tokenmachine
Grants SharedSecrets, Kerberos Keytabs, and Nonces to bearers of authorized tokens.
Authorization is accompolished by executing an OPA/Rego policy configured by the operator.
Shared Secrets are for applications that require a common shared secret. Keytabs are for
Kerberos authentication and Nonces are provided as a method to prevent replay attacks.

## Operation

When an application or users (client) desires a resource from the Tokenmachine they must first
acquire an OAUTH compliant token (bearer token) from their Identity Provider (IDP). Using 
this bearer token the client should request a Nonce from the Tokenmachine. The Tokenmachine will
authorize the Nonce request and if authroized return a time scoped Nonce. The Nonce will contain
both a secret and an expiration time. The client should obtain a new token from their IDP with
the Nonce secret encoded in the claims. The client should then use the bearer token to make one
or more request to the Tokenmachine before the Nonce expires.

SharedSecrets contain a Secret and an expiration (exp) time in epoch seconds. If the SharedSecret has
reached half life then the fields NextSecret and NextExp will be present. These fields represent the
next period Secret and expiration. It is up to the user to make use of the SharedSecret and cordinate
the period changes.

Keytabs contain a base64 Keytab file in the field Base64File and a Kerberos principal in the 
field Principal. It is not currently possible to provide the next Keytab as it is for a 
SharedSecret. Generally the user should obtain a Keytab when required and then discard it.

Authentication and authorization is performed by validating a bearers token signature and executing
a pre-configured OPA/Rego policy. The signature validation works by following the tokens issuer and
obtaining the public key. The issuer name must match the TLS certificate in the HTTPS request. The
OPA/Rego policy may also be used to prevent replay attack by checking to see if a nonce pattern is
inside the payload of the bearer token.

The Tokenmachine server works by being preconfigured with SharedSecrets and Keytabs. We will
referfer to these as entities for short. Both entities must be configured with a seed. This 
seed should be kept secret as it will be used to generate the secret for the SharedSecret
and the password for the Keytab principal. Each entity must also be configured with a Lifetime
in seconds. Each entity will have a time period calculated by taking the seconds since epoch and
dividing it by the Lifetime and removing the remainder. This will provide us with the current
period and by adding the lifetime to this we can determine the next period.

When a request for a SharedSecret is made the current time period is determined and the
seed is used to obtain a one time code. This is then appended to the seed and a SHA-256
hash is calculated. This will be converted into a alpha-numeric string.

Keytabs operate in a similiar method but there are signifigant differences. At the top of each
period Keytabs are created using a password that is derived in a similiar method to how SharedSecrets
are created. This is stored in a map that is read from when a request for a Keytab is made.

Resilency or redundancy can be achieved by running more then one instance of the Tokenmachine server.
This should work without conflict as long as the seeds for each entity are the same. Once again it is
important that the seeds remain secret.

## Format

The SharedSecret response format is as follows. Note that NextSecret and NextExp will only be
present if the half-life has been reached.
```golang
type SharedSecret struct {
	Name       string        `json:"name,omitempty" yaml:"name,omitempty"`
	Exp        int64         `json:"exp,omitempty" yaml:"exp,omitempty"`
	Secret     string        `json:"secret,omitempty" yaml:"secret,omitempty"`
	NextExp    int64         `json:"nextExp,omitempty" yaml:"nextExp,omitempty"`
	NextSecret string        `json:"nextSecret,omitempty" yaml:"nextSecret,omitempty"`
}
```

The Keytab response format is as follows.
```golang
type Keytab struct {
	Principal  string        `json:"principal,omitempty" yaml:"principal,omitempty"`
	Seed       string        `json:"seed,omitempty" yaml:"seed,omitempty"`
	Base64File string        `json:"base64file,omitempty" yaml:"base64file,omitempty"`
	Exp        int64         `json:"exp,omitempty" yaml:"exp,omitempty"`
	Lifetime   time.Duration `json:"lifetime,omitempty" yaml:"lifetime,omitempty"`
}
```

## Installation

For Windows it is intended to run the server as a Windows service. The binary is able
to install itself as a service and even start itself. Installation consist of placing
the binary in a desired location, setting the config(s), installing the service and
finally starting the service.

It is importnat that the service be configured to run as a Directory Admin. This is
required for the creation of keytabs.

The config will be covered below but on Windows a comma delineated list of configs is
stored in the Windows registry. This can be set and shown using the same binary.

Create a directory such as 'C:\Program Files\Tokens2Secrets' and install the binary
'tokens2Secrets.exe' into said direcotry. Using the Windows command prompt change
directory to the installation directory and perform the following task.

```bash
# Install the service
.\tokenmachine.exe service install

# Create an example config then edit it
.\tokenmachine.exe config example > example.yaml

# Set the config(s) where $CONFIG_FILES with the location to one or more config files or
#  http(s) urls
.\tokenmachine.exe service config set $CONFIG_FILES

# When ready start the service
.\tokenmachine.exe service start
```

For Linux and Darwin download and install the binary tokenmachine. Note that valid Keytabs will NOT
be issued on Linux and Darwin.

```bash
# Create an example config then edit it
./tokenmachine config example > example.yaml

# Run the server with
./tokenmachine --config example.yaml
```

## Configuration

The config is provided by one or more files that may be JSON, YAML, or Rego. The config
may be located on the local disk or on a remote http(s) server such as GitHub. The config
can be in a single file or multiple files. When the config is located in multiple files
the config will be processed in order. It is possible to use both local and remote
configs.

The configuration requires a secret seed. This should remain secret. This can be
accompolished by breaking the configuration up into two files where the main file
contains all of the non-sensitive config and the other file contains the seed secret.
The main file can be located on a repo an accessible via http(s) or on the local
filesystem. The seed config should be stored on the local filesystem and with restrictive
read/write permissions.

One method would be to locate the non-sensitive config on a remote http(s) server and the
sensitive config on the local disk with restrictive read and write permissions. Or both
configs could be kept on disk as seperate files. Another method would be to keep the
non-sensitive data on a repo and the seed in a vault. Then they could be combined into a
single config by a script at runtime and set with restricitive read and write
permissions.

This is an example with in-line supporting comments.
```yaml
# This is the config version.
apiVersion: V1
network:
  # THe interface to listen on. By default it is any.
  # listen: any
  # The port to run the HTTP and or HTTPS server on. If it is not set then it will not
  # be started. Either the HTTP or HTTPS server must be enabled. If the HTTPS server is
  # enabled both tlscert and tlsKey must be set
  httpPort: 8080
  # httpsPort: 8443
  # tlscert: |-
  #   -----BEGIN CERTIFICATE-----
  #   ................................................................
  #   ................................................................
  #   ................................................................
  #   ................................................................
  #   ................................................................
  #   ................................................................
  #   ................................................................
  #   ................................................................
  #   ................................................................
  #   ................................................................
  #   ................................................................
  #   ....................................
  #   -----END CERTIFICATE-----
  # tlsKey: |-
  #   -----BEGIN EC PRIVATE KEY-----
  #   ................................................................
  #   ................................................................
  #   ................................................................
  #   ................................
  #   -----END EC PRIVATE KEY-----
policy:
  #
  # This is a OPA/Rego Policy. It must have the package name main and expose
  # the boolean functions auth_get_nonce and auth_get_keytab.
  #
  # The input format structure has claims []string, principal and nonce string.
  #
  policy: |2

	  package main

	  default auth_get_nonce = false
	  default auth_get_keytab = false
	  default auth_get_secret = false
	
	  auth_base {
	    # Match Issuer
	    input.claims.iss == "abc123"
	  }
	
	  auth_get_nonce {
	    auth_base
	  }
	
	  auth_nonce {
	    # The input contains a set of all of the current valid nonces. For our
	    # example here we expect the claim audience to have a nonce that will match
	    # one of tne entries in the nonces set.
	    input.nonces[_] == input.claims.aud
	  }
	
	  auth_get_keytab {
	    # The nonce must be validated and then the principal. This is done by splitting the
	    # principals in the claim service.keytab by the comma into a set and checking for
	    # match with requested principal
	    auth_base
	    auth_nonce
	    split(input.claims.service.keytab,",")[_] == input.principal
	  }
	
	  auth_get_secret {
	    auth_base
	    auth_nonce
	    input.claims.service.secrets[_] == input.secret
	  }
  # Nonce lifetime
  nonceLifetime: 1m0s

  # This is the lifetime that wll be used for a Keytab if no lifetime is
  # configured for the specific KeyTab
  keytabLifetime: 10m0s

logging:
  # This should be debug, info, warn or error. Default is info.
  logLevel: info
  # This should be json or console. Default is json.
  logFormat: json

  # The outputPaths and errorOutputPaths should be one or more of stderr, stdout, and/or
  # file. By default it is stderr. Note that on Windows when running as a service events
  # will be sent to the Windows event logger irregardless of this setting.
  outputPaths:
    - stderr
  errorOutputPaths:
    - stderr
data:
  # This is where we configure all of the Keytabs and secrets. The field name
  # and seed is required and the field lifetime is optional. If the lifetine is
  # not specified then the default lifetime will be used. The default lifetime
  # is configurable in policy. Note that the seed is what we will derive passwords
  # and secrets from so it should be kept secret.
  keytabs:
    # Keytab with lifetime set
    - name: superman@EXAMPLE.COM
      seed: nIKSXX9nJU5klguCrzP3d
      lifetime: 1m0s
    # Keytab that will use the default lifetime from above
    - name: birdman@EXAMPLE.COM
      seed: CibIcE3XhRyXrngddsQzN
  secrets:
    # Secret with lifetime set
    - name: secret1
      seed: E17cUHMYtU+FvpK3kig7o5
      lifetime: 1m0s
    # Secret that will use the default lifetime
    - name: secret2
      seed: 7Y3dzQcEvx+cPpRl4Qgti2
```

## Example(s)

[Example client scripts](example)