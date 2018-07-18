package serveks

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"net/http"
)


var HostFile = "examples/dnsmasq.conf.EXAMPLE"
var KsTmplFile = "examples/node.ks.EXAMPLE"

var HMap HostsMap

var isInitialized = false

// There's really no reason to make this a datatype other than that it's convenient to reference fields
// when it gets passed to the template engine
type Host struct {
	IpAddr     string
	HostName   string
	Router     string
	NameServer string
	macAddr    string
}

type HostsMap struct {
	file  string
	Hosts map[string]Host
}

// NetHostsMap takes a path name to a dnsmasq.conf file and returns a HostsMap of the
// hosts contained therein keyed by IP (as the http server fills in the template based
// on the client's IP
func NewHostsMap() (HostsMap, error) {
	hm := HostsMap{file: HostFile}
	hm.Hosts = make(map[string]Host)

	file, err := os.Open(HostFile)
	if err != nil {
		log.Warn("Unable to open %s (%s)", HostFile, err)
		return hm, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var defaultRouter, nameServer string

	// Do a first scan for the nameserver and router.
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "dhcp-option=option:router") {
			defaultRouter = strings.Split(line, ",")[1]
		} else if strings.HasPrefix(line, "dhcp-option=option:dns-server") {
			nameServer = strings.Split(line, ",")[1]
		}
	}

	// reset the position and create a new scanner to go through a second time for hosts
	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)

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

			hm.Hosts[fields[1]] = Host{IpAddr: fields[1], HostName: fields[2], Router: defaultRouter, NameServer: nameServer}
		}
	}

	return hm, nil
}

func TemplateInit() error {
	var err error
	HMap, err = NewHostsMap()
	if err != nil {
		return err
	}

	isInitialized = true
	return nil
}

// RenderKsFile generates the text of the kickstart file from the template located in
// ksTmplFile with the Host data substituted in the template file.
func RenderKsFile(renderedTmpl *bytes.Buffer, host Host) error {
	log.Debugln("starting template render")
	tmpl, err := template.ParseFiles(KsTmplFile)
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

func generateKsConfig(remIpAddr string) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	host, ok := HMap.Hosts[remIpAddr]
	if !ok {
		log.Errorf("host %s not found in dnsmasq config", remIpAddr)
		return &buf, fmt.Errorf("host %s not found in dnsmasq.conf", remIpAddr)
	}

	log.Warnf("host: <<%s>>\n", host)
	err := RenderKsFile(&buf, host)
	if err != nil {
		return &buf, err
	}

	return &buf, nil
}

func getIPFromRemoteAddr(s string) string {
	return strings.Split(s, ":")[0]
}

func HttpKSHandler(w http.ResponseWriter, r *http.Request) {
	renderedKsTemplate, err := generateKsConfig(getIPFromRemoteAddr(r.RemoteAddr))
	if err != nil {
		errstr := fmt.Sprintf("could not render kickstart template: %s", err)
		http.Error(w, errstr, 500)
		return
	}

	fmt.Fprint(w, renderedKsTemplate.String())
}
