/* SPDX-License-Identifier: LGPL-2.1-or-later */

package main

import (
	"fmt"
	"os"

	"github.com/engelmi/bluechi-bridge/pkg"
	"github.com/engelmi/bluechi-bridge/pkg/bluechi"
	"github.com/engelmi/bluechi-bridge/pkg/k8s"
	"github.com/godbus/dbus/v5"
)

func main() {

	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to system bus: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	k8sClient, err := k8s.NewClient("bluechi", "bluechi-k8s-bridge", "laptop")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create k8s client: ", err)
		os.Exit(1)
	}
	bridge := pkg.NewBridge(bluechi.NewClient(conn), k8sClient)
	bridge.InitState()
	bridge.SetupMonitor()
	bridge.Start()
}
