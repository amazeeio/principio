/*

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
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	initv1alpha1 "github.com/amazeeio/principio/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// InitConfigReconciler reconciles a InitConfig object
type InitConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=init.amazee.io,resources=initconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=init.amazee.io,resources=initconfigs/status,verbs=get;update;patch

// all the things
// +kubebuilder:rbac:groups=*,resources=*,verbs=*

func (r *InitConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	opLog := r.Log.WithValues("initconfig", req.NamespacedName)

	var namespace corev1.Namespace
	if err := r.Get(ctx, req.NamespacedName, &namespace); err != nil {
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// grab all the configs in the cluster (or do we limit this to the controllers namespace only?)
	initConfigs := &initv1alpha1.InitConfigList{}
	if err := r.List(ctx, initConfigs); err != nil {
		opLog.Info(fmt.Sprintf("Unable to list configs in the cluster, there may be none or something went wrong: %v", err))
		return ctrl.Result{}, nil
	}
	// iterate over them
	for _, config := range initConfigs.Items {
		// always run config unless we have something in the labels that says otherwise
		// @TODO: probably this section needs redoing
		runConfig := true
		// for all the labels in our config "selector" :D check if the operator matches
		for _, label := range config.Spec.InitLabels {
			operatorCheck(&runConfig, label["operator"], namespace.ObjectMeta.Labels, label["key"])
		} // @TODO: probably this section needs redoing

		// if we get the go ahead, run the config
		if runConfig {
			// go through our items and create the objects
			for _, item := range config.Spec.InitItems {
				unstructObj := item.DeepCopy()
				unstructObj.SetNamespace(namespace.Name)

				// since this is only acting on a namespace creation, but on initial reconciliation it could get here
				// check the object doesn't already exist, we don't want to modify something that might already exist (or do we?)
				if err := r.Get(ctx, types.NamespacedName{
					Name:      unstructObj.GetName(),
					Namespace: unstructObj.GetNamespace(),
				}, unstructObj); err != nil {
					opLog.Info(fmt.Sprintf("Doesn't exist, creating %s/%s in %s", unstructObj.GetKind(), unstructObj.GetName(), unstructObj.GetNamespace()))
					// create it if it doesn't exist
					if err := r.Create(ctx, unstructObj); err != nil {
						return ctrl.Result{}, err
					}
				} else {
					// just mention it if it does
					opLog.Info(fmt.Sprintf("Resource %s/%s exists in %s", unstructObj.GetKind(), unstructObj.GetName(), unstructObj.GetNamespace()))
				}
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the watch on the namespace resource with an event filter (see controller_predicates.go)
func (r *InitConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		WithEventFilter(NamespacePredicates{}).
		Complete(r)
}

// will ignore not found errors
func ignoreNotFound(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

// check the "operator" logic
// @TODO: this could/should? be replaced with something better
func operatorCheck(runConfig *bool, operator string, labels map[string]string, key string) {
	if operator == "DoesNotExist" {
		if _, ok := labels[key]; ok {
			*runConfig = false
		}
	}
	if operator == "Exists" {
		if _, ok := labels[key]; ok {
			*runConfig = true
		}
	}
}
