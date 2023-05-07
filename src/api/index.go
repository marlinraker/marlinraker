package api

import (
	_ "embed"
	"html/template"
	"marlinraker/src/constants"
	"marlinraker/src/marlinraker"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/system_info"
	"net/http"
	"os"
)

var (
	//go:embed index.html
	indexHtml     string
	indexTemplate = template.Must(template.New("index.html").Parse(indexHtml))
)

type templateData struct {
	Hostname       string
	RequestIP      string
	Version        string
	WebsocketCount int
	State          marlinraker.KlippyState
	Arch           string
	Os             string
	CPU            string
}

func handleIndex(writer http.ResponseWriter, request *http.Request) error {

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	info, err := system_info.GetSystemInfo()
	if err != nil {
		return err
	}

	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(200)

	return indexTemplate.Execute(writer, templateData{
		Hostname:       hostname,
		RequestIP:      getAddress(request),
		Version:        constants.Version,
		WebsocketCount: len(connections.GetConnections()),
		State:          marlinraker.State,
		Arch:           info.CpuInfo.Processor,
		Os:             info.Distribution.Name,
		CPU:            info.CpuInfo.CpuDesc,
	})
}

func getAddress(request *http.Request) string {
	address := request.Header.Get("X-Real-Ip")
	if address == "" {
		address = request.Header.Get("X-Forwarded-For")
	}
	if address == "" {
		address = request.RemoteAddr
	}
	return address
}
