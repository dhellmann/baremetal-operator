package baremetalhost

import (
	"fmt"

	metal3shared "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/shared"
	metal3 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha2"
	"github.com/metal3-io/baremetal-operator/pkg/provisioner"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// hostStateMachine is a finite state machine that manages transitions between
// the states of a BareMetalHost.
type hostStateMachine struct {
	Host        *metal3.BareMetalHost
	NextState   metal3shared.ProvisioningState
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

func (hsm *hostStateMachine) handlers() map[metal3shared.ProvisioningState]stateHandler {
	return map[metal3shared.ProvisioningState]stateHandler{
		metal3shared.StateNone:                  hsm.handleNone,
		metal3shared.StateRegistering:           hsm.handleRegistering,
		metal3shared.StateRegistrationError:     hsm.handleRegistrationError,
		metal3shared.StateInspecting:            hsm.handleInspecting,
		metal3shared.StateExternallyProvisioned: hsm.handleExternallyProvisioned,
		metal3shared.StateMatchProfile:          hsm.handleMatchProfile,
		metal3shared.StateReady:                 hsm.handleReady,
		metal3shared.StateProvisioning:          hsm.handleProvisioning,
		metal3shared.StateProvisioningError:     hsm.handleProvisioningError,
		metal3shared.StateProvisioned:           hsm.handleProvisioned,
		metal3shared.StatePowerManagementError:  hsm.handlePowerManagementError,
		metal3shared.StateDeprovisioning:        hsm.handleDeprovisioning,
		metal3shared.StateDeleting:              hsm.handleDeleting,
	}
}

func recordStateBegin(host *metal3.BareMetalHost, state metal3shared.ProvisioningState, time metav1.Time) {
	if nextMetric := host.OperationMetricForState(state); nextMetric != nil {
		if nextMetric.Start.IsZero() || !nextMetric.End.IsZero() {
			*nextMetric = metal3shared.OperationMetric{
				Start: time,
			}
		}
	}
}

func recordStateEnd(info *reconcileInfo, host *metal3.BareMetalHost, state metal3shared.ProvisioningState, time metav1.Time) {
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

func (hsm *hostStateMachine) updateHostStateFrom(initialState metal3shared.ProvisioningState,
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
		hsm.NextState = metal3shared.StateDeleting
	case metal3shared.StateProvisioning, metal3shared.StateProvisioningError, metal3shared.StateProvisioned:
		hsm.NextState = metal3shared.StateDeprovisioning
	case metal3shared.StateDeprovisioning:
		// Allow state machine to run to continue deprovisioning.
		return false
	case metal3shared.StateDeleting:
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
		case metal3shared.StateNone:
		case metal3shared.StateRegistering, metal3shared.StateRegistrationError:
		case metal3shared.StateDeleting:
		}
	}
	if changeState {
		hsm.NextState = metal3shared.StateRegistering
	}
	return changeState
}

func (hsm *hostStateMachine) handleNone(info *reconcileInfo) actionResult {
	// Running the state machine at all means we have successfully validated
	// the BMC credentials once, so we can move to the Registering state.
	hsm.Host.ClearError()
	hsm.NextState = metal3shared.StateRegistering
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
			hsm.NextState = metal3shared.StateExternallyProvisioned
		case hsm.Host.WasProvisioned():
			hsm.NextState = metal3shared.StateProvisioned
		case hsm.Host.NeedsHardwareInspection():
			hsm.NextState = metal3shared.StateInspecting
		case hsm.Host.NeedsHardwareProfile():
			hsm.NextState = metal3shared.StateMatchProfile
		default:
			hsm.NextState = metal3shared.StateReady
		}
	case actionFailed:
		hsm.NextState = metal3shared.StateRegistrationError
	}
	return actResult
}

func (hsm *hostStateMachine) handleRegistrationError(info *reconcileInfo) actionResult {
	if !hsm.Host.Status.TriedCredentials.Match(*info.bmcCredsSecret) {
		info.log.Info("Modified credentials detected; will retry registration")
		hsm.NextState = metal3shared.StateRegistering
		return actionComplete{}
	}
	return actionFailed{}
}

func (hsm *hostStateMachine) handleInspecting(info *reconcileInfo) actionResult {
	actResult := hsm.Reconciler.actionInspecting(hsm.Provisioner, info)
	if _, complete := actResult.(actionComplete); complete {
		hsm.NextState = metal3shared.StateMatchProfile
	}
	return actResult
}

func (hsm *hostStateMachine) handleMatchProfile(info *reconcileInfo) actionResult {
	actResult := hsm.Reconciler.actionMatchProfile(hsm.Provisioner, info)
	if _, complete := actResult.(actionComplete); complete {
		hsm.NextState = metal3shared.StateReady
	}
	return actResult
}

