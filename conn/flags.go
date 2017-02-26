package conn

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"github.com/butlermatt/dslink/log"
)

const brokerDefault = "http://127.0.0.1:8080/conn"

var (
	brokerAddr string
	linkName   string
	home       string
	token      string
	basePath   string
	logFile    string
	logL       string
	help       bool
)

func init() {
	const (
		helpUsage   = "Display this help message"
		brokerUsage = "Broker `URL`"
		homeUsage   = "Connect to user `home` space"
		nameUsage   = "Link `Name`"
		tokenUsage  = "Authorization `Token`"
		baseUsage   = "Root `path` of the DSLink"
		logfUsage   = "Output file for logger"
		loglUsage   = "Enable the default logger"
	)

	flag.Usage = func() {
		ex := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", ex)
		flag.PrintDefaults()
	}

	flag.BoolVar(&help, "h", false, helpUsage)
	flag.BoolVar(&help, "help", false, helpUsage)
	flag.StringVar(&logL, "log", "", loglUsage)
	flag.StringVar(&brokerAddr, "broker", brokerDefault, brokerUsage)
	flag.StringVar(&brokerAddr, "b", brokerDefault, brokerUsage)
	flag.StringVar(&linkName, "name", "", nameUsage)
	flag.StringVar(&linkName, "n", "", nameUsage)
	flag.StringVar(&home, "home", "", homeUsage)
	flag.StringVar(&token, "token", "", tokenUsage)
	flag.StringVar(&basePath, "basepath", "", baseUsage)
	flag.StringVar(&logFile, "logfile", "", logfUsage)
}

func parseFlags(c *config) {
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if brokerAddr != brokerDefault || c.broker == "" {
		c.broker = brokerAddr
	}
	if linkName != "" {
		c.name = linkName
	}

	c.home = home
	c.token = token
	c.rootPath = basePath
	c.logFile = logFile
	if logL == "" {
		c.logLevel = log.DebugLevel
	} else {
		ll, err := log.ToLevel(logL)
		c.logLevel = ll
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unknown log level: %q Logging is disabled\n", logL)
		}
	}
}
