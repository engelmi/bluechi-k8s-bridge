package bluechi

type BlueChiState struct {
	Version uint64
	Nodes   Nodes
	Units   Units
}

type Nodes map[string]*Node

type Node struct {
	Name              string
	Status            string
	LastSeenTimestamp string
}

type Units map[string]*Unit

type Unit struct {
	Name string
}
