package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
)

func (api *API) GetLogs() http.HandlerFunc {
	type logLine map[string]interface{}

	type response struct {
		Logs []logLine `json:"logs"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := os.OpenFile(api.log.GetFilePath(), os.O_RDONLY, 0666)
		if err != nil {
			api.error(w, r, "Could not open log file.", err, http.StatusInternalServerError)
			return
		}
		defer file.Close()

		logs := make([]logLine, 0)
		data, err := io.ReadAll(file)
		if err != nil {
			api.error(w, r, "Could not read log file.", err, http.StatusInternalServerError)
			return
		}

		for _, line := range strings.Split(string(data), "\n") {
			if line == "" {
				continue
			}

			log := make(logLine)
			err = json.Unmarshal([]byte(line), &log)
			if err != nil {
				api.error(w, r, "Could not unmarshal log file.", err, http.StatusInternalServerError)
				return
			}
			logs = append(logs, log)
		}

		api.respond(w, r, response{Logs: logs}, http.StatusOK)
	}
}