func (hsm *hostStateMachine) handleExternallyProvisioned(info *reconcileInfo) actionResult {
	if hsm.Host.Spec.ExternallyProvisioned {
		actResult := hsm.Reconciler.actionManageSteadyState(hsm.Provisioner, info)
		if r, f := actResult.(actionFailed); f {
			switch r.ErrorType {
			case metal3shared.PowerManagementError:
				hsm.NextState = metal3shared.StatePowerManagementError
			case metal3shared.RegistrationError:
				hsm.NextState = metal3shared.StateRegistrationError
			}
		}
		return actResult
	}

	switch {
	case hsm.Host.NeedsHardwareInspection():
		hsm.NextState = metal3shared.StateInspecting
	case hsm.Host.NeedsHardwareProfile():
		hsm.NextState = metal3shared.StateMatchProfile
	default:
		hsm.NextState = metal3shared.StateReady
	}
	return actionComplete{}
}

func (hsm *hostStateMachine) handleReady(info *reconcileInfo) actionResult {
	if hsm.Host.Spec.ExternallyProvisioned {
		hsm.NextState = metal3shared.StateExternallyProvisioned
		return actionComplete{}
	}

	actResult := hsm.Reconciler.actionManageReady(hsm.Provisioner, info)

	switch r := actResult.(type) {
	case actionComplete:
		hsm.NextState = metal3shared.StateProvisioning
	case actionFailed:
		switch r.ErrorType {
		case metal3shared.PowerManagementError:
			hsm.NextState = metal3shared.StatePowerManagementError
		case metal3shared.RegistrationError:
			hsm.NextState = metal3shared.StateRegistrationError
		}
	}
	return actResult
}

func (hsm *hostStateMachine) handleProvisioning(info *reconcileInfo) actionResult {
	if hsm.Host.NeedsDeprovisioning() {
		hsm.NextState = metal3shared.StateDeprovisioning
		return actionComplete{}
	}

	actResult := hsm.Reconciler.actionProvisioning(hsm.Provisioner, info)
	switch actResult.(type) {
	case actionComplete:
		hsm.NextState = metal3shared.StateProvisioned
	case actionFailed:
		hsm.NextState = metal3shared.StateProvisioningError
	}
	return actResult
}

func (hsm *hostStateMachine) handleProvisioningError(info *reconcileInfo) actionResult {
	switch {
	case hsm.Host.Spec.ExternallyProvisioned:
		hsm.NextState = metal3shared.StateExternallyProvisioned
	default:
		hsm.NextState = metal3shared.StateDeprovisioning
	}
	return actionComplete{}
}

func (hsm *hostStateMachine) handleProvisioned(info *reconcileInfo) actionResult {
	if hsm.Host.NeedsDeprovisioning() {
		hsm.NextState = metal3shared.StateDeprovisioning
		return actionComplete{}
	}

	actResult := hsm.Reconciler.actionManageSteadyState(hsm.Provisioner, info)
	if r, f := actResult.(actionFailed); f {
		switch r.ErrorType {
		case metal3shared.PowerManagementError:
			hsm.NextState = metal3shared.StatePowerManagementError
		case metal3shared.RegistrationError:
			hsm.NextState = metal3shared.StateRegistrationError
		}
	}
	return actResult
}

func (hsm *hostStateMachine) handlePowerManagementError(info *reconcileInfo) actionResult {
	switch {
	case hsm.Host.Spec.ExternallyProvisioned:
		hsm.NextState = metal3shared.StateExternallyProvisioned
	case hsm.Host.WasProvisioned():
		hsm.NextState = metal3shared.StateProvisioned
	default:
		hsm.NextState = metal3shared.StateReady
	}
	return actionComplete{}
}

func (hsm *hostStateMachine) handleDeprovisioning(info *reconcileInfo) actionResult {
	actResult := hsm.Reconciler.actionDeprovisioning(hsm.Provisioner, info)

	switch actResult.(type) {
	case actionComplete:
		if !hsm.Host.DeletionTimestamp.IsZero() {
			hsm.NextState = metal3shared.StateDeleting
		} else {
			hsm.NextState = metal3shared.StateReady
		}
	case actionFailed:
		if !hsm.Host.DeletionTimestamp.IsZero() {
			// If the provisioner gives up deprovisioning and
			// deletion has been requested, continue to delete.
			// Note that this is entirely theoretical, as the
			// Ironic provisioner currently never gives up
			// trying to deprovision.
			hsm.NextState = metal3shared.StateDeleting
			info.postSaveCallbacks = append(info.postSaveCallbacks, deleteWithoutDeprov.Inc)
			actResult = actionComplete{}
		} else {
			hsm.NextState = metal3shared.StateProvisioningError
		}
	}
	return actResult
}

func (hsm *hostStateMachine) handleDeleting(info *reconcileInfo) actionResult {
	return hsm.Reconciler.actionDeleting(hsm.Provisioner, info)
}
