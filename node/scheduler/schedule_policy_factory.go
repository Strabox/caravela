package scheduler

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/scheduler/policies"
	"github.com/strabox/caravela/node/scheduler/policies/binpack"
	"github.com/strabox/caravela/node/scheduler/policies/spread"
	"strings"
)

// SchedulePolicyFactory represents a method that creates a new scheduling policy method.
type SchedulePolicyFactory func() (policies.SchedulingPolicy, error)

// schedulingPolicies holds all the registered scheduling policies available.
var schedulePolicies = make(map[string]SchedulePolicyFactory)

// init initializes our predefined scheduling policies.
func init() {
	RegisterSchedulePolicy("binpack", binpack.NewBinPackSchedulePolicy)
	RegisterSchedulePolicy("spread", spread.NewSpreadSchedulePolicy)

}

// RegisterSchedulePolicy can be used to register a schedule policy in order to be available.
func RegisterSchedulePolicy(schedulePolicyName string, factory SchedulePolicyFactory) {
	if factory == nil {
		log.Panic(fmt.Errorf("nil %s factory being registred", schedulePolicyName))
	}
	_, exist := schedulePolicies[schedulePolicyName]
	if exist {
		log.Warnf("%s scheduling policy is being overridden", schedulePolicyName)
	}
	schedulePolicies[schedulePolicyName] = factory
}

// CreateSchedulePolicy creates a new schedule policy depending on the configurations.
func CreateSchedulePolicy(config *configuration.Configuration) policies.SchedulingPolicy {
	configuredSchedulePolicy := config.SchedulingPolicy()

	schedulePolicyFactory, exist := schedulePolicies[configuredSchedulePolicy]
	if !exist {
		existingPolicies := make([]string, len(schedulePolicies))
		for policyName := range schedulePolicies {
			existingPolicies = append(existingPolicies, policyName)
		}
		err := fmt.Errorf("invalid %s schedule policy. Schedule policies available: %s",
			configuredSchedulePolicy, strings.Join(existingPolicies, ", "))
		log.Panic(err)
	}

	schedulePolicy, err := schedulePolicyFactory()
	if err != nil {
		log.Panic(err)
	}

	return schedulePolicy
}
