// Copyright 2018 gopcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/wmnsk/gopcua"
	uad "github.com/wmnsk/gopcua/datatypes"
	uid "github.com/wmnsk/gopcua/id"
)

func join(a, b string) string {
	if a == "" {
		return b
	}
	return a + "." + b
}

func browse(n *gopcua.Node, path string, level int) ([]string, error) {
	if level > 10 {
		return nil, nil
	}
	// nodeClass, err := n.NodeClass()
	// if err != nil {
	// 	return nil, err
	// }
	browseName, err := n.BrowseName()
	if err != nil {
		return nil, err
	}
	path = join(path, browseName.Name)

	typeDefs := uad.NewTwoByteNodeID(uid.HasTypeDefinition)
	refs, err := n.References(typeDefs)
	if err != nil {
		return nil, err
	}
	// todo(fs): example still incomplete
	log.Printf("refs: %#v err: %v", refs, err)
	return nil, nil
}

func main() {
	endpoint := flag.String("endpoint", "opc.tcp://localhost:4840", "OPC UA Endpoint URL")
	flag.Parse()

	c := gopcua.NewClient(*endpoint, nil)
	if err := c.Open(); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	root := c.Node(uad.NewStringNodeID(1, "Root"))

	nodeList, err := browse(root, "", 0)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range nodeList {
		fmt.Println(s)
	}
}
