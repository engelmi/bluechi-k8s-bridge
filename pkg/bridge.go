package pkg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/engelmi/bluechi-bridge/pkg/bluechi"
	"github.com/engelmi/bluechi-bridge/pkg/k8s"
	"github.com/godbus/dbus/v5"
)

type BlueChiToK8sBridge struct {
	blueChiClient *bluechi.Client
	k8sClient     *k8s.Client

	State *bluechi.BlueChiState
}

func NewBridge(bluechiClient *bluechi.Client, k8sClient *k8s.Client) *BlueChiToK8sBridge {
	return &BlueChiToK8sBridge{
		blueChiClient: bluechiClient,
		k8sClient:     k8sClient,
	}
}

func (b *BlueChiToK8sBridge) InitState() error {
	nodes, err := b.blueChiClient.GetNodes()
	if err != nil {
		return fmt.Errorf("failed to get initial nodes: %v", err)
	}

	b.State = &bluechi.BlueChiState{
		Nodes: nodes,
		Units: make(bluechi.Units),
	}

	if err = b.k8sClient.Create(b.State); err != nil {
		return err
	}
	return nil
}

func (b *BlueChiToK8sBridge) SetupMonitor() error {

	monitor, err := b.blueChiClient.CreateMonitor()
	if err != nil {
		return err
	}
	//if err = monitor.AddSubscription("*", "*"); err != nil {
	//	return err
	//}

	nodeNames := make([]string, 0, len(b.State.Nodes))
	for name := range b.State.Nodes {
		nodeNames = append(nodeNames, name)
	}
	if err = monitor.AddNodeStatusChangedSignal(nodeNames); err != nil {
		return err
	}

	return nil
}

func (b *BlueChiToK8sBridge) Start() error {
	c := make(chan *dbus.Signal, 10)
	b.blueChiClient.Conn.Signal(c)
	fmt.Println("Starting...")
	for s := range c {
		if strings.HasPrefix(s.Name, bluechi.MonitorInterface) {
			b.handleUnitEvent(s)
		} else if strings.HasPrefix(s.Name, bluechi.DBusPropertyInterface) {
			b.handleNodeEvent(s)
		}
		if err := b.k8sClient.Update(b.State); err != nil {
			fmt.Printf("Failed to update state: %v", err)
		}
	}
	return nil
}

func (b *BlueChiToK8sBridge) handleNodeEvent(event *dbus.Signal) {
	changedValues, ok := event.Body[1].(map[string]dbus.Variant)
	if !ok {
		fmt.Println("Received invalid property changed signal")
		return
	}

	if val, ok := changedValues["Status"]; ok {
		nodeName := strings.Replace(string(event.Path), bluechi.NodePathPrefix+"/", "", 1)
		status, _ := strconv.Unquote(val.String())
		b.State.Nodes[nodeName].Status = status
	}
}

func (b *BlueChiToK8sBridge) handleUnitEvent(event *dbus.Signal) {
	if strings.HasSuffix(event.Name, bluechi.SignalMonitorUnitNew) {
		unitName := fmt.Sprintf("%s", event.Body[1])
		reason := fmt.Sprintf("%s", event.Body[2])

		if reason == "real" {
			b.State.Units[unitName] = &bluechi.Unit{
				Name: unitName,
			}
		}
	} else if strings.HasSuffix(event.Name, bluechi.SignalMonitorUnitRemoved) {
		unitName := fmt.Sprintf("%s", event.Body[1])
		reason := fmt.Sprintf("%s", event.Body[2])

		if reason == "real" {
			delete(b.State.Units, unitName)
		}
	} else if strings.HasSuffix(event.Name, bluechi.SignalMonitorUnitStateChanged) {
		fmt.Println("state changed")
	} else if strings.HasSuffix(event.Name, bluechi.SignalMonitorUnitPropertiesChanged) {
		fmt.Println("props changed")
	}
}
