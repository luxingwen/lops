package quicnet

import (
	"crypto/tls"
	"io"
	"encoding/binary"
	"context"


	quic "github.com/quic-go/quic-go"
)

type Client struct {
	session quic.Connection 
	stream  quic.Stream
}

func NewClient(serverAddr string, tlsCfg *tls.Config, quicCfg *quic.Config) (*Client, error) {
	session, err := quic.DialAddr(serverAddr, tlsCfg, quicCfg)
	if err != nil {
		return nil, err
	}

	

	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{
		session: session,
		stream:  stream,
	}, nil
}

func (c *Client) Close() {
	c.stream.Close()
	c.session.CloseWithError(0, "")
}

func (c *Client) Write(data []byte) (int, error) {
	return writePacket(c.stream, data)
}

func (c *Client) Read() ([]byte, error) {
	return readPacket(c.stream)
}

func writePacket(stream quic.Stream, data []byte) (int, error) {
	packetLength := uint32(len(data))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, packetLength)

	n1, err := stream.Write(lengthBytes)
	if err != nil {
		return n1, err
	}
	n2, err := stream.Write(data)
	return n1 + n2, err
}

func readPacket(stream quic.Stream) ([]byte, error) {
	lengthBytes := make([]byte, 4)
	_, err := io.ReadFull(stream, lengthBytes)
	if err != nil {
		return nil, err
	}

	packetLength := binary.BigEndian.Uint32(lengthBytes)
	data := make([]byte, packetLength)
	_, err = io.ReadFull(stream, data)
	return data, err
}
