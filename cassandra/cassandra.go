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
	"strconv"
	"time"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

const (
	// Name of plugin
	name = "cassandra"
	// Version of plugin
	version = 1
	// Type of plugin
	pluginType = plugin.CollectorPluginType

	// Timeout duration
	timeout = 5 * time.Second

	cassURL    = "url"
	port       = "port"
	hostname   = "hostname"
	invalidURL = "Invalid URL in Global configuration"
	noHostname = "No hostname define in Global configuration"
)

// Meta returns the snap plug.PluginMeta type
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

//  NewCassandraCollector returns a new instance of Cassandra struct
func NewCassandraCollector() *Cassandra {
	return &Cassandra{}
}

type Cassandra struct {
}

// CollectMetrics collects metrics from Cassandra through JMX
func (p *Cassandra) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {
	metrics := []plugin.PluginMetricType{}
	client, err := initClient(mts[0])
	if err != nil {
		return nil, err
	}

	for _, m := range mts {
		dpt, err := client.getData(m.Namespace())
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, dpt...)
	}
	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by Elasticsearch
func (p *Cassandra) GetMetricTypes(cfg plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
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
	items, err := config.GetConfigItems(cfg, []string{cassURL, port, hostname})
	if err != nil {
		return nil, err
	}

	url := items[cassURL].(string)
	hostname := items[hostname].(string)
	port := items[port].(int)

	if url == "" {
		return nil, errors.New(invalidURL)
	}
	if hostname == "" {
		return nil, errors.New(noHostname)
	}

	server := url + ":" + strconv.Itoa(port)

	return NewCassClient(server, hostname), nil
}
