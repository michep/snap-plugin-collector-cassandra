//
// #+build small

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
	"testing"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"fmt"
	"github.com/intelsdi-x/snap/core"
	"time"
)

func TestESPlugin(t *testing.T) {
	Convey("Meta should return metadata for the plugin", t, func() {
		meta := Meta()
		So(meta.Name, ShouldResemble, Name)
		So(meta.Version, ShouldResemble, Version)
		So(meta.Type, ShouldResemble, plugin.CollectorPluginType)
	})

	Convey("Create Cassandra Collector", t, func() {
		cassCol := NewCassandraCollector()
		Convey("So cassCol should not be nil", func() {
			So(cassCol, ShouldNotBeNil)
		})
		Convey("So cassCol should be of Cassandra type", func() {
			So(cassCol, ShouldHaveSameTypeAs, &Cassandra{})
		})
		Convey("cassCol.GetConfigPolicy() should return a config policy", func() {
			configPolicy, _ := cassCol.GetConfigPolicy()
			Convey("So config policy should not be nil", func() {
				So(configPolicy, ShouldNotBeNil)
			})
			Convey("So config policy should be a cpolicy.ConfigPolicy", func() {
				So(configPolicy, ShouldHaveSameTypeAs, &cpolicy.ConfigPolicy{})
			})
		})
	})
}


func Test_BuildMetricAPI(t *testing.T) {
	var err error
	cfg := cdata.NewNode()
	cfg.AddItem("url", ctypes.ConfigValueStr{Value: "pushtst00.mfms"})
	cfg.AddItem("port", ctypes.ConfigValueInt{8081})
	cass := NewCassandraCollector()
	cass.client, err = initClient(plugin.ConfigType{ConfigDataNode: cfg})
	if err != nil {
		panic(err)
	}
	err = cass.client.buildMetricAPI()
	if err != nil {
		panic(err)
	}
}

func Test_BuildMetricType(t *testing.T) {
	var err error
	var mts []plugin.MetricType
	cfg := cdata.NewNode()
	cfg.AddItem("url", ctypes.ConfigValueStr{Value: "pushtst00.mfms"})
	cfg.AddItem("port", ctypes.ConfigValueInt{8081})
	cass := NewCassandraCollector()
	mts, err = cass.client.buildMetricType(plugin.ConfigType{ConfigDataNode: cfg})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", mts)
}

func Test_CollectMetrics(t *testing.T) {
	//var err error
	mts := []plugin.MetricType{}
	cfg := cdata.NewNode()
	cfg.AddItem("url", ctypes.ConfigValueStr{Value: "pushtst00.mfms"})
	cfg.AddItem("port", ctypes.ConfigValueInt{8081})
	m := plugin.NewMetricType(makeNamespace2(), time.Now(), nil, "", nil)
	m.Config_ = cfg
	mts = append(mts, *m)
	cass := NewCassandraCollector()
	mts, _ = cass.CollectMetrics(mts)
	fmt.Printf("%+v", len(mts))

}


func makeNamespace1() core.Namespace {
	ns := core.NewNamespace("intel", "cassandra", "node")
	ns = ns.AddDynamicElement("node", "node")
	ns = ns.AddStaticElements("org_apache_cassandra_metrics", "type")
	ns = ns.AddDynamicElement("type", "type")
	ns = ns.AddStaticElement("scope")
	ns = ns.AddDynamicElement("scope", "scope")
	ns = ns.AddStaticElement("name")
	ns = ns.AddDynamicElement("name", "name")
	ns = ns.AddStaticElement("Count")
	return ns
}

func makeNamespace2() core.Namespace {
	ns := core.NewNamespace("intel", "cassandra", "node")
	ns = ns.AddDynamicElement("node", "node")
	ns = ns.AddStaticElements("org_apache_cassandra_metrics", "type")
	ns = ns.AddDynamicElement("type", "type")
	ns = ns.AddStaticElement("keyspace")
	ns = ns.AddDynamicElement("keyspace", "keyspace")
	ns = ns.AddStaticElement("scope")
	ns = ns.AddDynamicElement("scope", "scope")
	ns = ns.AddStaticElement("name")
	ns = ns.AddDynamicElement("name", "name")
	ns = ns.AddStaticElement("Value")
	return ns
}
