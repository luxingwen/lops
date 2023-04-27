package quicnet

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"

	quic "github.com/quic-go/quic-go"
)

type Server struct {
	listener       quic.Listener
	cm             *ClientManager
	messageHandler *MessageHandler
}

func NewServer(addr string, tlsCfg *tls.Config, quicCfg *quic.Config) (*Server, error) {
	listener, err := quic.ListenAddr(addr, tlsCfg, quicCfg)
	if err != nil {
		return nil, err
	}

	return &Server{
		listener: listener,
		cm:       NewClientManager(),
	}, nil
}

func (s *Server) Close() {
	s.listener.Close()
}

func (s *Server) Accept() (*Client, error) {
	session, err := s.listener.Accept(context.Background())
	if err != nil {
		return nil, err
	}

	stream, err := session.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{
		session: session,
		stream:  stream,
	}, nil
}

func (s *Server) HandleHeartbeat(client *Client, msg *Message) {
	var heartbeat HeartbeatData
	err := json.Unmarshal(msg.Data, &heartbeat)
	if err != nil {
		log.Printf("Failed to unmarshal heartbeat message: %s", err)
		return
	}

	client.IP = heartbeat.IP
	client.MachineID = heartbeat.MachineID
	client.Hostname = heartbeat.Hostname
	s.cm.AddClient(client)
}

func (s *Server) Start() {
	for {
		client, err := s.Accept()
		if err != nil {
			log.Printf("Failed to accept client: %s", err)
			continue
		}

		go func() {
			for {
				msg, err := client.Read()
				if err != nil {
					log.Printf("Failed to read message from client: %s", err)
					return
				}
				_ = msg
				//s.messageHandler.SubmitMessage(msg)
			}
		}()
	}
}
