package main

import (
	"flag"
	"net/http"

	"github.com/davidr/atomic-ks-httpserver/internal/serveks"
	log "github.com/sirupsen/logrus"
	"fmt"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.Debugln("parsing flags")

	flag.StringVar(&serveks.HostFile, "config", "/etc/dnsmasq.conf", "dnsmasq.conf for host definitions")
	flag.StringVar(&serveks.KsTmplFile, "template", "/srv/kickstart/ks.tmpl", "Kickstart template file")
	listenIp := flag.String("interface", "127.0.0.1", "IP of interface on which to listen")
	portNumber := flag.Int("port", 8080, "port number")
	flag.Parse()

	log.Debugln("starting program")

	err := serveks.TemplateInit()
	if err != nil {
		log.Fatal("Could not initialize template engine: %s", err)
	}
	http.HandleFunc("/kickstart", serveks.HttpKSHandler)

	listenIfaceString := fmt.Sprintf("%s:%d", *listenIp, *portNumber)
	log.Fatal(http.ListenAndServe(listenIfaceString, nil))
}
