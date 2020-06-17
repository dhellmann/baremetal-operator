package baremetalhost

import (
	"fmt"

	metal3 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha2"
	"github.com/metal3-io/baremetal-operator/pkg/provisioner"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// hostStateMachine is a finite state machine that manages transitions between
// the states of a BareMetalHost.
type hostStateMachine struct {
	Host        *metal3.BareMetalHost
	NextState   metal3.ProvisioningState
	Reconciler  *ReconcileBareMetalHost
	Provisioner provisioner.Provisioner
}

func newHostStateMachine(host *metal3.BareMetalHost,
	reconciler *ReconcileBareMetalHost,
	provisioner provisioner.Provisioner) *hostStateMachine {
	currentState := host.Status.Provisioning.State
	r := hostStateMachine{
		Host:        host,
		NextState:   currentState, // Remain in current state by default
		Reconciler:  reconciler,
		Provisioner: provisioner,
	}
	return &r
}

type stateHandler func(*reconcileInfo) actionResult

func (hsm *hostStateMachine) handlers() map[metal3.ProvisioningState]stateHandler {
	return map[metal3.ProvisioningState]stateHandler{
		metal3.StateNone:                  hsm.handleNone,
		metal3.StateRegistering:           hsm.handleRegistering,
		metal3.StateRegistrationError:     hsm.handleRegistrationError,
		metal3.StateInspecting:            hsm.handleInspecting,
		metal3.StateExternallyProvisioned: hsm.handleExternallyProvisioned,
		metal3.StateMatchProfile:          hsm.handleMatchProfile,
		metal3.StateReady:                 hsm.handleReady,
		metal3.StateProvisioning:          hsm.handleProvisioning,
		metal3.StateProvisioningError:     hsm.handleProvisioningError,
		metal3.StateProvisioned:           hsm.handleProvisioned,
		metal3.StatePowerManagementError:  hsm.handlePowerManagementError,
		metal3.StateDeprovisioning:        hsm.handleDeprovisioning,
		metal3.StateDeleting:              hsm.handleDeleting,
	}
}

func recordStateBegin(host *metal3.BareMetalHost, state metal3.ProvisioningState, time metav1.Time) {
	if nextMetric := host.OperationMetricForState(state); nextMetric != nil {
		if nextMetric.Start.IsZero() || !nextMetric.End.IsZero() {
			*nextMetric = metal3.OperationMetric{
				Start: time,
			}
		}
	}
}

func recordStateEnd(info *reconcileInfo, host *metal3.BareMetalHost, state metal3.ProvisioningState, time metav1.Time) {
	if prevMetric := host.OperationMetricForState(state); prevMetric != nil {
		if !prevMetric.Start.IsZero() {
			prevMetric.End = time
			info.postSaveCallbacks = append(info.postSaveCallbacks, func() {
				observer := stateTime[state].With(hostMetricLabels(info.request))
				observer.Observe(prevMetric.Duration().Seconds())
			})
		}
	}
}

func (hsm *hostStateMachine) updateHostStateFrom(initialState metal3.ProvisioningState,
	info *reconcileInfo) {
	if hsm.NextState != initialState {
		info.log.Info("changing provisioning state",
			"old", initialState,
			"new", hsm.NextState)
		now := metav1.Now()
		recordStateEnd(info, hsm.Host, initialState, now)
		recordStateBegin(hsm.Host, hsm.NextState, now)
		info.postSaveCallbacks = append(info.postSaveCallbacks, func() {
			stateChanges.With(stateChangeMetricLabels(initialState, hsm.NextState)).Inc()
		})
		hsm.Host.Status.Provisioning.State = hsm.NextState
	}
}

func (hsm *hostStateMachine) ReconcileState(info *reconcileInfo) actionResult {
	initialState := hsm.Host.Status.Provisioning.State
	defer hsm.updateHostStateFrom(initialState, info)

	if hsm.checkInitiateDelete() {
		info.log.Info("Initiating host deletion")
		return actionComplete{}
	}
	// TODO: In future we should always re-register the host if required,
	// rather than initiate a transistion back to the Registering state.
	if hsm.shouldInitiateRegister(info) {
		info.log.Info("Initiating host registration")
		hostRegistrationRequired.Inc()
		return actionComplete{}
	}

	if stateHandler, found := hsm.handlers()[initialState]; found {
		return stateHandler(info)
	}

	info.log.Info("No handler found for state", "state", initialState)
	return actionError{fmt.Errorf("No handler found for state \"%s\"", initialState)}
}

func (hsm *hostStateMachine) checkInitiateDelete() bool {
	if hsm.Host.DeletionTimestamp.IsZero() {
		// Delete not requested
		return false
	}

	switch hsm.NextState {
	default:
		hsm.NextState = metal3.StateDeleting
	case metal3.StateProvisioning, metal3.StateProvisioningError, metal3.StateProvisioned:
		hsm.NextState = metal3.StateDeprovisioning
	case metal3.StateDeprovisioning:
		// Allow state machine to run to continue deprovisioning.
		return false
	case metal3.StateDeleting:
		// Already in deleting state. Allow state machine to run.
		return false
	}
	return true
}

func (hsm *hostStateMachine) shouldInitiateRegister(info *reconcileInfo) bool {
	changeState := false
	if hsm.Host.DeletionTimestamp.IsZero() {
		switch hsm.NextState {
		default:
			changeState = !hsm.Host.Status.GoodCredentials.Match(*info.bmcCredsSecret)
		case metal3.StateNone:
		case metal3.StateRegistering, metal3.StateRegistrationError:
		case metal3.StateDeleting:
		}
	}
	if changeState {
		hsm.NextState = metal3.StateRegistering
	}
	return changeState
}

func (hsm *hostStateMachine) handleNone(info *reconcileInfo) actionResult {
	// Running the state machine at all means we have successfully validated
	// the BMC credentials once, so we can move to the Registering state.
	hsm.Host.ClearError()
	hsm.NextState = metal3.StateRegistering
	return actionComplete{}
}

func (hsm *hostStateMachine) handleRegistering(info *reconcileInfo) actionResult {
	actResult := hsm.Reconciler.actionRegistering(hsm.Provisioner, info)

	switch actResult.(type) {
	case actionComplete:
		// TODO: In future this state should only occur before the host is
		// registered the first time (though we must always check and
		// re-register the host regardless of the current state). That will
		// eliminate the need to determine which state we came from here.
		switch {
		case hsm.Host.Spec.ExternallyProvisioned:
			hsm.NextState = metal3.StateExternallyProvisioned
		case hsm.Host.WasProvisioned():
			hsm.NextState = metal3.StateProvisioned
		case hsm.Host.NeedsHardwareInspection():
			hsm.NextState = metal3.StateInspecting
		case hsm.Host.NeedsHardwareProfile():
			hsm.NextState = metal3.StateMatchProfile
		default:
			hsm.NextState = metal3.StateReady
		}
	case actionFailed:
		hsm.NextState = metal3.StateRegistrationError
	}
	return actResult
}

func (hsm *hostStateMachine) handleRegistrationError(info *reconcileInfo) actionResult {
	if !hsm.Host.Status.TriedCredentials.Match(*info.bmcCredsSecret) {
		info.log.Info("Modified credentials detected; will retry registration")
		hsm.NextState = metal3.StateRegistering
		return actionComplete{}
	}
	return actionFailed{}
}

func (hsm *hostStateMachine) handleInspecting(info *reconcileInfo) actionResult {
	actResult := hsm.Reconciler.actionInspecting(hsm.Provisioner, info)
	if _, complete := actResult.(actionComplete); complete {
		hsm.NextState = metal3.StateMatchProfile
	}
	return actResult
}

func (hsm *hostStateMachine) handleMatchProfile(info *reconcileInfo) actionResult {
	actResult := hsm.Reconciler.actionMatchProfile(hsm.Provisioner, info)
	if _, complete := actResult.(actionComplete); complete {
		hsm.NextState = metal3.StateReady
	}
	return actResult
}

func (hsm *hostStateMachine) handleExternallyProvisioned(info *reconcileInfo) actionResult {
	if hsm.Host.Spec.ExternallyProvisioned {
		actResult := hsm.Reconciler.actionManageSteadyState(hsm.Provisioner, info)
		if r, f := actResult.(actionFailed); f {
			switch r.ErrorType {
			case metal3.PowerManagementError:
				hsm.NextState = metal3.StatePowerManagementError
			case metal3.RegistrationError:
				hsm.NextState = metal3.StateRegistrationError
			}
		}
		return actResult
	}

	switch {
	case hsm.Host.NeedsHardwareInspection():
		hsm.NextState = metal3.StateInspecting
	case hsm.Host.NeedsHardwareProfile():
		hsm.NextState = metal3.StateMatchProfile
	default:
		hsm.NextState = metal3.StateReady
	}
	return actionComplete{}
}

func (hsm *hostStateMachine) handleReady(info *reconcileInfo) actionResult {
	if hsm.Host.Spec.ExternallyProvisioned {
		hsm.NextState = metal3.StateExternallyProvisioned
		return actionComplete{}
	}

	actResult := hsm.Reconciler.actionManageReady(hsm.Provisioner, info)

	switch r := actResult.(type) {
	case actionComplete:
		hsm.NextState = metal3.StateProvisioning
	case actionFailed:
		switch r.ErrorType {
		case metal3.PowerManagementError:
			hsm.NextState = metal3.StatePowerManagementError
		case metal3.RegistrationError:
			hsm.NextState = metal3.StateRegistrationError
		}
	}
	return actResult
}

func (hsm *hostStateMachine) handleProvisioning(info *reconcileInfo) actionResult {
	if hsm.Host.NeedsDeprovisioning() {
		hsm.NextState = metal3.StateDeprovisioning
		return actionComplete{}
	}

	actResult := hsm.Reconciler.actionProvisioning(hsm.Provisioner, info)
	switch actResult.(type) {
	case actionComplete:
		hsm.NextState = metal3.StateProvisioned
	case actionFailed:
		hsm.NextState = metal3.StateProvisioningError
	}
	return actResult
}

func (hsm *hostStateMachine) handleProvisioningError(info *reconcileInfo) actionResult {
	switch {
	case hsm.Host.Spec.ExternallyProvisioned:
		hsm.NextState = metal3.StateExternallyProvisioned
	default:
		hsm.NextState = metal3.StateDeprovisioning
	}
	return actionComplete{}
}

func (hsm *hostStateMachine) handleProvisioned(info *reconcileInfo) actionResult {
	if hsm.Host.NeedsDeprovisioning() {
		hsm.NextState = metal3.StateDeprovisioning
		return actionComplete{}
	}

	actResult := hsm.Reconciler.actionManageSteadyState(hsm.Provisioner, info)
	if r, f := actResult.(actionFailed); f {
		switch r.ErrorType {
		case metal3.PowerManagementError:
			hsm.NextState = metal3.StatePowerManagementError
		case metal3.RegistrationError:
			hsm.NextState = metal3.StateRegistrationError
		}
	}
	return actResult
}

func (hsm *hostStateMachine) handlePowerManagementError(info *reconcileInfo) actionResult {
	switch {
	case hsm.Host.Spec.ExternallyProvisioned:
		hsm.NextState = metal3.StateExternallyProvisioned
	case hsm.Host.WasProvisioned():
		hsm.NextState = metal3.StateProvisioned
	default:
		hsm.NextState = metal3.StateReady
	}
	return actionComplete{}
}

func (hsm *hostStateMachine) handleDeprovisioning(info *reconcileInfo) actionResult {
	actResult := hsm.Reconciler.actionDeprovisioning(hsm.Provisioner, info)

	switch actResult.(type) {
	case actionComplete:
		if !hsm.Host.DeletionTimestamp.IsZero() {
			hsm.NextState = metal3.StateDeleting
		} else {
			hsm.NextState = metal3.StateReady
		}
	case actionFailed:
		if !hsm.Host.DeletionTimestamp.IsZero() {
			// If the provisioner gives up deprovisioning and
			// deletion has been requested, continue to delete.
			// Note that this is entirely theoretical, as the
			// Ironic provisioner currently never gives up
			// trying to deprovision.
			hsm.NextState = metal3.StateDeleting
			info.postSaveCallbacks = append(info.postSaveCallbacks, deleteWithoutDeprov.Inc)
			actResult = actionComplete{}
		} else {
			hsm.NextState = metal3.StateProvisioningError
		}
	}
	return actResult
}

func (hsm *hostStateMachine) handleDeleting(info *reconcileInfo) actionResult {
	return hsm.Reconciler.actionDeleting(hsm.Provisioner, info)
}
