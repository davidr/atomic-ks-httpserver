package serveks

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
)

// There's really no reason to make this a datatype other than that it's convenient to reference fields
// when it gets passed to the template engine
type Host struct {
	IpAddr   string
	HostName string
	macAddr  string
}

type HostsMap struct {
	file  string
	Hosts map[string]Host
}

// NetHostsMap takes a path name to a dnsmasq.conf file and returns a HostsMap of the
// hosts contained therein keyed by IP (as the http server fills in the template based
// on the client's IP
func NewHostsMap(hostFile string) (HostsMap, error) {
	hm := HostsMap{file: hostFile}
	hm.Hosts = make(map[string]Host)

	file, err := os.Open(hostFile)
	if err != nil {
		log.Warn("Unable to open %s (%s)", hostFile, err)
		return hm, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "dhcp-host=") {
			fields := strings.Split(line, ",")
			if len(fields) != 4 {
				// For some reason, we don't have four fields here, which shouldn't be possible.
				// Either way, just continue on so we don't get an index error when we try to assign
				// from the slice
				log.Warnf("line did not have four fields, skipping: %s", line)
				continue
			}

			hm.Hosts[fields[1]] = Host{IpAddr: fields[1], HostName: fields[2]}
		}
	}

	return hm, nil
}

// RenderKsFile generates the text of the kickstart file from the template located in
// ksTmplFile with the Host data substituted in the template file.
func RenderKsFile(renderedTmpl *bytes.Buffer, host Host, ksTmplFile string) error {
	log.Debugln("starting template render")
	fmt.Println(host)
	tmpl, err := template.ParseFiles("examples/node.ks.EXAMPLE")
	if err != nil {
		log.Fatalf("could not create template: %s", err)
	}

	err = tmpl.Execute(renderedTmpl, host)
	if err != nil {
		log.Warnf("could not execute template: %s", err)
		return nil
	}

	return nil
}
