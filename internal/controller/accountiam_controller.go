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
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	GlobalRealmValue    string
	GlobalAccountIDP    string
	GlobalAccountAud    string
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
//+kubebuilder:rbac:groups=liberty.websphere.ibm.com,resources=webspherelibertyapplications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings;roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=use

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

	if err := r.verifyPrereq(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileOperandResources(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AccountIAMReconciler) verifyPrereq(ctx context.Context, instance *operatorv1alpha1.AccountIAM) error {
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

	dbPass, err := generatePassword()
	if err != nil {
		return err
	}
	storedPass := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: "account-im-db-password"}, storedPass); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		storedPass.Name = "account-im-db-password"
		storedPass.Namespace = instance.Namespace
		storedPass.Data = make(map[string][]byte, 1)
		storedPass.Data["password"] = dbPass
		if err := r.Create(ctx, storedPass); err != nil {
			return err
		}
	}
	if _, ok := storedPass.Data["password"]; !ok {
		return errors.New("account-im-db-password secret is missing password")
	}

	bootSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: "account-iam-bootstrap"}, bootSecret); err != nil {
		return err
	}

	bootstrapConverter, err := yaml.Marshal(bootSecret.Data)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(bootstrapConverter, &BootstrapData); err != nil {
		return err
	}

	BootstrapData.PGPassword = string(storedPass.Data["password"])

	if err := r.cleanJob(ctx, instance.Namespace); err != nil {
		return err
	}

	return nil
}

func generatePassword() ([]byte, error) {
	random := make([]byte, 20)
	_, err := rand.Read(random)
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(random)
	encoded2 := base64.StdEncoding.EncodeToString([]byte(encoded))
	result := []byte(encoded2)
	return result, nil
}

func (r *AccountIAMReconciler) cleanJob(ctx context.Context, ns string) error {
	object := &unstructured.Unstructured{}
	manifest := []byte(res.DB_MIGRATION_MCSPID)
	if err := yaml.Unmarshal(manifest, object); err != nil {
		return err
	}
	object.SetNamespace(ns)
	log.Log.Info("", "job object", object)
	background := metav1.DeletePropagationBackground
	if err := r.Delete(ctx, object, &client.DeleteOptions{
		PropagationPolicy: &background,
	}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	object = &unstructured.Unstructured{}
	manifest = []byte(res.DB_BOOTSTRAP_JOB)
	if err := yaml.Unmarshal(manifest, object); err != nil {
		return err
	}
	object.SetNamespace(ns)
	if err := r.Delete(ctx, object, &client.DeleteOptions{
		PropagationPolicy: &background,
	}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (r *AccountIAMReconciler) reconcileOperandResources(ctx context.Context, instance *operatorv1alpha1.AccountIAM) error {

	// TODO: will need to find a better place to initialize the database
	object := &unstructured.Unstructured{}
	manifest := []byte(res.DB_BOOTSTRAP_JOB)
	if err := yaml.Unmarshal(manifest, object); err != nil {
		return err
	}
	object.SetNamespace(instance.Namespace)
	if err := controllerutil.SetControllerReference(instance, object, r.Scheme); err != nil {
		return err
	}
	if err := r.createOrUpdate(ctx, object); err != nil {
		return err
	}

	// manifests which need data injected before creation
	tmpl := template.New("template bootstrap secrets")
	var tmplWriter bytes.Buffer
	for _, v := range res.APP_SECRETS {
		manifest := v
		tmplWriter.Reset()

		tmpl, err := tmpl.Parse(manifest)
		if err != nil {
			return err
		}
		if err := tmpl.Execute(&tmplWriter, BootstrapData); err != nil {
			return err
		}

		if err := yaml.Unmarshal(tmplWriter.Bytes(), object); err != nil {
			return err
		}
		object.SetNamespace(instance.Namespace)
		if err := controllerutil.SetControllerReference(instance, object, r.Scheme); err != nil {
			return err
		}
		if err := r.createOrUpdate(ctx, object); err != nil {
			return err
		}
	}

	// static manifests which do not change
	staticYamls := append(res.APP_STATIC_YAMLS, res.CertRotationYamls...)
	for _, v := range staticYamls {
		manifest := []byte(v)
		if err := yaml.Unmarshal(manifest, object); err != nil {
			return err
		}
		object.SetNamespace(instance.Namespace)
		if err := controllerutil.SetControllerReference(instance, object, r.Scheme); err != nil {
			return err
		}
		if err := r.createOrUpdate(ctx, object); err != nil {
			return err
		}
	}

	return nil
}

func (r *AccountIAMReconciler) createOrUpdate(ctx context.Context, obj *unstructured.Unstructured) error {
	// err := r.Update(ctx, obj)
	// if err != nil {
	// 	if !k8serrors.IsNotFound(err) {
	// 		return err
	// 	}
	// }
	// if err == nil {
	// 	return nil
	// }

	// only reachable if update DID see error IsNotFound
	err := r.Create(ctx, obj)
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	if err == nil {
		return nil
	}

	fromCluster := &unstructured.Unstructured{}
	fromCluster.SetKind(obj.GetKind())
	fromCluster.SetAPIVersion(obj.GetAPIVersion())
	if err := r.Get(ctx, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, fromCluster); err != nil {
		return err
	}
	obj.SetResourceVersion(fromCluster.GetResourceVersion())
	if err := r.Update(ctx, obj); err != nil {
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
