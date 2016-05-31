/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2016 Intel Corporation

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

package cassandra

import (
	"errors"
	"fmt"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

const (
	// Name of plugin
	Name = "cassandra"
	// Version of plugin
	Version = 2
	// Type of plugin
	PluginType = plugin.CollectorPluginType

	// Timeout duration
	DefaultTimeout = 5 * time.Second

	CassURL    = "url"
	Port       = "port"
	Hostname   = "hostname"
	InvalidURL = "Invalid URL in Global configuration"
	NoHostname = "No hostname define in Global configuration"
)

// Meta returns the snap plug.PluginMeta type
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, PluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

// NewCassandraCollector returns a new instance of Cassandra struct
func NewCassandraCollector() *Cassandra {
	return &Cassandra{}
}

// Cassandra struct
type Cassandra struct {
}

// CollectMetrics collects metrics from Cassandra through JMX
func (p *Cassandra) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	metrics := []plugin.MetricType{}
	client, err := initClient(mts[0])
	if err != nil {
		return nil, err
	}

	for _, m := range mts {
		dpt, err := client.getData(m.Namespace().Strings())
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, dpt...)
	}
	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by Elasticsearch
func (p *Cassandra) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	client, err := initClient(cfg)
	if err != nil {
		return nil, err
	}
	metricTypes := client.getMetricType()
	return metricTypes, nil
}

// GetConfigPolicy returns a ConfigPolicy
func (p *Cassandra) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	return c, nil
}

func initClient(cfg interface{}) (*CassClient, error) {
	items, err := config.GetConfigItems(cfg, CassURL, Port, Hostname)
	if err != nil {
		return nil, err
	}

	url := items[CassURL].(string)
	hostname := items[Hostname].(string)
	port := items[Port].(int)

	if url == "" {
		return nil, errors.New(InvalidURL)
	}
	if hostname == "" {
		return nil, errors.New(NoHostname)
	}

	server := fmt.Sprintf("%s:%d", url, port)

	return NewCassClient(server, Hostname), nil
}
