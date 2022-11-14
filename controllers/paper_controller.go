/*
Copyright 2022 Bernhard Aichinger.

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

package controllers

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	papermciov1 "github.com/baichinger/papermc-operator/api/v1"
	"github.com/baichinger/papermc-operator/pkg/papermc/reconciler"
)

var (
	noRequeue       = ctrl.Result{}
	shortishRequeue = ctrl.Result{RequeueAfter: 90 * time.Second}
	longishRequeue  = ctrl.Result{RequeueAfter: 2 * time.Hour}
)

// PaperController reconciles a Paper object
type PaperController struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (c *PaperController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&papermciov1.Paper{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(c)
}

// +kubebuilder:rbac:groups=papermc.io,resources=papers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=papermc.io,resources=papers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=v1,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (c *PaperController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("reconciliation event")

	p := &papermciov1.Paper{}
	if err := c.Get(ctx, req.NamespacedName, p); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Paper resource not found, ignoring, must be deleted")
			return noRequeue, nil
		}
		return noRequeue, err
	}

	r := reconciler.NewPaperReconciler(c.Client, c.Scheme, ctx, p)

	// initialize status (.status.conditions)
	if res := r.InitializeConditions(); res.Failed() {
		return noRequeue, res.GetError()
	} else if res.Updated() {
		logger.Info("reconciled initial status")
		return noRequeue, nil
	}

	// figure desired version/artifact details
	if res := r.ReconcileDesiredVersion(); res.Failed() {
		return noRequeue, res.GetError()
	} else if res.Updated() {
		logger.Info("reconciled desired version")
		return noRequeue, nil
	}

	// setup PVC for version/artifact
	if res := r.ReconcilePersistentVolumeClaimForDesiredVersion(); res.Failed() {
		return noRequeue, res.GetError()
	} else if res.Updated() {
		logger.Info("reconciled pvc for desired version")
		return noRequeue, nil
	}

	// download new version/artifact
	if res := r.ReconcileProvisionerForDesiredVersion(); res.Failed() {
		return noRequeue, res.GetError()
	} else if res.Updated() {
		logger.Info("reconciled provisioner for desired version")
		return noRequeue, nil
	}

	// setup PVC for instance
	if res := r.ReconcilePersistentVolumeClaimForPaperInstance(); res.Failed() {
		return noRequeue, res.GetError()
	} else if res.Updated() {
		logger.Info("reconciled pvc for instance")
		return noRequeue, nil
	}

	// run instance with desired version
	if res := r.ReconcilePaperInstanceForDesiredVersion(); res.Failed() {
		return noRequeue, res.GetError()
	} else if res.Updated() {
		logger.Info("reconciled pod for instance")
		return noRequeue, nil
	}

	// update status
	if res := r.ReconcileStatus(); res.Failed() {
		return noRequeue, res.GetError()
	} else if res.Updated() {
		logger.Info("reconciled status")
		return noRequeue, nil
	}

	logger.Info("reconciliation done")

	return longishRequeue, nil
}
