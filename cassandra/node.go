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
	"io/ioutil"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	// Wildcard the * string
	Wildcard = "*"
	// Pipe the | string
	Pipe = "|"
	// Slash the slash symbol string
	Slash = "/"
)

type node struct {
	Name     string
	Children map[string]*node
	Target   *nodeTarget
	Data     *nodeData
}

// nodeTarget defines the callable host and the endpoint.
// Only leaf nodes have this property.
type nodeTarget struct {
	// URI the target URI such as  org.apache.cassandra.metrics&type=CQL,name=PreparedStatementsCount
	URI    string
	Loaded bool
}

// nodeData defines the key and value pair of the node data.
// Only leaf nodes have this property.
type nodeData struct {
	Path string
	Data interface{}
}

// newNode returns a new instance with the node name
// and the child node initialized.
func newNode(name string) *node {
	return &node{Name: name, Children: map[string]*node{}}
}

// newNodeTarget returns a new instance with the invocable host
// and the endpoint uri defined.
func newNodeTarget(uri string) *nodeTarget {
	return &nodeTarget{URI: uri}
}

// newNodeData returns the data point's path and the value.
func newNodeData(path string, data interface{}) *nodeData {
	return &nodeData{Path: path, Data: data}
}

// Add adds a path into the tree. Each entry in names is a part of a path between two slashes.
func (n *node) Add(names []string, index int, uri string) {
	if index == len(names) {
		n.Target = newNodeTarget(uri)
		return
	}

	c, ok := n.Children[names[index]]
	// add c if it doesn't exist
	if !ok {
		c = newNode(names[index])
		n.Children[names[index]] = c
	}
	c.Add(names, index+1, uri)
}

// Get returns results that match the specified path which may contain wildcards and |'s which serve as OR booleans.
// For example /a/b/*/d will return all nodes under "b" which themselves have a child "d".
// Another example is /a/b/c|d/e which returns /a/b/c/e and /a/b/d/e.
func (n *node) Get(url string, names []string, index int, results *[]nodeData) (err error) {
	// we've reached the end of the path, so add to the results if this node has anything to add.
	if index == len(names) {
		if n.Data != nil {
			*results = append(*results, *n.Data)
		}
		return
	}

	// Go through each substring if a pipe exists inside a string
	tokens := strings.Split(names[index], Pipe)
	if len(tokens) > 1 {
		for _, token := range tokens {
			err = n.getSpecific(url, token, names, index, results)
		}
	} else {
		err = n.getSpecific(url, names[index], names, index, results)
	}
	return nil
}

// getSpecific traverses through the node and finds the matching data set.
// If requested, the XML will be loaded into child nodes as it is needed. Once loaded it serves as a cache so the same url
// won't be reloaded over and over if multiple values are required from the same page.
// The results will be empty if no matches are found.
func (n *node) getSpecific(url, name string, names []string, index int, results *[]nodeData) (err error) {
	if len(n.Children) == 0 && n.Target != nil {
		// load XML if we're in a leaf node and there is a url to load from.
		err = n.loadElements(url)
	} else if n.Target != nil {
		// load XML if it's an end node of a callable target
		// and the searching name does not exist in its children
		_, ok := n.Children[name]
		if !ok {
			err = n.loadElements(url)
		}
	}

	if name == Wildcard {
		// traverse all children to find matches if it is *
		for _, child := range n.Children {
			err = child.Get(url, names, index+1, results)
		}
	} else {
		child, ok := n.Children[name]
		if ok {
			err = child.Get(url, names, index+1, results)
		}
	}
	return nil
}

// addXMLAttibutes adds XML attributes into the tree
func (n *node) addXMLAttibutes(ns string, attrs []XMLAttribute) {
	for _, attr := range attrs {
		switch {
		case attr.Type == JavaCompositeType:
			nc, ok := n.Children[attr.Name]
			if !ok {
				nc = newNode(attr.Name)
				//nc.Data = newNodeData(ns+Slash+attr.Name, attr.Value)
				n.Children[attr.Name] = nc
				nc.addCompositeElements(ns+Slash+attr.Name, attr.Value)
			}

		default:
			nc, ok := n.Children[attr.Name]
			if !ok && checkFloatValue(attr.Value) {
				nc = newNode(attr.Name)
				nc.Data = newNodeData(ns+Slash+attr.Name, attr.Value)
				n.Children[attr.Name] = nc
			}
		}
	}
}

func (n *node) addCompositeElements(ns string, coposite string) {
	items := getCompositeItems(coposite)
	for k, v := range items {
		nc := newNode(k)
		nc.Data = newNodeData(ns + Slash + k, v)
		n.Children[k] = nc
	}
}

// loadElements loads the XML hasn't been loaded into the tree yet, load it and add it to the tree.
func (n *node) loadElements(url string) error {
	if n.Target.Loaded {
		return nil
	}
	resp, err := getResp(url, n.Target.URI)
	if err != nil {
		cassLog.WithFields(log.Fields{
			"_block": "loadElements",
			"error":  err,
		}).Error(ReadDocErr)
		return err
	}
	ns := makeLitteralNamespace(n.Target.URI, "")
	n.addXMLAttibutes(strings.Join(ns, "/"), resp)
	n.Target.Loaded = true
	return nil
}

func getResp(u, uri string) ([]XMLAttribute, error) {
	client := NewHTTPClient(u, "", DefaultTimeout)
	resp, err := client.httpClient.Get(u + MbeanQuery + url.QueryEscape(uri) + QuerySuffix)
	if err != nil {
		cassLog.WithFields(log.Fields{
			"_block": "getResp",
			"error":  err,
		}).Error(ReadDocErr)
		return nil, err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil || string(contents) == EmptyRespErr {
		cassLog.WithFields(log.Fields{
			"_block": "getResp",
			"error":  err,
		}).Error(QueryDocErr)
		return nil, errors.New(QueryDocErr)
	}

	return readXMLAttrbutes(contents)
}

// Print prints out the tree to the specified depth.
func (n *node) Print(depth int) {
	for i := 0; i < depth; i++ {
		fmt.Print(" ")
	}
	fmt.Println(n.Name, " : data=", n.Data)

	for _, v := range n.Children {
		v.Print(depth + 1)
	}
}
