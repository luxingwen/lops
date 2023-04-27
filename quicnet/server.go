package quicnet

import (
	"context"
	"crypto/tls"

	quic "github.com/quic-go/quic-go"
)

type Server struct {
	listener quic.Listener
}

func NewServer(addr string, tlsCfg *tls.Config, quicCfg *quic.Config) (*Server, error) {
	listener, err := quic.ListenAddr(addr, tlsCfg, quicCfg)
	if err != nil {
		return nil, err
	}

	return &Server{
		listener: listener,
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
