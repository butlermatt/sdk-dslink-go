package main

import (
	"fmt"
	//"github.com/butlermatt/dslink"
	"github.com/butlermatt/dslink/conn"
	"github.com/butlermatt/dslink/nodes"
)

func main() {
	l := conn.NewLink("MyRequester-", conn.IsRequester, conn.OnConnected(connected))
	l.Init()

	l.Start()
}

func connected(l *conn.Link) {
	req := l.GetRequester()

	fmt.Println("In Connected")
	go test("/downstream/Example", req)
}

func test(path string, req *nodes.Requester) {
	n, err := req.GetRemoteNode(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	printNode(n)
	for _, k := range n.Children() {
		go test(k.Path(), req)
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