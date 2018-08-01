package types

import "errors"

type ContainerConfig struct {
	Name         string        `json:"Name"`
	ImageKey     string        `json:"ImageKey"`
	Args         []string      `json:"Args"`
	PortMappings []PortMapping `json:"PortMappings"`
	Resources    Resources     `json:"Resources"`
	GroupPolicy  GroupPolicy   `json:"GroupPolicy"`
}

type ContainerStatus struct {
	ContainerConfig `json:"ContainerConfig"`
	SupplierIP      string `json:"SupplierIP"`
	ContainerID     string `json:"ContainerID"`
	Status          string `json:"Status"`
}

type PortMapping struct {
	HostPort      int `json:"HostPort"`
	ContainerPort int `json:"ContainerPort"`
}

// ======================= Container Group Policy ========================

type GroupPolicy uint

const (
	SpreadGroupPolicy GroupPolicy = iota
	CoLocationGroupPolicy
)

var containerGroupPolicies = []string{"spread", "co-location"}

func (gp GroupPolicy) name() string {
	return containerGroupPolicies[gp]
}
func (gp GroupPolicy) ordinal() int {
	return int(gp)
}
func (gp GroupPolicy) String() string {
	return containerGroupPolicies[gp]
}
func (gp GroupPolicy) values() *[]string {
	return &containerGroupPolicies
}

func (gp *GroupPolicy) ValueOf(arg string) error {
	for i, name := range containerGroupPolicies {
		if name == arg {
			*gp = GroupPolicy(i)
			return nil
		}
	}
	return errors.New("invalid enum value")
}
