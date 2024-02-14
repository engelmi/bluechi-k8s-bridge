package k8s

import (
	"context"
	"fmt"
	"strconv"

	"github.com/engelmi/bluechi-bridge/pkg/bluechi"
	"github.com/engelmi/bluechi-server/pkg/apis/bluechi/v1alpha1"
	"github.com/engelmi/bluechi-server/pkg/generated/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Client struct {
	config *rest.Config
	client *versioned.Clientset

	namespace  string
	manager    string
	systemName string
}

func NewClient(namespace string, manager string, systemName string) (*Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s config: %v", err)
	}

	clientSet, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new config: %v", err)
	}

	return &Client{
		config: cfg,
		client: clientSet,

		namespace:  namespace,
		manager:    manager,
		systemName: systemName,
	}, nil
}

func (c *Client) Create(state *bluechi.BlueChiState) error {
	nodes := make(v1alpha1.BlueChiNodes, 0, len(state.Nodes))
	for _, node := range state.Nodes {
		nodes = append(nodes, v1alpha1.BlueChiNode{
			Name:              node.Name,
			Status:            v1alpha1.NodeStatus(node.Status),
			LastSeenTimestamp: node.LastSeenTimestamp,
		})
	}

	// TODO: units

	system := v1alpha1.BlueChiSystem{
		Spec: v1alpha1.BlueChiSystemSpec{
			Nodes: nodes,
		},
		Status:   "unknown",
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name:      c.systemName,
			Namespace: c.namespace,
		},
	}

	createdSystem, err := c.client.OrgV1alpha1().BlueChiSystems(c.namespace).Create(context.Background(),
		&system,
		v1.CreateOptions{
			FieldManager: c.manager,
		})
	if err != nil {
		return fmt.Errorf("failed to create bluechi system in k8s: %v", err)
	}

	version, err := strconv.ParseUint(createdSystem.ResourceVersion, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse new version '%s': %v", createdSystem.ResourceVersion, err)
	}
	state.Version = version

	return nil
}

func (c *Client) Update(changes *bluechi.BlueChiState) error {
	nodes := make(v1alpha1.BlueChiNodes, 0, len(changes.Nodes))
	for _, node := range changes.Nodes {
		fmt.Printf("node: '%s'\n", node.Status)
		nodes = append(nodes, v1alpha1.BlueChiNode{
			Name:              node.Name,
			Status:            v1alpha1.NodeStatus(node.Status),
			LastSeenTimestamp: node.LastSeenTimestamp,
		})
	}

	// TODO: units

	newVersion := changes.Version + 1
	v := strconv.FormatUint(newVersion, 10)
	system := v1alpha1.BlueChiSystem{
		Spec: v1alpha1.BlueChiSystemSpec{
			Nodes: nodes,
		},
		Status:   "unknown",
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name:            c.systemName,
			Namespace:       c.namespace,
			ResourceVersion: v,
		},
	}

	updatedSystem, err := c.client.OrgV1alpha1().BlueChiSystems(c.namespace).Update(context.Background(),
		&system,
		v1.UpdateOptions{
			FieldManager: c.manager,
		})
	if err != nil {
		return fmt.Errorf("failed to create bluechi system in k8s: %v", err)
	}

	version, err := strconv.ParseUint(updatedSystem.ResourceVersion, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse new version '%s': %v", updatedSystem.ResourceVersion, err)
	}
	changes.Version = version

	return nil
}
