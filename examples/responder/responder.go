package main

import (
	"fmt"
	"time"
	"github.com/butlermatt/dslink"
	"github.com/butlermatt/dslink/conn"
	"github.com/butlermatt/dslink/nodes"
)

func main() {
	l := conn.NewLink("MyTest-", conn.OnConnected(connected))
	l.Init()

	prov := l.GetProvider()
	root := prov.GetRoot()

	n := nodes.NewNode("Test", prov)
	n.AddAction(Tester, []dslink.Params{
		{
			dslink.ParamName: "test",
			dslink.ParamType: dslink.ValueBool,
			dslink.ParamDef: true,
		},
	}, []dslink.Column{
		{
			Name: "success",
			Type: dslink.ValueBool,
		},
		{
			Name: "message",
			Type: dslink.ValueString,
			Default: "Hello",
		},
	}, dslink.ResultValues)
	root.AddChild(n)

	n = nodes.NewNode("Set_Me", prov)
	n.SetConfig(dslink.ConfigName, "Set Me")
	n.SetType(dslink.ValueString)
	n.UpdateValue("Hello World")
	n.EnableSet(dslink.PermWrite, func(n dslink.Node, v interface{}) bool {
		dslink.Log.Printf("Going to set value: %v", v)
		return true
	})
	root.AddChild(n)

	n = nodes.NewNode("TestValue", prov)
	n.SetType(dslink.ValueString)
	n.UpdateValue("Hello There!")
	root.AddChild(n)

	l.Start()
}

func Tester(params map[string]interface{}, ret chan<-[]interface{}) {
	dslink.Log.Println("I'm in the invoke!")

	dslink.Log.Printf("Got params: %v\n", params)
	ret <- []interface{}{true, "Success!"}
	close(ret)
}

func connected(l *conn.Link) {
	dslink.Log.Println("I'm connected!")

	p := l.GetProvider()
	nd := p.GetNode("/TestValue")
	if nd == nil {
		dslink.Log.Println("Node was missing!?")
		return
	}
	node, ok := nd.(dslink.Valued)
	if !ok {
		dslink.Log.Println("Couldn't make it a valued")
		return
	}

	t := time.NewTicker(time.Second * 2)

	dslink.Log.Printf("I'm in here now?\n")
	j := 0
	var nn *nodes.LocalNode
	for i := range t.C {
		j++
		dslink.Log.Printf("i Is: %v\n", i)
		node.UpdateValue(fmt.Sprintf("I'm now %d", j))
		if j >= 10 {
			t.Stop()
			p.RemoveNode("/TestValue/Matthew")
		}
		if j == 5 {
			nn = nodes.NewNode("Matthew", p)
			nn.SetType(dslink.ValueString)
			nn.UpdateValue("Matt")
			nd.AddChild(nn)
		}
	}

}