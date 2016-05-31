//
// +build integration

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
	"log"
	"os"
	"testing"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCassandraCollectMetrics(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	server := os.Getenv("SNAP_CASSANDRA_HOST")
	if server == "" {
		log.Fatal("server SNAP_CASSANDRA_HOST is not exported.")
	}
	cfg := setupCfg(os.Getenv("SNAP_CASSANDRA_HOST"), hostname, 8082)

	Convey("Cassandra collector", t, func() {
		p := NewCassandraCollector()
		p.GetMetricTypes(cfg)

		Convey("collect multiple metrics", func() {
			mts := []plugin.MetricType{
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "cassandra", "node", hostname,
						"org.apache.cassandra.metrics", "type", "ThreadPools", "path", "internal", "scope", "ValidationExecutor",
						"name", "MaxPoolSize", "Value"),
					Config_: cfg.ConfigDataNode,
				},
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "cassandra", "node", hostname,
						"org.apache.cassandra.metrics", "type", "Keyspace", "keyspace", "system", "name", "TombstoneScannedHistogram",
						"Max"),
					Config_: cfg.ConfigDataNode,
				},
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "cassandra", "node", hostname,
						"org.apache.cassandra.metrics", "type", "Keyspace", "keyspace", "system", "name", "BloomFilterOffHeapMemoryUsed",
						"Value"),
					Config_: cfg.ConfigDataNode,
				},
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "cassandra", "node", hostname,
						"org.apache.cassandra.metrics", "type", "Keyspace", "keyspace", "system_auth", "name", "ReadLatency",
						"75thPercentile"),
					Config_: cfg.ConfigDataNode,
				},

				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "cassandra", "node", hostname,
						"org.apache.cassandra.metrics", "type", "Cache", "scope", "RowCache", "name",
						"Hits", "*"),
					Config_: cfg.ConfigDataNode,
				},
			}
			metrics, err := p.CollectMetrics(mts)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeNil)
		})

		Convey("collect a metric with a wrong namespace", func() {
			mts := []plugin.MetricType{
				plugin.MetricType{
					Namespace_: core.NewNamespace("node", hostname,
						"org.apache.cassandra.metrics", "type", "ThreadPools", "path", "internal", "scope", "ValidationExecutor",
						"name", "MaxPoolSize", "Value"),
					Config_: cfg.ConfigDataNode,
				},
			}
			metrics, err := p.CollectMetrics(mts)
			So(err, ShouldNotBeNil)
			So(metrics, ShouldBeNil)
		})

		Convey("collect a metric with an empty namespace", func() {
			mts := []plugin.MetricType{
				plugin.MetricType{
					Namespace_: core.NewNamespace(""),
					Config_:    cfg.ConfigDataNode,
				},
			}
			metrics, err := p.CollectMetrics(mts)
			So(err, ShouldNotBeNil)
			So(metrics, ShouldBeNil)
		})

		Convey("collect a metric with a not found namespace", func() {
			mts := []plugin.MetricType{
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "cassandra", "node", hostname,
						"org.apache.cassandra.metrics", "type", "abc"),
					Config_: cfg.ConfigDataNode,
				},
				plugin.MetricType{
					Namespace_: core.NewNamespace("intel", "cassandra", "node", hostname,
						"org.apache.cassandra.metrics", "type, Keyspace", "keyspace", "system_auth", "name", "ReadLatency",
						"90thPercentile"),
					Config_: cfg.ConfigDataNode,
				},
			}

			metrics, err := p.CollectMetrics(mts)
			So(err, ShouldNotBeNil)
			So(metrics, ShouldBeNil)
		})
	})
}

func setupCfg(url, hostname string, port int) plugin.ConfigType {
	node := cdata.NewNode()
	node.AddItem("url", ctypes.ConfigValueStr{Value: url})
	node.AddItem("hostname", ctypes.ConfigValueStr{Value: hostname})
	node.AddItem("port", ctypes.ConfigValueInt{Value: port})
	return plugin.ConfigType{ConfigDataNode: node}
}
