package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
} // use default options
func (srv *Server) ws(w http.ResponseWriter, req *http.Request) {
	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	wsConnections[c.RemoteAddr().String()] = c
	defer closeWS(c)
	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		var data map[string]interface{}
		err = json.Unmarshal(msg, &data)
		if err != nil {
			log.Println(err)
			break
		}

		var msgType string
		switch v := data["type"].(type) {
		case string:
			msgType = v
		default:
			log.Println("message type data sent was invalid")
		}

		var resp []byte
		switch strings.ToLower(msgType) {
		// case "getbackups":
		// msg, err := getBackups()
		// if err != nil {
		// 	log.Println(err)
		// }

		// 	resp = msg
		// case "getrepos":
		// 	msg, err := getRepos()
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "getagents":
		// 	msg, err := getAgents()
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "newbackup":
		// 	msg, err := srv.newBackup(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "newrepo":
		// 	msg, err := newRepo(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "newagent":
		// 	msg, err := newAgent(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "updatebackup":
		// 	msg, err := srv.updateBackup(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "updaterepo":
		// 	msg, err := updateRepo(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "updateagent":
		// 	msg, err := updateAgent(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "deletebackup":
		// 	msg, err := deleteBackup(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "deleterepo":
		// 	msg, err := deleteRepo(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "deleteagent":
		// 	msg, err := deleteAgent(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "getsubscribers":
		// 	msg, err := getSubscribers()
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "updatesubscribers":
		// 	msg, err := updateSubscribers(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "getsnapshots":
		// 	msg, err := getSnapshots(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "restoresnapshot":
		// 	msg, err := srv.restoreSnapshot(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "deletesnapshot":
		// 	msg, err := deleteSnapshot(data)
		// 	if err != nil {
		// 		log.Println(err)
		// 	}

		// 	resp = msg
		// case "getjobs":
		// 	msg, err := srv.getJobs()
		// 	if err != nil {
		// 		log.Println(err)
		// }
		default:
			log.Println("message type sent was invalid", msgType)

			resp = msg
		}

		err = c.WriteMessage(mt, resp)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
