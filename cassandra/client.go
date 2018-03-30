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

	"github.com/intelsdi-x/snap/control/plugin"
	log "github.com/sirupsen/logrus"
	"fmt"
)

// const defines constant varaibles
const (
	MetricQuery    = "/serverbydomain?querynames=org.apache.cassandra.metrics:*&template=identity"
	MbeanQuery     = "/mbean?objectname="
	QuerySuffix    = "&template=identity"
	JavaStringType = "java.lang.String"
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

//XMLAttributes represents list of Attribute elements
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
	Root   *node
}

// NewCassClient returns a new instance of CassClient
func NewCassClient(url, host string) *CassClient {
	return &CassClient{
		client: NewHTTPClient(url, "", DefaultTimeout),
		host:   host,
		Root:   &node{Name: Root, Children: map[string]*node{}},
	}
}

// NewEmptyCassClient returns an empty instance of CassClient
func NewEmptyCassClient() *CassClient {
	return &CassClient{}
}

// getMetricType returns all available metric types. It reads from
// CassandraMetricType.json file.It builds metric list only when the file does not exist or it's empty.
func (cc *CassClient) getMetricType(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	types, err := readMetricType()
	if err != nil {
		return cc.buildMetricType(cfg)
	}
	return types, nil
}

func (cc *CassClient) BuildMetricType(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	return cc.buildMetricType(cfg)
}

// buildMetricType builds all metric types and write them into
// CassandraMetricType.json file.
func (cc *CassClient) buildMetricType(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	cc, err := initClient(cfg)
	if err != nil {
		return nil, err
	}

	resp, err := cc.client.httpClient.Get(cc.client.GetUrl() + MetricQuery)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	mbeans, err := readObjectname(resp.Body)
	if err != nil {
		return nil, err
	}

	nspace := map[string]plugin.MetricType{}
	fmt.Println(len(mbeans))
	a := 0
	for _, mbean := range mbeans {
		// mbean.Objectname represents each callable measurement
		ns, _ := cc.getElementTypes(mbean.Objectname)
		for _, n := range ns {
			nspace[n.Namespace().String()] = n
		}
		fmt.Println(a)
		a++
	}

	mtsType := []plugin.MetricType{}
	for _, v := range nspace {
		mtsType = append(mtsType, v)
	}

	writeMetricTypes(mtsType)
	return mtsType, nil
}

func (cc *CassClient) BuidMetricAPI() error {
	return cc.buildMetricAPI()
}

// buildMetricAPI builds the base searchable tree and write it
// into CassandraMetricAPI.json file.
func (cc *CassClient) buildMetricAPI() error {
	resp, err := cc.client.httpClient.Get(cc.client.GetUrl() + MetricQuery)
	if err != nil {
		return err
	}

	mbeans, err := readObjectname(resp.Body)
	if err != nil {
		return err
	}

	for _, mbean := range mbeans {
		nodes := makeLitteralNamespace(mbean.Objectname, "")
		cc.Root.Add(nodes, 0, mbean.Objectname)
	}
	writeMetricAPIs(cc.Root)
	return nil
}

// getElementTypes returns specific XML element namespace along with its unit
func (cc *CassClient) getElementTypes(url string) ([]plugin.MetricType, error) {
	resp, err := cc.client.httpClient.Get(cc.client.GetUrl() + MbeanQuery + url + QuerySuffix)
	if err != nil {
		cassLog.WithFields(log.Fields{
			"_block": "getTypes",
			"error":  err,
		}).Error(ReadDocErr)
		return nil, err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil || string(contents) == EmptyRespErr {
		cassLog.WithFields(log.Fields{
			"_block": "getTypes",
			"error":  err,
		}).Error(QueryDocErr)
		return nil, errors.New(QueryDocErr)
	}

	attrs, _ := readXMLAttrbutes(contents)
	ns := []plugin.MetricType{}
	for _, attr := range attrs {
		if attr.Type != JavaStringType {
			ns = append(ns, plugin.MetricType{
				Namespace_: makeDynamicNamespace(cc.host, url, attr.Name),
				Unit_:      attr.Type,
			})
		}
	}
	return ns, nil
}

// getQueryURL returns the MX4J URL from the giving metric namespace
func (cc *CassClient) getQueryURL(ns []string) (string, error) {
	if len(ns) == 0 || len(ns) < 6 {
		cassLog.WithFields(log.Fields{
			"_block": "getUrl",
			"error":  errors.New(InvalidNamespaceErr),
		}).Error(EmptyNamespaceErr)
		return "", errors.New(InvalidNamespaceErr)
	} else if ns[0] != "intel" && ns[1] != "cassandra" && ns[2] != "node" && ns[4] != "org.apache.cassandra.metrics" {
		cassLog.WithFields(log.Fields{
			"_block": "getUrl",
			"error":  errors.New(InvalidNamespaceErr + strings.Join(ns, "/")),
		}).Error("To be collected metric namespace is invalid")
		return "", errors.New(InvalidNamespaceErr)
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

func readObjectname(reader io.Reader) ([]XMLMBean, error) {
	var xmlServer XMLServer
	err := xml.NewDecoder(reader).Decode(&xmlServer)
	if err != nil {
		return nil, err
	}
	return xmlServer.Domain.MBeans, nil
}

func readXMLAttrbutes(content []byte) ([]XMLAttribute, error) {
	var xmlAttributes XMLAttributes
	xml.Unmarshal(content, &xmlAttributes)
	return xmlAttributes.Attributes, nil
}
