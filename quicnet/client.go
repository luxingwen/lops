package quicnet

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
	quic "github.com/quic-go/quic-go"
)

type Client struct {
	session        quic.Connection
	stream         quic.Stream
	tm             *TaskManager
	msg            chan *Message
	messageHandler *MessageHandler

	MachineID string
	Hostname  string
	IP        string
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

	udpAddr := getLocalIP(session)
	if udpAddr == nil {

	}

	machineID, err := machineid.ID()
	if err != nil {
		return nil, err
	}

	hostname, err := machineid.Hostname()
	if err != nil {
		return nil, err
	}

	c := &Client{
		IP:             udpAddr.IP.String(),
		MachineID:      machineID,
		session:        session,
		stream:         stream,
		Hostname:       hostname,
		tm:             NewTaskManager(),
		msg:            make(chan *Message, 100),
		messageHandler: NewMessageHandler(100),
	}
	c.messageHandler.RegisterHandler("script_task", HandlerScriptTask)
	go c.prosessMsg()
	go c.run()
	go c.StartHeartbeat(60 * time.Second)
	return c, nil
}

func (c *Client) Close() {
	c.stream.Close()
	c.session.CloseWithError(0, "")
	close(c.msg)
}

func (c *Client) SendMsg(msg *Message) error {
	c.msg <- msg
	return nil
}

func (c *Client) prosessMsg() {
	for msg := range c.msg {
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		c.Write(data)
	}
}

func (c *Client) run() {
	for {
		data, err := c.Read()
		if err != nil {
			continue
		}
		var msg Message
		err = json.Unmarshal(data, &msg)
		if err != nil {
			continue
		}
		c.messageHandler.SubmitMessage(&msg)
	}
}

func (c *Client) Write(data []byte) (int, error) {
	return writePacket(c.stream, data)
}

func (c *Client) Read() ([]byte, error) {
	return readPacket(c.stream)
}

type HeartbeatData struct {
	MachineID string `json:"machine_id"`
	Hostname  string `json:"hostname"`
	IP        string `json:"ip"`
}

func (c *Client) StartHeartbeat(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			heartbeatData := HeartbeatData{
				MachineID: c.MachineID,
				Hostname:  c.Hostname,
				IP:        c.IP,
			}

			data, err := json.Marshal(heartbeatData)
			if err != nil {
				fmt.Println("Failed to marshal heartbeat data:", err)
				continue
			}

			msg := Message{
				ID:   uuid.New().String(),
				Type: "heartbeat",
				Data: data,
			}

			c.SendMsg(&msg)
		}
	}()
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

func getLocalIP(session quic.Session) *net.UDPAddr {
	udpConn := session.ConnectionState().UDPConn
	if udpConn == nil {
		return nil
	}

	localAddr := udpConn.LocalAddr().(*net.UDPAddr)
	return localAddr
}
