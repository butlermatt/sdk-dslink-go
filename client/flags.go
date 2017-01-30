package client

import "flag"

var brokerAddr string
var linkName string
var home string
var token string
const brokerDefault = "http://127.0.0.1:8080/conn"

func init() {
	const (
		brokerUsage = "Broker URL"
		homeUsage = "Connect to user home space"
		nameUsage = "Link Name"
		tokenUsage = "Token"
	)
	flag.StringVar(&brokerAddr, "broker", brokerDefault, brokerUsage)
	flag.StringVar(&brokerAddr, "b", brokerDefault, brokerUsage)
	flag.StringVar(&linkName, "name", "", nameUsage)
	flag.StringVar(&linkName, "n", "", nameUsage)
	flag.StringVar(&home, "home", "", homeUsage)
	flag.StringVar(&token, "Token", "", tokenUsage)
}

func parseFlags(c *Config) {
	flag.Parse()

	if brokerAddr != brokerDefault || c.broker == "" {
		c.broker = brokerAddr
	}
	if linkName != "" {
		c.name = linkName
	}

	c.home = home
	c.token = token
}