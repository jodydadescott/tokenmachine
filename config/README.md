# Configuration

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