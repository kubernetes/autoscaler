/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kamatera

import (
	"context"
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// Instance implements cloudprovider.Instance interface. Instance contains
// configuration info and functions to control a single Kamatera server instance.
type Instance struct {
	// Id is the cloud provider id.
	Id string
	// Status represents status of the node. (Optional)
	Status *cloudprovider.InstanceStatus
	// Kamatera specific fields
	PowerOn           bool
	Tags              []string
	StatusCommandId   string
	StatusCommandCode InstanceCommandCode
}

// InstanceErrorCode represents error codes for instance operations
type InstanceErrorCode string

const (
	// InstanceErrorDidNotStart indicates that the command failed to start and we don't have a command ID yet.
	InstanceErrorDidNotStart InstanceErrorCode = "InstanceErrorDidNotStart"
	// InstanceErrorGetCommandStatusFailed indicates that we failed to get the status of a command.
	InstanceErrorGetCommandStatusFailed InstanceErrorCode = "InstanceErrorGetCommandStatusFailed"
	// InstanceErrorCommandFailed indicates that the command failed during execution.
	InstanceErrorCommandFailed InstanceErrorCode = "InstanceErrorCommandFailed"
)

// InstanceCommandCode represents the command being executed on the instance
type InstanceCommandCode int

const (
	// InstanceCommandNone indicates that no command is being executed
	InstanceCommandNone InstanceCommandCode = 0
	// InstanceCommandPoweroff indicates that a poweroff command is being executed
	InstanceCommandPoweroff InstanceCommandCode = 1
	// InstanceCommandPoweron indicates that a poweron command is being executed
	InstanceCommandPoweron InstanceCommandCode = 2
	// InstanceCommandTerminate indicates that a terminate command is being executed
	InstanceCommandTerminate InstanceCommandCode = 3
	// InstanceCommandCreating indicates that a create server command is being executed
	InstanceCommandCreating InstanceCommandCode = 4
)

func (i *Instance) poweroff(client kamateraAPIClient, providerIDPrefix string) error {
	ctx := context.Background()
	serverName := parseKamateraProviderID(providerIDPrefix, i.Id)
	commandId, err := client.StartServerRequest(ctx, ServerRequestPoweroff, serverName)
	i.StatusCommandCode = InstanceCommandPoweroff
	if err != nil {
		klog.Errorf("Failed to poweroff server %s: %v", serverName, err)
		if i.Status != nil {
			i.Status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    string(InstanceErrorDidNotStart),
				ErrorMessage: err.Error(),
			}
		}
		return err
	}
	klog.V(2).Infof("Poweroff command started for server %s, command ID: %s", serverName, commandId)
	i.StatusCommandId = commandId
	return nil
}

func (i *Instance) terminate(client kamateraAPIClient, providerIDPrefix string) error {
	ctx := context.Background()
	serverName := parseKamateraProviderID(providerIDPrefix, i.Id)
	commandId, err := client.StartServerTerminate(ctx, serverName, true)
	i.StatusCommandCode = InstanceCommandTerminate
	if err != nil {
		klog.Errorf("Failed to terminate server %s: %v", serverName, err)
		if i.Status != nil {
			i.Status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    string(InstanceErrorDidNotStart),
				ErrorMessage: err.Error(),
			}
		}
		return err
	}
	klog.V(2).Infof("Terminate command started for server %s, command ID: %s", serverName, commandId)
	i.StatusCommandId = commandId
	return nil
}

// update instance status to deleting and start poweroff or terminate command
func (i *Instance) delete(client kamateraAPIClient, providerIDPrefix string, powerOffOnScaleDown bool) error {
	i.Status = &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting}
	if i.PowerOn {
		return i.poweroff(client, providerIDPrefix)
	}
	if !powerOffOnScaleDown {
		return i.terminate(client, providerIDPrefix)
	}
	i.Status = nil
	return nil
}

