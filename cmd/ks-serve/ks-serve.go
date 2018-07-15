package main

import (
	"bytes"
	"fmt"

	"github.com/davidr/atomic-ks-httpserver/internal/serveks"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.Debugln("starting program")

	hostFile := "examples/dnsmasq.conf.EXAMPLE"
	ksTmplFile := "examples/node.ks.EXAMPLE"

	hostsMap, _ := serveks.NewHostsMap(hostFile)

	var t bytes.Buffer
	serveks.RenderKsFile(&t, hostsMap.Hosts["10.0.0.101"], ksTmplFile)
	fmt.Println(t.String())

}
