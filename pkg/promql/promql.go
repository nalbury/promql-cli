/*
Copyright © 2020 Nick Albury nickalbury@gmail.com

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

// promql provides an abstraction on the prometheus HTTP API
package promql

import (
	"fmt"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
)

// Client is our prometheus v1 API interface
type Client interface {
	v1.API
}

// CreateClient creates a Client interface for the provided hostname
func CreateClient(host string) (Client, error) {
	a, err := api.NewClient(api.Config{
		Address: host,
	})
	if err != nil {
		return nil, err
	}
	return v1.NewAPI(a), nil
}

// CreateClientWithAuth creates a Client interface witht the provided hostname and auth config
func CreateClientWithAuth(host string, authCfg config.Authorization) (Client, error) {
	cfg := api.Config{
		Address: host,
	}
	if authCfg != (config.Authorization{}) {
		switch {
		case authCfg.Type == "":
			return nil, fmt.Errorf("please specify an authentication type, run promql --help for more details")
		case authCfg.Credentials != "" && authCfg.CredentialsFile != "":
			return nil, fmt.Errorf("please specify either auth credentials or an auth credential file, not both")
		case authCfg.Credentials != "":
			cfg.RoundTripper = config.NewAuthorizationCredentialsRoundTripper(authCfg.Type, config.Secret(authCfg.Credentials), api.DefaultRoundTripper)
		default:
			cfg.RoundTripper = config.NewAuthorizationCredentialsFileRoundTripper(authCfg.Type, authCfg.CredentialsFile, api.DefaultRoundTripper)
		}
	}
	a, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return v1.NewAPI(a), nil
}
