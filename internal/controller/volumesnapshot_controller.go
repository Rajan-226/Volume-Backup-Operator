/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	backupv1 "github.com/Rajan-226/volume-snapshot-operator/api/v1"
	snapv1 "github.com/kubernetes-csi/external-snapshotter/client/v7/apis/volumesnapshot/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	PvcSelectorLabelKey = "pvcselectorlabel"
)

// VolumeSnapshotReconciler reconciles a VolumeSnapshot object
type VolumeSnapshotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=backup.infra.dev=volumesnapshots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=backup.infra.dev=volumesnapshots/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=backup.infra.dev=volumesnapshots/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VolumeSnapshot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *VolumeSnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	fmt.Printf("Received a reconcile call for %s!!!\n\n", req.NamespacedName)
	vsObject := &backupv1.VolumeSnapshot{}
	err := r.Get(ctx, req.NamespacedName, vsObject)
	if err != nil {
		fmt.Printf("Volume snapshot %s got deleted\n\n", req.NamespacedName)
		fmt.Println(err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	pvcList, err := r.getPVCList(ctx, vsObject)
	if err != nil {
		fmt.Printf("Could not get Pvc list due to %s\n\n", err.Error())
		return ctrl.Result{}, nil
	}

	for _, pvc := range pvcList.Items {
		fmt.Printf("Creating Snapshot for PVC %s\n\n", pvc.Name)
		err = r.createSnapshot(ctx, pvc)
		if err != nil {
			fmt.Printf("Not able to create snapshot for PVC %s due to %s\n\n", pvc.Name, err.Error())
			return ctrl.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *VolumeSnapshotReconciler) getPVCList(ctx context.Context, vsObject *backupv1.VolumeSnapshot) (*corev1.PersistentVolumeClaimList, error) {
	pvcList := &corev1.PersistentVolumeClaimList{}
	filters := []client.ListOption{
		client.MatchingLabels{
			PvcSelectorLabelKey: vsObject.Spec.PvcSelectorLabel,
		},
		client.InNamespace(vsObject.Namespace),
	}

	err := r.Client.List(ctx, pvcList, filters...)

	if err != nil {
		return nil, err
	}

	return pvcList, nil
}

func (r *VolumeSnapshotReconciler) createSnapshot(ctx context.Context, pvc corev1.PersistentVolumeClaim) error {

	// if present, err := r.isPresent(ctx, pvc); err != nil {
	// 	return fmt.Errorf(err.Error() + " while checking if snapshot is already present")
	// } else if present {
	// 	fmt.Printf("Snapshot for PVC %s already exists\n\n", pvc.Name)
	// 	return nil
	// }

	snpClass := "test-snapclass"

	snp := &snapv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvc.Name + "-" + strconv.FormatInt(time.Now().UnixMilli(), 10),
			Namespace: pvc.Namespace,
			Labels: map[string]string{
				"pvc_name": pvc.Name,
			},
		},
		Spec: snapv1.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &snpClass,
			Source: snapv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvc.Name,
			},
		},
	}

	err := r.Client.Create(ctx, snp)

	if err != nil {
		return fmt.Errorf(err.Error() + " while creating snapshot")
	}

	fmt.Printf("Created Snapshot for PVC %s successully\n\n", pvc.Name)
	return nil
}

func (r *VolumeSnapshotReconciler) isPresent(ctx context.Context, pvc corev1.PersistentVolumeClaim) (bool, error) {
	/*
		User will create a new object for same pvc, so snapshots for same pvc can be possible, but should have unique name
		on requeue, we should not create a snapshot again for the pvcs we were able to create earlier

		Status{
			pendingPVCs
			status
		}


		for the first call of reconcile, status field would be a empty string
		list all the pvcs for which snapshots needs to be created and persist it in pendingPVCs array
		set the status to init


		Use a pendingPVCs array in status to identify for what pvcs we need to create snapshots

		now keep polling from that array, try to create snapshot
		success - remove from pendingPVCs array
		failure - put back in pendingPVCs array

		set requeue to true in the end if pendingPVCs is not empty and status to pending


		On requeue attempt:
			status would not be empty string, get pvcs from pendingPVCs array and process them


		User will create a new object for same pvc, so snapshots for same pvc can be possible, but should have unique name - done
		on requeue, we should not create a snapshot again for the pvcs we were able to create earlier - done
	*/
	snp := &snapv1.VolumeSnapshotList{}

	filters := []client.ListOption{
		client.MatchingLabels{
			"pvc_name": pvc.Name,
		},
		client.InNamespace(pvc.Namespace),
	}

	err := r.Client.List(ctx, snp, filters...)

	if err == nil {
		return len(snp.Items) > 0, nil
	}

	return false, client.IgnoreNotFound(err)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeSnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&backupv1.VolumeSnapshot{}).
		Complete(r)
}
