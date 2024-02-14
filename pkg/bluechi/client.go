package bluechi

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	ServiceName    = "org.eclipse.bluechi"
	BaseObjectPath = "/org/eclipse/bluechi"

	ControllerInterface   = "org.eclipse.bluechi.Controller"
	NodeInterface         = "org.eclipse.bluechi.Node"
	MonitorInterface      = "org.eclipse.bluechi.Monitor"
	DBusPropertyInterface = "org.freedesktop.DBus.Properties"

	ControllerMethodListNodes     = "ListNodes"
	ControllerMethodCreateMonitor = "CreateMonitor"

	MonitorMethodSubscribe = "Subscribe"

	SignalMonitorUnitNew               = "UnitNew"
	SignalMonitorUnitRemoved           = "UnitRemoved"
	SignalMonitorUnitStateChanged      = "UnitStateChanged"
	SignalMonitorUnitPropertiesChanged = "UnitPropertiesChanged"

	NodePathPrefix = "/org/eclipse/bluechi/node"
)

type Client struct {
	Conn *dbus.Conn
}

func NewClient(conn *dbus.Conn) *Client {
	return &Client{
		Conn: conn,
	}
}

func (m *Client) CreateMonitor() (*Monitor, error) {
	return CreateMonitor(m.Conn)
}

func (m *Client) GetNodes() (Nodes, error) {
	var nodes [][]interface{}
	busObject := m.Conn.Object(ServiceName, BaseObjectPath)
	err := busObject.Call(fmt.Sprintf("%s.%s", ControllerInterface, ControllerMethodListNodes), 0).Store(&nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	bluechiNodes := make(Nodes)
	for _, node := range nodes {
		name := fmt.Sprintf("%s", node[0])
		status := fmt.Sprintf("%s", node[2])
		bluechiNodes[name] = &Node{
			Name:              name,
			Status:            status,
			LastSeenTimestamp: "",
		}
	}
	return bluechiNodes, nil
}
