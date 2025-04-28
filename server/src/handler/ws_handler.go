package handler

import (
	"net/http"
	"wisp/src/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implementar verificação de origem
		log.Debug().Msg("Verificando origem do WebSocket")
		return true
	},
}

type WSHandler struct {
	hub *ws.Hub
}

func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

func (h *WSHandler) HandleConnection(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não autenticado"})
		return
	}
	userId := userIdVal.(string)

	deviceId := c.Query("deviceId")
	if deviceId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do dispositivo é obrigatório"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Falha ao atualizar para WebSocket")
		return
	}

	client := &ws.Client{
		Hub:      h.hub,
		UserID:   userId,
		DeviceID: deviceId,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
