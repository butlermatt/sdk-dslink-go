package main

import (
	"fmt"
	//"github.com/butlermatt/dslink"
	"github.com/butlermatt/dslink/conn"
)

func main() {
	l := conn.NewLink("MyRequester-", conn.IsNotResponder, conn.IsRequester, conn.OnConnected(connected))
	l.Init()

	l.Start()
}

func connected(l *conn.Link) {
	req := l.GetRequester()

	fmt.Println("In Connected")
	n, err := req.GetRemoteNode("/downstream/Example")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Got node: %#v\n", n)
}