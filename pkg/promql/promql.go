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
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
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
