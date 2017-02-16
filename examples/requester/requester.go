package main

import (
	"fmt"
	//"github.com/butlermatt/dslink"
	"github.com/butlermatt/dslink/link"
)

func main() {
	l := link.NewLink("MyRequester-", link.IsRequester, link.OnConnected(connected))
	l.Init()

	l.Start()
}

func connected(l link.Link) {
	req := l.GetRequester()

	fmt.Println("In Connected")
	n, err := req.GetRemoteNode("/downstream/Example")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Got node: %v\n", n)
}