/*
Copyright © 2020 Jody Scott <jody@thescottsweb.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"encoding/json"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	examplePolicy = `
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
   # 1) Validate defaults with auth_base
   # 2) Validate the nonce with auth_nonce
   # 3) Validate the claims are allowed to obtain the keytab with the given name
   #    by checking to see if name exist in claim keytabs. We split the value
   #    on colon and look for any match.
   auth_base
   auth_nonce
   split(input.claims.service.keytabs,":")[_] == input.name
}

auth_get_secret {
   # 1) Validate defaults with auth_base
   # 2) Validate the nonce with auth_nonce
   # 3) Validate the claims are allowed to obtain the secret with the given name
   #    by checking to see if name exist in claim secrets. We split the value
   #    on comma and look for any match.
   auth_base
   auth_nonce
   split(input.claims.service.secrets,":")[_] == input.name
}
`

	exampleTLSCert = `-----BEGIN CERTIFICATE-----
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
................................................................
....................................
-----END CERTIFICATE-----`

	exampleTLSKey = `-----BEGIN EC PRIVATE KEY-----
................................................................
................................................................
................................................................
................................
-----END EC PRIVATE KEY-----`
)

// NewV1ExampleConfig New example config
func NewV1ExampleConfig() *Config {
	return &Config{
		APIVersion: "V1",
		Network: &Network{
			Listen:    "any",
			HTTPPort:  8080,
			HTTPSPort: 8443,
			TLSCert:   exampleTLSCert,
			TLSKey:    exampleTLSKey,
		},
		Policy: &Policy{
			Policy:         examplePolicy,
			NonceLifetime:  time.Duration(60) * time.Second,
			KeytabLifetime: time.Duration(600) * time.Second,
		},
		Logging: &Logging{
			LogLevel:         "debug",
			LogFormat:        "json",
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		},
		Data: &Data{
			Keytabs: []*Keytab{
				&Keytab{
					Name:      "superman",
					Principal: "superman@EXAMPLE.COM",
					Seed:      "nIKSXX9nJU5klguCrzP3d",
					Lifetime:  time.Duration(60) * time.Second,
				},
				&Keytab{
					Name:      "birdman",
					Principal: "birdman@EXAMPLE.COM",
					Seed:      "CibIcE3XhRyXrngddsQzN",
					Lifetime:  time.Duration(60) * time.Second,
				},
			},

			SharedSecrets: []*SharedSecret{
				&SharedSecret{
					Name:     "secret1",
					Seed:     "E17cUHMYtU+FvpK3kig7o5",
					Lifetime: time.Duration(60) * time.Second,
				},
				&SharedSecret{
					Name:     "secret2",
					Seed:     "7Y3dzQcEvx+cPpRl4Qgti2",
					Lifetime: time.Duration(120) * time.Second,
				},
				&SharedSecret{
					Name:     "secret3",
					Seed:     "6zarcky7proZTYw8PEVzzT",
					Lifetime: time.Duration(240) * time.Second,
				},
			},
		},
	}
}

// ExampleConfigJSON Return example config as YAML
func ExampleConfigJSON() string {
	j, _ := json.Marshal(NewV1ExampleConfig())
	return string(j)
}

// ExampleConfigYAML Return example config as YAML
func ExampleConfigYAML() string {
	j, _ := yaml.Marshal(NewV1ExampleConfig())
	return string(j)
}
