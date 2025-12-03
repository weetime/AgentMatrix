package service

import (
	"net/http"

	"github.com/weetime/agent-matrix/internal/kit"
)

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	kit.GetWebSocket().NodeWebSocketHandler(w, r)
}
