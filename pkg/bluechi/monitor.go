package bluechi

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

type Monitor struct {
	dBusCon *dbus.Conn

	MonitorPath     string
	SubscriptionIDS []uint64
}

func CreateMonitor(conn *dbus.Conn) (*Monitor, error) {
	baseStub := conn.Object(ServiceName, BaseObjectPath)

	monitorPath := ""
	err := baseStub.Call(fmt.Sprintf("%s.%s", ControllerInterface, ControllerMethodCreateMonitor), 0).Store(&monitorPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Monitor: %v", err)
	}
	monitorStub := conn.Object(ServiceName, dbus.ObjectPath(monitorPath))
	err = monitorStub.AddMatchSignal("org.eclipse.bluechi.Monitor", "UnitNew").Err
	if err != nil {
		return nil, fmt.Errorf("failed to add signal to UnitNew: %v", err)
	}

	return &Monitor{
		dBusCon: conn,

		MonitorPath:     monitorPath,
		SubscriptionIDS: nil,
	}, nil
}

func (m *Monitor) getMonitorStub() dbus.BusObject {
	return m.dBusCon.Object(ServiceName, dbus.ObjectPath(m.MonitorPath))
}

func (m *Monitor) AddSubscription(node string, unit string) error {
	var subscriptionID uint64
	err := m.getMonitorStub().Call(
		fmt.Sprintf("%s.%s", MonitorInterface, MonitorMethodSubscribe),
		0, node, unit,
	).Store(&subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to add subscription on node '%s' for unit(s) '%s': %v", node, unit, err)
	}
	m.SubscriptionIDS = append(m.SubscriptionIDS, subscriptionID)
	return nil
}

func (m *Monitor) AddNodeStatusChangedSignal(nodeNames []string) error {
	for _, name := range nodeNames {
		path := dbus.ObjectPath(fmt.Sprintf("%s/%s", NodePathPrefix, name))
		err := m.dBusCon.AddMatchSignal(
			dbus.WithMatchSender(ServiceName),
			dbus.WithMatchObjectPath(path),
			dbus.WithMatchInterface(DBusPropertyInterface),
			dbus.WithMatchMember("PropertiesChanged"),
		)
		if err != nil {
			return fmt.Errorf("failed to add match for node status signal: %v", err)
		}
	}
	return nil
}
