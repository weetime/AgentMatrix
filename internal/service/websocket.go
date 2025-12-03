package service

import (
	"net/http"

	"nova/internal/kit"
)

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	kit.GetWebSocket().NodeWebSocketHandler(w, r)
}