// update instance status to creating and start poweron command
func (i *Instance) createPoweron(client kamateraAPIClient, providerIDPrefix string) error {
	i.Status = &cloudprovider.InstanceStatus{State: cloudprovider.InstanceCreating}
	ctx := context.Background()
	serverName := parseKamateraProviderID(providerIDPrefix, i.Id)
	commandId, err := client.StartServerRequest(ctx, ServerRequestPoweron, serverName)
	i.StatusCommandCode = InstanceCommandPoweron
	if err != nil {
		klog.Errorf("Failed to poweron server %s: %v", serverName, err)
		if i.Status != nil {
			i.Status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    string(InstanceErrorDidNotStart),
				ErrorMessage: err.Error(),
			}
		}
		return err
	}
	klog.V(2).Infof("Poweron command started for server %s, command ID: %s", serverName, commandId)
	i.StatusCommandId = commandId
	return nil
}

func (i *Instance) extendedDebug() string {
	state := ""
	if i.Status == nil {
		state = "Unknown"
	} else if i.Status.State == cloudprovider.InstanceRunning {
		state = "Running"
	} else if i.Status.State == cloudprovider.InstanceCreating {
		state = "Creating"
	} else if i.Status.State == cloudprovider.InstanceDeleting {
		state = "Deleting"
	}
	return fmt.Sprintf("instance ID: %s state: %s powerOn: %v commandID: %s", i.Id, state, i.PowerOn, i.StatusCommandId)
}

// refresh updates the instance status by checking the status of any ongoing command
// if it returns true it means the instance needs to be deleted
func (i *Instance) refresh(
	client kamateraAPIClient, providerIDPrefix string, powerOffOnScaleDown bool,
	kubeClient kubernetes.Interface, hasServer bool,
) (needToDelete bool) {
	ctx := context.Background()
	serverName := parseKamateraProviderID(providerIDPrefix, i.Id)
	logPrefix := fmt.Sprintf("Kamatera server '%s'", serverName)
	if i.StatusCommandId == "" {
		if i.Status != nil && i.Status.State == cloudprovider.InstanceDeleting && !i.PowerOn {
			klog.V(2).Infof("%s: instance deleted and not powered on - setting statue to nil", logPrefix)
			i.Status = nil
		} else if !hasServer && i.Status == nil {
			klog.Warningf("%s: server not found and has no status - removing from instances", logPrefix)
			return true
		} else if !hasServer && i.Status != nil && i.Status.State == cloudprovider.InstanceDeleting {
			klog.V(2).Infof("%s: server not found and in deleting state - removing from instances", logPrefix)
			return true
		} else if i.PowerOn && i.Status == nil {
			klog.V(2).Infof("%s: node is powered on and status is nil, setting state to running", logPrefix)
			i.Status = &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}
		} else if i.PowerOn && i.Status.State == cloudprovider.InstanceCreating {
			klog.V(2).Infof("%s: node is powered on and status is creating, setting state to running", logPrefix)
			i.Status = &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}
		}
	} else {
		commandID := i.StatusCommandId
		commandCode := i.StatusCommandCode
		commandLogPrefix := fmt.Sprintf("%s command ID '%s':", logPrefix, commandID)
		// refresh the state of an ongoing command
		commandStatus, err := client.getCommandStatus(ctx, commandID)
		if err != nil {
			// failed to get command status - nothing we can do here, just update the ErrorInfo which will cause
			// the CA to handle the error appropriately
			klog.Errorf("%s failed to get command status: %v", commandLogPrefix, err)
			if i.Status != nil {
				i.Status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OtherErrorClass,
					ErrorCode:    string(InstanceErrorGetCommandStatusFailed),
					ErrorMessage: fmt.Sprintf("failed to get command %s status: %v", commandID, err),
				}
			}
		} else if commandStatus == CommandStatusError {
			// command completed with error - clear the command ID and set the error info
			klog.Errorf("%s command completed with error, check Kamatera console for details", commandLogPrefix)
			i.StatusCommandId = ""
			i.StatusCommandCode = InstanceCommandNone
			if i.Status != nil {
				i.Status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.OtherErrorClass,
					ErrorCode:    string(InstanceErrorCommandFailed),
					ErrorMessage: fmt.Sprintf("command %s ended with error", commandID),
				}
			}
		} else if commandStatus == CommandStatusComplete {
			// command completed successfully - clear the command ID and update the instance state depending on the command
			klog.V(0).Infof("%s command completed successfully", commandLogPrefix)
			i.StatusCommandId = ""
			i.StatusCommandCode = InstanceCommandNone
			if i.Status == nil {
				klog.Warningf("%s instance status is nil, will not update status", commandLogPrefix)
			} else {
				if i.Status.ErrorInfo != nil {
					klog.V(2).Infof("%s clearing previous error info", commandLogPrefix)
				}
				i.Status.ErrorInfo = nil
				if i.Status.State == cloudprovider.InstanceCreating {
					if i.PowerOn {
						klog.V(2).Infof("%s server created and powered on", i.Id)
					} else {
						klog.V(2).Infof("%s server created but not powered on yet", i.Id)
					}
				} else if i.Status.State == cloudprovider.InstanceDeleting {
					// instance deletion process - update state and complete the deletion process
					if commandCode == InstanceCommandPoweroff {
						// poweroff completed - now we can continue to terminate the instance
						if powerOffOnScaleDown {
							// poweroff is the last step - we clear the status to indicate instance is not managed by CA anymore
							// but we don't delete it because it might be powered on later
							klog.V(2).Infof("Instance %s powered off for scale down - clearing status", serverName)
							if err := clearAutoscalerMetadataFromNode(kubeClient, serverName); err != nil {
								klog.Errorf("Instance %s powered off for scale down - failed to clear Kubernetes node metadata: %v", serverName, err)
							}
							i.Status = nil
						} else {
							// continue to terminate the instance
							// the terminate function will set the appropriate status and command ID
							// and update ErrorInfo if needed
							klog.V(2).Infof("Instance %s powered off for scale down - continuing to terminate", serverName)
							_ = i.terminate(client, providerIDPrefix)
						}
					} else if commandCode == InstanceCommandTerminate {
						// terminate completed - in this case we delete the instance
						klog.V(2).Infof("Instance %s terminated for scale down - marking for deletion", serverName)
						return true
					} else {
						klog.Warningf("Instance %s command completed but unexpected command code: %v", serverName, commandCode)
					}
				} else {
					klog.Warningf("Instance %s command completed but instance is in unexpected state: %v", serverName, i.Status.State)
				}
			}
		}
	}
	return false
}

