package quicnet

import (
	"log"
)

type HandlerFunc func(request *Message, client *Client) error

type MessageHandler struct {
	handlers map[string]HandlerFunc
	in       chan *Message
}

func NewMessageHandler(bufferSize int) *MessageHandler {
	return &MessageHandler{
		handlers: make(map[string]HandlerFunc),
		in:       make(chan *Message, bufferSize),
	}
}

func (h *MessageHandler) RegisterHandler(messageType string, handler HandlerFunc) {
	h.handlers[messageType] = handler
}

func (h *MessageHandler) HandleMessages(client *Client, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go func() {
			for msg := range h.in {
				if handler, ok := h.handlers[msg.Type]; ok {
					err := handler(msg, client)
					if err != nil {
						log.Printf("Error handling message: %s", err)
					}
				} else {
					log.Printf("No handler registered for message type: %s", msg.Type)
				}
			}
		}()
	}
}

func (h *MessageHandler) SubmitMessage(msg *Message) {
	h.in <- msg
}
