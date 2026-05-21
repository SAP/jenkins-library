package config

import (
	"strconv"

	"github.com/SAP/jenkins-library/pkg/log"
)

// authAlternativeStatus tracks, per step, which authentication alternative
// groups have all of their parameters resolved. It is built once during
// GetStepConfig after every cheaper credential source (config, env, CLI,
// System Trust) has had a chance to populate parameters and is then consulted
// by Vault resolution to skip groups that are already satisfied.
type authAlternativeStatus struct {
	groups          []AuthAlternative
	paramToGroupID  map[string]string
	resolvedGroupID map[string]bool
}

// newAuthAlternativeStatus builds the gate. It validates the declared groups
// (drops invalid entries with a warning) and computes which groups are
// already resolved given the current step config.
//
// A group is "resolved" when EVERY parameter listed in it has a non-empty
// value in stepConfig at the moment this function is called. Callers are
// expected to invoke it AFTER System Trust resolution, so a token that was
// just fetched from System Trust counts as resolving its group.
func newAuthAlternativeStatus(groups []AuthAlternative, stepConfig map[string]interface{}, knownParams []StepParameters) *authAlternativeStatus {
	status := &authAlternativeStatus{
		paramToGroupID:  map[string]string{},
		resolvedGroupID: map[string]bool{},
	}
	if len(groups) == 0 {
		return status
	}
	// Some callers (e.g. getConfig --contextConfig) intentionally pass an
	// empty knownParams slice while leaving the rest of the step metadata
	// populated. They never run Vault resolution, so there is nothing to
	// gate; treat this as a no-op rather than logging spurious "unknown
	// parameter" warnings for every authAlternatives entry.
	if len(knownParams) == 0 {
		return status
	}

	declaredParams := map[string]bool{}
	for _, p := range knownParams {
		declaredParams[p.Name] = true
	}

	for i, group := range groups {
		groupID := group.ID
		if groupID == "" {
			groupID = autoGroupID(i)
		}
		validParams := []string{}
		for _, paramName := range group.Params {
			if !declaredParams[paramName] {
				log.Entry().Warnf("authAlternatives: group %q references unknown parameter %q; ignoring this entry", groupID, paramName)
				continue
			}
			if existingGroup, ok := status.paramToGroupID[paramName]; ok {
				log.Entry().Warnf("authAlternatives: parameter %q already assigned to group %q; ignoring duplicate assignment to %q", paramName, existingGroup, groupID)
				continue
			}
			status.paramToGroupID[paramName] = groupID
			validParams = append(validParams, paramName)
		}
		if len(validParams) == 0 {
			continue
		}
		status.groups = append(status.groups, AuthAlternative{ID: groupID, Params: validParams})
	}

	for _, group := range status.groups {
		if allParamsHaveValue(stepConfig, group.Params) {
			status.resolvedGroupID[group.ID] = true
			log.Entry().Debugf("authAlternatives: group %q resolved; alternative Vault lookups will be skipped", group.ID)
		}
	}
	return status
}

// shouldSkipVault reports whether Vault resolution must be skipped for the
// given parameter. Vault is skipped when:
//   - the parameter belongs to one of the declared alternative groups, AND
//   - at least one OTHER group has all of its parameters resolved.
//
// Parameters that are not part of any group (e.g. dockerConfigJSON in
// sapCallStagingService) are never skipped by this gate.
func (s *authAlternativeStatus) shouldSkipVault(paramName string) (bool, string) {
	if s == nil || len(s.groups) == 0 {
		return false, ""
	}
	ownGroup, inGroup := s.paramToGroupID[paramName]
	if !inGroup {
		return false, ""
	}
	for resolvedID := range s.resolvedGroupID {
		if resolvedID != ownGroup {
			return true, resolvedID
		}
	}
	return false, ""
}

func allParamsHaveValue(config map[string]interface{}, params []string) bool {
	for _, name := range params {
		value, ok := config[name].(string)
		if !ok || value == "" {
			return false
		}
	}
	return true
}

func autoGroupID(index int) string {
	return "group-" + strconv.Itoa(index)
}
