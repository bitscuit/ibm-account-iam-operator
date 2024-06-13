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
	"errors"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/IBM/ibm-account-iam-operator/api/v1alpha1"
	olmapi "github.com/operator-framework/api/pkg/operators/v1"
)

// AccountIAMReconciler reconciles a AccountIAM object
type AccountIAMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operator.ibm.com,resources=accountiams,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.ibm.com,resources=accountiams/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.ibm.com,resources=accountiams/finalizers,verbs=update
//+kubebuilder:rbac:groups=operators.coreos.com,resources=operatorgroups,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AccountIAM object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *AccountIAMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("#Reconciling AccountIAM using fid image")

	// pre-req check: edb, websphere
	if err := r.verifyPrereq(ctx); err != nil {
		return ctrl.Result{}, err
	}

	// run version.reconcile
	// reconcile resources in account-iam-automation/scripts/fyre/out/manifests.yaml
	// load configuration from secret-bootstrap
	// load configuration from configmap-bootstrap
	// create secrets and configmaps with data from bootstrap configuration
	// create WLA CR

	// reconcile resources in account-iam-automation/scripts/fyre/bedrock/iam-cert-rotation.yaml

	// what resources have no dependencies?
	// what dependencies needed for other resources
	// set OwnerReference to this CR
	// can have multiple runtimes?

	return ctrl.Result{}, nil
}

func (r *AccountIAMReconciler) verifyPrereq(ctx context.Context) error {
	og := &olmapi.OperatorGroupList{}
	err := r.Client.List(ctx, og, &client.ListOptions{
		Namespace: os.Getenv("WATCH_NAMESPACE"),
	})
	if err != nil {
		return err
	}
	if len(og.Items) != 1 {
		return errors.New("there should be exactly one OperatorGroup in this namespace")
	}
	providedApis := og.Items[0].Annotations["olm.providedAPIs"]

	if !strings.Contains(providedApis, "postgresql") {
		return errors.New("missing EDB prereq")
	}

	if !strings.Contains(providedApis, "WebSphereLibertyApplication") {
		return errors.New("missing Websphere Liberty prereq")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AccountIAMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.AccountIAM{}).
		Complete(r)
}
