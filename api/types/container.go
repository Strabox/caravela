package types

import "errors"

type ContainerConfig struct {
	Name         string        `json:"N"`
	ImageKey     string        `json:"IK"`
	Args         []string      `json:"A"`
	PortMappings []PortMapping `json:"PM"`
	Resources    Resources     `json:"FR"`
	GroupPolicy  GroupPolicy   `json:"GP"`
}

type ContainerStatus struct {
	ContainerConfig `json:"CC"`
	SupplierIP      string `json:"SIp"`
	ContainerID     string `json:"CId"`
	Status          string `json:"S"`
}

type PortMapping struct {
	HostPort      int    `json:"HP"`
	ContainerPort int    `json:"CP"`
	Protocol      string `json:"P"`
}

// ======================= Container Group Policy ========================

type GroupPolicy uint

const (
	SpreadGroupPolicyStr     = "spread"
	CoLocationGroupPolicyStr = "co-location"
)

const (
	SpreadGroupPolicy GroupPolicy = iota
	CoLocationGroupPolicy
)

var containerGroupPolicies = []string{SpreadGroupPolicyStr, CoLocationGroupPolicyStr}

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
