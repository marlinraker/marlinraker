package executors

import (
	log "github.com/sirupsen/logrus"
	"marlinraker/src/files"
	"marlinraker/src/marlinraker/connections"
	"marlinraker/src/util"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
)

type ServerLogsRolloverResult struct {
	RolledOver []string          `json:"rolled_over"`
	Failed     map[string]string `json:"failed"`
}

func ServerLogsRollover(*connections.Connection, *http.Request, Params) (any, error) {

	statusFile := filepath.Join(files.DataDir, "logrotate.status")
	if out, err := exec.Command("logrotate", "-s", statusFile, "-f", "/etc/logrotate.d/marlinraker").CombinedOutput(); err != nil {
		result := strings.TrimSpace(string(out))
		log.Error(result)
		util.LogError(err)

		return ServerLogsRolloverResult{
			RolledOver: []string{},
			Failed: map[string]string{
				"marlinraker": result,
			},
		}, nil
	}

	return ServerLogsRolloverResult{
		RolledOver: []string{"marlinraker"},
		Failed:     map[string]string{},
	}, nil
}
