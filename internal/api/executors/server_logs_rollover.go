package executors

import (
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/marlinraker/marlinraker/internal/files"
	"github.com/marlinraker/marlinraker/internal/marlinraker/connections"
)

type ServerLogsRolloverResult struct {
	RolledOver []string          `json:"rolled_over"`
	Failed     map[string]string `json:"failed"`
}

func ServerLogsRollover(*connections.Connection, *http.Request, Params) (any, error) {

	statusFile := filepath.Join(files.DataDir, "logrotate.status")
	if out, err := exec.Command("logrotate", "-s", statusFile, "-f", "/etc/logrotate.d/marlinraker").CombinedOutput(); err != nil {
		result := strings.TrimSpace(string(out))
		log.Errorf("Failed to rollover logs: %v", err)
		log.Errorln(result)

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
