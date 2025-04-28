package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"
	"wisp/src/model"
	"wisp/src/repository"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Client struct {
	Hub      *Hub
	UserID   string
	DeviceID string
	Conn     *websocket.Conn
	Send     chan []byte
	mu       sync.Mutex
}

type Hub struct {
	Clients    map[string]map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *model.Message
	mu         sync.RWMutex
	msgRepo    *repository.MessageRepo
}

func NewHub(msgRepo *repository.MessageRepo) *Hub {
	return &Hub{
		Clients:    make(map[string]map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *model.Message),
		msgRepo:    msgRepo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if _, ok := h.Clients[client.UserID]; !ok {
				h.Clients[client.UserID] = make(map[string]*Client)
			}
			h.Clients[client.UserID][client.DeviceID] = client
			h.mu.Unlock()

			go h.sendPendingMessages(client)

		case client := <-h.Unregister:
			h.mu.Lock()
			if devices, ok := h.Clients[client.UserID]; ok {
				if _, ok := devices[client.DeviceID]; ok {
					delete(devices, client.DeviceID)
					close(client.Send)

					if len(devices) == 0 {
						delete(h.Clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			delivered := h.deliverMessage(message)

			if !delivered {
				h.storePendingMessage(message)
			}
		}
	}
}

func (h *Hub) deliverMessage(message *model.Message) bool {
	h.mu.RLock()
	devices, exists := h.Clients[message.To]
	h.mu.RUnlock()

	if !exists || len(devices) == 0 {
		return false
	}

	msgJSON, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao serializar mensagem")
		return false
	}

	delivered := false
	for _, client := range devices {
		select {
		case client.Send <- msgJSON:
			delivered = true
		default:
		}
	}

	return delivered
}

func (h *Hub) storePendingMessage(message *model.Message) {
	pendingMsg := &model.PendingMessage{
		From:      message.From,
		To:        message.To,
		Payload:   message.Content,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, err := h.msgRepo.Insert(ctx, pendingMsg)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao armazenar mensagem pendente")
	} else {
		log.Debug().Str("id", id.Hex()).Msg("Mensagem armazenada para entrega posterior")
	}
}

func (h *Hub) sendPendingMessages(client *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userId := client.UserID
	pendingMsgs, err := h.msgRepo.GetPendingFor(ctx, userId)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao buscar mensagens pendentes")
		return
	}

	for _, pm := range pendingMsgs {
		msg := &model.Message{
			Type:      "message",
			From:      pm.From,
			To:        pm.To,
			Content:   pm.Payload,
			Timestamp: pm.CreatedAt.Unix(),
			ID:        pm.ID.Hex(),
		}

		msgJSON, err := json.Marshal(msg)
		if err != nil {
			log.Error().Err(err).Msg("Erro ao serializar mensagem pendente")
			continue
		}

		client.Send <- msgJSON
	}
}

func (h *Hub) ProcessMessageAck(ack *model.Ack) {
	if ack.Type != "ack" {
		return
	}

	msgID, err := primitive.ObjectIDFromHex(ack.MessageID)
	if err != nil {
		log.Error().Err(err).Msg("ID de mensagem inválido em ACK")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.msgRepo.Delete(ctx, msgID)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao excluir mensagem pendente após ACK")
	}

	h.notifySender(ack)
}

func (h *Hub) notifySender(ack *model.Ack) {
	h.mu.RLock()
	devices, exists := h.Clients[ack.From]
	h.mu.RUnlock()

	if !exists || len(devices) == 0 {
		return
	}

	ackJSON, err := json.Marshal(ack)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao serializar ACK")
		return
	}

	for _, client := range devices {
		select {
		case client.Send <- ackJSON:
		default:
		}
	}
}

func (c *Client) SendMessage(msg []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Conn != nil {
		c.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Str("userId", c.UserID).Str("deviceId", c.DeviceID).Msg("Erro de conexão WebSocket")
			}
			break
		}

		var data map[string]any
		if err := json.Unmarshal(message, &data); err != nil {
			log.Error().Err(err).Msg("Formato de mensagem inválido")
			continue
		}

		if msgType, ok := data["type"].(string); ok {
			switch msgType {
			case "message":
				var msg model.Message
				if err := json.Unmarshal(message, &msg); err == nil {
					if msg.From != c.UserID {
						log.Warn().Str("claimed", msg.From).Str("actual", c.UserID).Msg("Tentativa de envio com ID falsificado")
						continue
					}
					if msg.Timestamp == 0 {
						msg.Timestamp = time.Now().Unix()
					}
					c.Hub.Broadcast <- &msg
				}
			case "ack":
				var ack model.Ack
				if err := json.Unmarshal(message, &ack); err == nil {
					if ack.To != c.UserID {
						continue
					}
					c.Hub.ProcessMessageAck(&ack)
				}
			}
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for range n {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
