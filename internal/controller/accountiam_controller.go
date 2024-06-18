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
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"text/template"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/IBM/ibm-account-iam-operator/api/v1alpha1"
	res "github.com/IBM/ibm-account-iam-operator/internal/resources/yamls"
	"github.com/ghodss/yaml"
	olmapi "github.com/operator-framework/api/pkg/operators/v1"
)

// AccountIAMReconciler reconciles a AccountIAM object
type AccountIAMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type BootstrapSecret struct {
	Realm               string
	ClientID            string
	ClientSecret        string
	DiscoveryEndpoint   string
	UserValidationAPIV2 string
	PGPassword          string
	DefaultAUDValue     string
	DefaultIDPValue     string
	DefaultRealmValue   string
}

var BootstrapData BootstrapSecret

//+kubebuilder:rbac:groups=operator.ibm.com,resources=accountiams,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.ibm.com,resources=accountiams/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.ibm.com,resources=accountiams/finalizers,verbs=update
//+kubebuilder:rbac:groups=operators.coreos.com,resources=operatorgroups,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete

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

	instance := &operatorv1alpha1.AccountIAM{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("CR instance not found, don't requeue")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

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
	if err := r.reconcileOperandResources(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

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

func (r *AccountIAMReconciler) reconcileOperandResources(ctx context.Context, instance *operatorv1alpha1.AccountIAM) error {
	if err := r.reconcileNetworkPolicy(ctx, instance); err != nil {
		return err
	}

	if err := r.reconcileConfigmap(ctx, instance); err != nil {
		return err
	}

	if err := r.reconcileSecret(ctx, instance); err != nil {
		return err
	}

	if err := r.reconcileJob(ctx, instance); err != nil {
		return err
	}

	return nil
}

func (r *AccountIAMReconciler) reconcileNetworkPolicy(ctx context.Context, instance *operatorv1alpha1.AccountIAM) error {
	ingress := &netv1.NetworkPolicy{}
	if err := yaml.Unmarshal([]byte(res.INGRESS), ingress); err != nil {
		return err
	}
	ingress.Namespace = instance.Namespace
	if err := controllerutil.SetControllerReference(instance, ingress, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, ingress); err != nil {
		return err
	}

	egress := &netv1.NetworkPolicy{}
	if err := yaml.Unmarshal([]byte(res.EGRESS), egress); err != nil {
		return err
	}
	egress.Namespace = instance.Namespace
	if err := controllerutil.SetControllerReference(instance, egress, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, egress); err != nil {
		return err
	}

	return nil
}

func (r *AccountIAMReconciler) reconcileConfigmap(ctx context.Context, instance *operatorv1alpha1.AccountIAM) error {
	configmap := &corev1.ConfigMap{}
	if err := yaml.Unmarshal([]byte(res.CONFIG_ENV), configmap); err != nil {
		return err
	}
	configmap.Namespace = instance.Namespace
	if err := controllerutil.SetControllerReference(instance, configmap, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, configmap); err != nil {
		return err
	}

	jwtConfig := &corev1.ConfigMap{}
	if err := yaml.Unmarshal([]byte(res.CONFIG_JWT), jwtConfig); err != nil {
		return err
	}
	jwtConfig.Namespace = instance.Namespace
	if err := controllerutil.SetControllerReference(instance, jwtConfig, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, jwtConfig); err != nil {
		return err
	}

	return nil
}

func (r *AccountIAMReconciler) reconcileSecret(ctx context.Context, instance *operatorv1alpha1.AccountIAM) error {

	var tmplWriter bytes.Buffer
	tmpl, err := template.New("template bootstrap secrets").Parse(res.ClientAuth)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&tmplWriter, BootstrapData); err != nil {
		return err
	}

	secret := &corev1.Secret{}
	if err := yaml.Unmarshal(tmplWriter.Bytes(), secret); err != nil {
		return err
	}
	secret.Namespace = instance.Namespace
	if err := controllerutil.SetControllerReference(instance, secret, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, secret); err != nil {
		return err
	}

	tmplWriter.Reset()
	tmpl, err = tmpl.Parse(res.OKD_Auth)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&tmplWriter, BootstrapData); err != nil {
		return err
	}

	secret = &corev1.Secret{}
	if err := yaml.Unmarshal(tmplWriter.Bytes(), secret); err != nil {
		return err
	}
	secret.Namespace = instance.Namespace
	if err := controllerutil.SetControllerReference(instance, secret, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, secret); err != nil {
		return err
	}

	tmplWriter.Reset()
	tmpl, err = tmpl.Parse(res.DatabaseSecret)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&tmplWriter, BootstrapData); err != nil {
		return err
	}

	secret = &corev1.Secret{}
	if err := yaml.Unmarshal(tmplWriter.Bytes(), secret); err != nil {
		return err
	}
	secret.Namespace = instance.Namespace
	if err := controllerutil.SetControllerReference(instance, secret, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, secret); err != nil {
		return err
	}

	tmplWriter.Reset()
	tmpl, err = tmpl.Parse(res.MpConfig)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&tmplWriter, BootstrapData); err != nil {
		return err
	}

	secret = &corev1.Secret{}
	if err := yaml.Unmarshal(tmplWriter.Bytes(), secret); err != nil {
		return err
	}
	secret.Namespace = instance.Namespace
	if err := controllerutil.SetControllerReference(instance, secret, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, secret); err != nil {
		return err
	}

	return nil
}

func (r *AccountIAMReconciler) reconcileJob(ctx context.Context, instance *operatorv1alpha1.AccountIAM) error {

	sa := &corev1.ServiceAccount{}
	if err := yaml.Unmarshal([]byte(res.DB_MIGRATION_MCSPID_SA), sa); err != nil {
		return err
	}
	sa.Namespace = instance.Namespace
	if err := controllerutil.SetControllerReference(instance, sa, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, sa); err != nil {
		return err
	}

	job := &batchv1.Job{}
	if err := yaml.Unmarshal([]byte(res.DB_MIGRATION_MCSPID), job); err != nil {
		return err
	}
	job.Namespace = instance.Namespace
	if err := controllerutil.SetControllerReference(instance, job, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, job); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AccountIAMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.AccountIAM{}).
		Complete(r)
}
