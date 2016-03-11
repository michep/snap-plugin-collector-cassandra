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
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
)

const (
	metricQuery    = "/serverbydomain?querynames=org.apache.cassandra.metrics:*&template=identity"
	mbeanQuery     = "/mbean?objectname="
	querySuffix    = "&template=identity"
	javaStringType = "java.lang.String"

	emptyRespErr        = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"
	readDocErr          = "Read document error"
	queryDocErr         = "Queried document not found"
	emptyNamespaceErr   = "To be collected metric namespace is empty"
	invalidNamespaceErr = "To be collected metric namespace is invalid"
)

var (
	cassLog = log.WithField("_module", "cass-collector-client")
)

// XMLServer represents Server element
type XMLServer struct {
	XMLName xml.Name  `xml:"Server"`
	Domain  XMLDomain `xml:"Domain"`
}

// XMLDomain represents Domain element
type XMLDomain struct {
	XMLName xml.Name   `xml:"Domain"`
	MBeans  []XMLMBean `xml:"MBean"`
}

// XMLMBean represents MBean element
type XMLMBean struct {
	XMLName    xml.Name `xml:"MBean"`
	Objectname string   `xml:"objectname,attr"`
}

//XMLAttribute represents list of Attribute elements
type XMLAttributes struct {
	XMLName    xml.Name       `xml:"MBean"`
	Attributes []XMLAttribute `xml:"Attribute"`
}

// XMLAttribute represents Attribute element
type XMLAttribute struct {
	XMLName xml.Name `xml:"Attribute"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
	Value   float64  `xml:"value,attr"`
}

// CassClient defines the URL of Cassandra
type CassClient struct {
	client *HTTPClient
	host   string
}

// NewCassClient returns a new instance of CassClient
func NewCassClient(url, host string) *CassClient {
	return &CassClient{
		client: NewHTTPClient(url, "", timeout),
		host:   host,
	}
}

// getMetricType returns all available metric types. It exits if a fatal error occurs.
func (cc *CassClient) getMetricType() []plugin.PluginMetricType {
	resp, err := cc.client.httpClient.Get(cc.client.GetUrl() + metricQuery)
	if err != nil {
		log.Fatal(err.Error())
	}

	mbeans, err := readObjectname(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
	}

	mtsType := []plugin.PluginMetricType{}
	for _, mbean := range mbeans {
		ns := []string{"intel", "cassandra", "node", cc.host}
		ns = append(ns, makeNamespace(mbean.Objectname, "*")...)

		mtsType = append(mtsType, plugin.PluginMetricType{
			Namespace_: ns,
			Labels_:    []core.Label{{Index: len(ns) - 1, Name: "details"}}})
	}
	return mtsType
}

// getData returns a list of collected metrics giving namespaces.
// It logs invalid URLs(namespaces) but ignores the errors.
func (cc *CassClient) getData(ns []string) ([]plugin.PluginMetricType, error) {
	url, err := cc.getQueryURL(ns)
	if err != nil {
		return nil, err
	}
	return cc.worker(url)
}

func (cc *CassClient) worker(url string) ([]plugin.PluginMetricType, error) {
	resp, err := cc.client.httpClient.Get(cc.client.GetUrl() + mbeanQuery + url + querySuffix)
	if err != nil {
		cassLog.WithFields(log.Fields{
			"_block": "worker",
			"error":  err,
		}).Error(readDocErr)
		return nil, err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil || string(contents) == emptyRespErr {
		cassLog.WithFields(log.Fields{
			"_block": "worker",
			"error":  err,
		}).Error(queryDocErr)
		return nil, errors.New(queryDocErr)
	}

	attrs, _ := readAttrbutes(resp.Body)
	mts := []plugin.PluginMetricType{}
	for _, attr := range attrs {
		if attr.Type != javaStringType {
			mts = append(mts, cc.buildMetric(attr.Name, url, attr.Value))
		}
	}
	return mts, nil
}

// getQueryURL returns the MX4J URL from the giving metric namespace
func (cc *CassClient) getQueryURL(ns []string) (string, error) {
	if len(ns) == 0 || len(ns) < 6 {
		cassLog.WithFields(log.Fields{
			"_block": "getUrl",
			"error":  errors.New(invalidNamespaceErr),
		}).Error(emptyNamespaceErr)
		return "", errors.New(invalidNamespaceErr)
	} else if ns[0] != "intel" && ns[1] != "cassandra" && ns[2] != "node" && ns[4] != "org.apache.cassandra.metrics" {
		cassLog.WithFields(log.Fields{
			"_block": "getUrl",
			"error":  errors.New(invalidNamespaceErr + strings.Join(ns, "/")),
		}).Error("To be collected metric namespace is invalid")
		return "", errors.New(invalidNamespaceErr)
	}

	// Builds MX4J query URL and
	// ignores the last one while building the url params
	params := []string{}
	for i := 5; i < len(ns)-1; i = (i + 2) {
		params = append(params, ns[i]+"="+ns[i+1])
	}
	url := ns[4] + ":" + strings.Join(params, ",")
	return url, nil
}

func (cc *CassClient) buildMetric(name, url string, value float64) plugin.PluginMetricType {
	ns := makeNamespace(url, name)
	mts := plugin.PluginMetricType{
		Namespace_: ns,
		Data_:      value,
		Source_:    cc.host,
		Timestamp_: time.Now(),
	}
	return mts
}

func makeNamespace(url, name string) []string {
	ns := []string{}

	sp := strings.Split(url, ":")
	ns = append(ns, sp[0])

	sp1 := strings.Split(sp[1], ",")
	for _, s := range sp1 {
		v := strings.Split(s, "=")
		ns = append(ns, v...)
	}

	if name != "" {
		ns = append(ns, name)
	}
	return ns
}

func readObjectname(reader io.Reader) ([]XMLMBean, error) {
	var xmlServer XMLServer
	err := xml.NewDecoder(reader).Decode(&xmlServer)
	if err != nil {
		return nil, err
	}
	return xmlServer.Domain.MBeans, nil
}

func readAttrbutes(reader io.Reader) ([]XMLAttribute, error) {
	var xmlAttributes XMLAttributes
	if err := xml.NewDecoder(reader).Decode(&xmlAttributes); err != nil {
		return nil, err
	}
	return xmlAttributes.Attributes, nil
}