const (
	kubeNodeUpdateMaxRetryDeadline      = 5 * time.Second
	kubeNodeUpdateConflictRetryInterval = 750 * time.Millisecond
)

func clearAutoscalerMetadataFromNode(kubeClient kubernetes.Interface, nodeName string) error {
	if kubeClient == nil || nodeName == "" {
		return nil
	}

	ctx := context.Background()
	retryDeadline := time.Now().Add(kubeNodeUpdateMaxRetryDeadline)

	for {
		node, err := kubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		updated := node.DeepCopy()
		changed := false

		if updated.Spec.Unschedulable {
			updated.Spec.Unschedulable = false
			changed = true
		}

		if len(updated.Spec.Taints) > 0 {
			newTaints := make([]apiv1.Taint, 0, len(updated.Spec.Taints))
			for _, taint := range updated.Spec.Taints {
				if taint.Key == taints.ToBeDeletedTaint || taint.Key == taints.DeletionCandidateTaint().Key {
					changed = true
					continue
				}
				newTaints = append(newTaints, taint)
			}
			updated.Spec.Taints = newTaints
		}

		if !changed {
			return nil
		}

		_, err = kubeClient.CoreV1().Nodes().Update(ctx, updated, metav1.UpdateOptions{})
		if err != nil && apierrors.IsConflict(err) && time.Now().Before(retryDeadline) {
			time.Sleep(kubeNodeUpdateConflictRetryInterval)
			continue
		}
		return err
	}
}
