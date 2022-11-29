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
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
) // Client is our prometheus v1 API interface

type Client interface {
	v1.API
}

// CreateClient creates a Client interface for the provided hostname
func CreateClient(host string) (v1.API, error) {
	a, err := api.NewClient(api.Config{
		Address: host,
	})
	if err != nil {
		return nil, err
	}
	return v1.NewAPI(a), nil
}

// CreateClientWithAuth creates a Client interface witht the provided hostname and auth config
func CreateClientWithAuth(host string, authCfg config.Authorization, tlsCfg config.TLSConfig) (v1.API, error) {
	cfg := api.Config{
		Address: host,
	}
	cmmnConfig := config.HTTPClientConfig{
		TLSConfig: tlsCfg,
	}
	rt, err := config.NewRoundTripperFromConfig(cmmnConfig, "promql")
	if err != nil {
		return nil, err
	}
	if authCfg != (config.Authorization{}) {
		switch {
		case authCfg.Type == "":
			return nil, fmt.Errorf("please specify an authentication type, run promql --help for more details")
		case authCfg.Credentials != "" && authCfg.CredentialsFile != "":
			return nil, fmt.Errorf("please specify either auth credentials or an auth credential file, not both")
		case authCfg.Credentials != "":
			cfg.RoundTripper = config.NewAuthorizationCredentialsRoundTripper(authCfg.Type, config.Secret(authCfg.Credentials), rt)
		default:
			cfg.RoundTripper = config.NewAuthorizationCredentialsFileRoundTripper(authCfg.Type, authCfg.CredentialsFile, rt)
		}
	}
	cfg.RoundTripper = rt
	a, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return v1.NewAPI(a), nil
}

// Cfg conatins the final configuration params parsed from a combo of flags, config file values, and env vars.
type PromQL struct {
	Host            string
	Step            string
	Output          string
	TimeoutDuration time.Duration
	CfgFile         string
	Time            time.Time
	Start           string
	End             string
	NoHeaders       bool
	Auth            config.Authorization
	Client          v1.API
	TLSConfig       config.TLSConfig
}

// InstantQuery performs an instant query and returns the result
func (p *PromQL) InstantQuery(queryString string) (model.Vector, v1.Warnings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.TimeoutDuration)
	defer cancel()

	result, warnings, err := p.Client.Query(ctx, queryString, p.Time)
	if err != nil {
		return nil, warnings, fmt.Errorf("error querying prometheus: %v", err)
	}

	if result, ok := result.(model.Vector); ok {
		return result, warnings, nil
	} else {
		return nil, warnings, fmt.Errorf("did not receive an instant vector result")
	}
}

// getRange creates a prometheus range from the provided start, end, and step options
func (p *PromQL) getRange() (r v1.Range, err error) {
	// At minimum we need a start time so we attempt to parse that first
	if s, err := time.Parse(time.RFC3339, p.Start); err == nil {
		r.Start = s
	} else if l, err := time.ParseDuration(p.Start); err == nil {
		r.Start = time.Now().Add(-l)
	} else {
		err = fmt.Errorf("unable to parse range start time, %v", err)
		return r, err
	}

	// Set up defaults for the range end and step values
	r.End = time.Now()
	r.Step = time.Minute

	// If the user provided a step value, parse it as a time.Duration and override the default
	if p.Step != "" {
		r.Step, err = time.ParseDuration(p.Step)
		if err != nil {
			err = fmt.Errorf("unable to parse step duration, %v", err)
			return r, err
		}
	}

	// If the user provided an end value, parse it to a time struct and override the default
	if p.End != "now" {
		e, err := time.Parse(time.RFC3339, p.End)
		if err != nil {
			err = fmt.Errorf("unable to parse range end time, %v", err)
			return r, err
		}
		r.End = e
	}

	return r, err
}

// rangeQuery performs a range query and writes the results to stdout
func (p *PromQL) RangeQuery(queryString string) (model.Matrix, v1.Warnings, error) {
	// create context with a timeout,
	ctx, cancel := context.WithTimeout(context.Background(), p.TimeoutDuration)
	defer cancel()

	r, err := p.getRange()
	if err != nil {
		return nil, nil, err
	}
	// execute query
	result, warnings, err := p.Client.QueryRange(ctx, queryString, r)
	if err != nil {
		return nil, warnings, err
	}

	if result, ok := result.(model.Matrix); ok {
		return result, warnings, err
	} else {
		return nil, warnings, fmt.Errorf("did not receive a range result")
	}
}

// LabelsQuery runs a labels query and returns the result
func (p *PromQL) LabelsQuery(query string) (model.Vector, v1.Warnings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.TimeoutDuration)
	defer cancel()

	result, warnings, err := p.Client.Query(ctx, query, time.Now())
	if err != nil {
		return nil, warnings, err
	}

	// if result is the expected type, Write it out in the
	// desired output format
	if result, ok := result.(model.Vector); ok {
		return result, warnings, err
	} else {
		return nil, warnings, fmt.Errorf("did not recieve an instant vector")
	}
}

// MetaQuery returns prometheus metrics metadata. Used for our metrics and meta commands
func (p *PromQL) MetaQuery(query string) (map[string][]v1.Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.TimeoutDuration)
	defer cancel()

	result, err := p.Client.Metadata(ctx, query, "")
	if err != nil {
		return map[string][]v1.Metadata{}, fmt.Errorf("Error querying metadata endpoint: %v", err)
	}
	return result, nil
}
