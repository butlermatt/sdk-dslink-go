package main

import (
	"fmt"
	"github.com/butlermatt/dslink"
	"github.com/butlermatt/dslink/link"
)

func main() {
	l := link.NewLink("MyRequester-", link.IsNotResponder, link.IsRequester)
	l.Init()

	req := l.GetRequester()

}

func connected(l link.Link) {
	
}