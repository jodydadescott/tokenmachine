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

	"github.com/jinzhu/copier"
	"gopkg.in/yaml.v2"
)

// Config Config
type Config struct {
	APIVersion string   `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Network    *Network `json:"network,omitempty" yaml:"network,omitempty"`
	Policy     *Policy  `json:"policy,omitempty" yaml:"policy,omitempty"`
	Logging    *Logging `json:"logging,omitempty" yaml:"logging,omitempty"`
	Data       *Data    `json:"data,omitempty" yaml:"data,omitempty"`
}

// Network Config
type Network struct {
	Listen    string `json:"listen,omitempty" yaml:"listen,omitempty"`
	HTTPPort  int    `json:"httpPort,omitempty" yaml:"httpPort,omitempty"`
	HTTPSPort int    `json:"httpsPort,omitempty" yaml:"httpsPort,omitempty"`
	TLSCert   string `json:"tlscert,omitempty" yaml:"tlscert,omitempty"`
	TLSKey    string `json:"tlsKey,omitempty" yaml:"tlsKey,omitempty"`
}

// Policy Config
type Policy struct {
	Policy               string        `json:"policy,omitempty" yaml:"policy,omitempty"`
	NonceLifetime        time.Duration `json:"nonceLifetime,omitempty" yaml:"nonceLifetime,omitempty"`
	SharedSecretLifetime time.Duration `json:"sharedSecretLifetime,omitempty" yaml:"sharedSecretLifetime,omitempty"`
	KeytabLifetime       time.Duration `json:"keytabLifetime,omitempty" yaml:"keytabLifetime,omitempty"`
}

// Logging Config
type Logging struct {
	LogLevel         string   `json:"logLevel,omitempty" yaml:"logLevel,omitempty"`
	LogFormat        string   `json:"logFormat,omitempty" yaml:"logFormat,omitempty"`
	OutputPaths      []string `json:"outputPaths,omitempty" yaml:"outputPaths,omitempty"`
	ErrorOutputPaths []string `json:"errorOutputPaths,omitempty" yaml:"errorOutputPaths,omitempty"`
}

// Data Config
type Data struct {
	SharedSecrets []*SharedSecret `json:"sharedSecrets,omitempty" yaml:"sharedSecrets,omitempty"`
	Keytabs       []*Keytab       `json:"keytabs,omitempty" yaml:"keytabs,omitempty"`
}

// SharedSecret Config
type SharedSecret struct {
	Name     string        `json:"name,omitempty" yaml:"name,omitempty"`
	Seed     string        `json:"seed,omitempty" yaml:"seed,omitempty"`
	Lifetime time.Duration `json:"lifetime,omitempty" yaml:"lifetime,omitempty"`
}

// Keytab Config
type Keytab struct {
	Name      string        `json:"name,omitempty" yaml:"name,omitempty"`
	Principal string        `json:"principal,omitempty" yaml:"principal,omitempty"`
	Seed      string        `json:"seed,omitempty" yaml:"seed,omitempty"`
	Lifetime  time.Duration `json:"lifetime,omitempty" yaml:"lifetime,omitempty"`
}

func (t *Data) addSharedSecret(sharedSecret *SharedSecret) {

	var existing *SharedSecret
	for _, v := range t.SharedSecrets {
		if v.Name == sharedSecret.Name {
			existing = v
		}
	}

	if existing == nil {
		t.SharedSecrets = append(t.SharedSecrets, sharedSecret)
		return
	}

	if sharedSecret.Seed != "" {
		existing.Seed = sharedSecret.Seed
	}

	if sharedSecret.Lifetime > 0 {
		existing.Lifetime = sharedSecret.Lifetime
	}

}

func (t *Data) addKeytab(keytab *Keytab) {

	var existing *Keytab
	for _, v := range t.Keytabs {
		if v.Name == keytab.Name {
			existing = v
		}
	}

	if existing == nil {
		t.Keytabs = append(t.Keytabs, keytab)
		return
	}

	if keytab.Principal != "" {
		existing.Principal = keytab.Principal
	}

	if keytab.Seed != "" {
		existing.Seed = keytab.Seed
	}

	if keytab.Lifetime > 0 {
		existing.Lifetime = keytab.Lifetime
	}

}

// NewConfig Returns new V1 Config
func NewConfig() *Config {
	return &Config{
		APIVersion: "V1",
		Network: &Network{
			Listen: "any",
		},
		Policy:  &Policy{},
		Logging: &Logging{},
		Data:    &Data{},
	}
}

// JSON Return JSON String representation
func (t *Config) JSON() string {
	j, _ := json.Marshal(t)
	return string(j)
}

// YAML Return YAML String representation
func (t *Config) YAML() string {
	j, _ := yaml.Marshal(t)
	return string(j)
}

// Copy return copy of entity
func (t *Config) Copy() *Config {
	clone := &Config{}
	copier.Copy(&clone, &t)
	return clone
}

// Merge Config into existing config
func (t *Config) Merge(config *Config) {

	config = config.Copy()

	if config.Network != nil {

		if t.Network == nil {
			t.Network = &Network{}
		}

		if config.Network.Listen != "" {
			t.Network.Listen = config.Network.Listen
		}

		if config.Network.HTTPPort > 0 {
			t.Network.HTTPPort = config.Network.HTTPPort
		}

		if config.Network.HTTPSPort > 0 {
			t.Network.HTTPSPort = config.Network.HTTPSPort
		}

		if config.Network.TLSKey != "" {
			t.Network.TLSKey = config.Network.TLSKey
		}

		if config.Network.TLSCert != "" {
			t.Network.TLSCert = config.Network.TLSCert
		}

	}

	if config.Policy != nil {

		if t.Policy == nil {
			t.Policy = &Policy{}
		}

		if config.Policy.Policy != "" {
			t.Policy.Policy = config.Policy.Policy
		}

		if config.Policy.NonceLifetime > 0 {
			t.Policy.NonceLifetime = config.Policy.NonceLifetime
		}

		if config.Policy.KeytabLifetime > 0 {
			t.Policy.KeytabLifetime = config.Policy.KeytabLifetime
		}

		if config.Policy.SharedSecretLifetime > 0 {
			t.Policy.SharedSecretLifetime = config.Policy.SharedSecretLifetime
		}

	}

	if config.Logging != nil {

		if t.Logging == nil {
			t.Logging = &Logging{}
		}

		if config.Logging.LogLevel != "" {
			t.Logging.LogLevel = config.Logging.LogLevel
		}

		if config.Logging.LogFormat != "" {
			t.Logging.LogFormat = config.Logging.LogFormat
		}

		if config.Logging.OutputPaths != nil {
			for _, s := range config.Logging.OutputPaths {
				if s != "" {
					t.Logging.OutputPaths = append(t.Logging.OutputPaths, s)
				}
			}
		}

		if config.Logging.ErrorOutputPaths != nil {
			for _, s := range config.Logging.ErrorOutputPaths {
				if s != "" {
					t.Logging.ErrorOutputPaths = append(t.Logging.ErrorOutputPaths, s)
				}
			}
		}

	}

	if config.Data != nil {

		if t.Data == nil {
			t.Data = &Data{}
		}

		if config.Data.Keytabs != nil {
			for _, s := range config.Data.Keytabs {
				t.Data.addKeytab(s)
			}
		}

		if config.Data.SharedSecrets != nil {
			for _, s := range config.Data.SharedSecrets {
				t.Data.addSharedSecret(s)
			}
		}

	}

}
