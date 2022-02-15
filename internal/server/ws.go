package server

import (
	"encoding/json"
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
		srv.log.Error("ws: error upgrading connection", err)
		return
	}
	wsConnections[c.RemoteAddr().String()] = c
	defer closeWS(c)
	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				srv.log.Error("ws: error reading message", err)
			}
			break
		}

		var data map[string]interface{}
		err = json.Unmarshal(msg, &data)
		if err != nil {
			srv.log.Error("ws: error unmarshalling message", err)
			break
		}

		var msgType string
		switch v := data["type"].(type) {
		case string:
			msgType = v
		default:
			srv.log.Debug("message type data sent was invalid")
		}

		var resp []byte
		switch strings.ToLower(msgType) {
		case "":
		default:
			srv.log.Debug("message type sent was invalid, got:", msgType)
			resp = msg
		}

		err = c.WriteMessage(mt, resp)
		if err != nil {
			srv.log.Error("ws: error writing message", err)
			break
		}
	}
}
