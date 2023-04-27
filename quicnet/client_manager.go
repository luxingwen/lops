package quicnet

import (
	"fmt"
	"sync"
)

type ClientManager struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*Client),
	}
}

func (cm *ClientManager) AddClient(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.clients[client.MachineID] = client
}

func (cm *ClientManager) RemoveClient(machineID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.clients, machineID)
}

func (cm *ClientManager) GetClient(machineID string) *Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clients[machineID]
}

func (cm *ClientManager) HandleHeartbeat(heartbeatData *HeartbeatData) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	client, exists := cm.clients[heartbeatData.MachineID]
	if !exists {
		fmt.Printf("Received heartbeat from unknown client: %s\n", heartbeatData.MachineID)
		return
	}

	// Update client data from heartbeat
	client.Hostname = heartbeatData.Hostname
	client.IP = heartbeatData.IP
}
