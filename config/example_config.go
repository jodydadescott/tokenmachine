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

var examplePolicy = `
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
   # Verify that the request nonce matches the expected nonce. Our token provider
   # has the nonce in the audience field under claims
   input.claims.aud == input.nonce
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
   # Verify that the request nonce matches the expected nonce. Our token provider
   # has the nonce in the audience field under claims
   auth_base
   auth_nonce
}
`

var exampleTLSCert = `-----BEGIN CERTIFICATE-----
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

var exampleTLSKey = `-----BEGIN EC PRIVATE KEY-----
................................................................
................................................................
................................................................
................................
-----END EC PRIVATE KEY-----`

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
			NonceLifetime:  60,
			KeytabLifetime: 300,
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
					Principal: "superman@EXAMPLE.COM",
					Seed:      "nIKSXX9nJU5klguCrzP3d",
					Lifetime:  time.Duration(60) * time.Second,
				},
				&Keytab{
					Principal: "birdman@EXAMPLE.COM",
					Seed:      "CibIcE3XhRyXrngddsQzN",
					Lifetime:  time.Duration(60) * time.Second,
				},
			},

			Secrets: []*Secret{
				&Secret{
					Name:     "secret1",
					Seed:     "E17cUHMYtU+FvpK3kig7o5",
					Lifetime: time.Duration(60) * time.Second,
				},
				&Secret{
					Name:     "secret2",
					Seed:     "7Y3dzQcEvx+cPpRl4Qgti2",
					Lifetime: time.Duration(120) * time.Second,
				},
				&Secret{
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
