package main

import (
	"fmt"
	//"github.com/butlermatt/dslink"
	"github.com/butlermatt/dslink/conn"
	"github.com/butlermatt/dslink/nodes"
	"time"
)

func main() {
	l := conn.NewLink("MyRequester-", conn.IsRequester, conn.OnConnected(connected))
	l.Init()

	l.Start()
}

func connected(l *conn.Link) {
	req := l.GetRequester()

	fmt.Println("In Connected")
	testGetNode("/downstream/Example", req)
	testListNode("/downstream/Example", req)

	fmt.Println("Done all!")
	l.Stop()
}

func testGetNode(path string, req *nodes.Requester) {
	n, err := req.GetRemoteNode(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printNode(n)
		for _, k := range n.Children() {
			testGetNode(k.Path(), req)
		}
	}
}

func printNode(n *nodes.RemoteNode) {
	fmt.Println("Got Node")
	fmt.Printf("\tPath: %s\n", n.Path())
	fmt.Printf("\tName: %s\n", n.Name())
	fmt.Printf("\tType: %s\n", n.Type())
	fmt.Println("\tAttributes:")
	for k, v := range n.Attributes() {
		fmt.Printf("\t\t%q: %v\n", k, v)
	}
	fmt.Println("\tConfigs:")
	for k, v := range n.Configs() {
		fmt.Printf("\t\t%q: %v\n", k, v)
	}
	fmt.Println("\tChildren:")
	for k := range n.Children() {
		fmt.Printf("\t\t%q\n", k)
	}
}

func testListNode(path string, req *nodes.Requester) {
	tl := time.After(time.Second * 5)
	uChan := make(chan []interface{})
	rid := req.List(path, uChan)
	for {
		select {
		case up, ok := <-uChan:
			if !ok {
				uChan = nil
			} else {
				fmt.Printf("Update contains %d items\n", len(up))
				for i, u := range up {
					fmt.Printf("\t%d: %v\n", i, u)
				}
			}
		case <-tl:
			req.CloseRequest(rid)
		}
		if uChan == nil {
			break
		}
	}
}
