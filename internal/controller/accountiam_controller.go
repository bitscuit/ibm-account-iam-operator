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
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"

	ocproute "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatorv1alpha1 "github.com/IBM/ibm-user-management-operator/api/v1alpha1"
	"github.com/IBM/ibm-user-management-operator/internal/resources"
	res "github.com/IBM/ibm-user-management-operator/internal/resources/yamls"
	"github.com/ghodss/yaml"
	olmapi "github.com/operator-framework/api/pkg/operators/v1"
)

// AccountIAMReconciler reconciles a AccountIAM object
type AccountIAMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *rest.Config
}

type BootstrapSecret struct {
	Realm               string
	ClientID            string
	ClientSecret        string
	DiscoveryEndpoint   string
	PGPassword          string
	DefaultAUDValue     string
	DefaultIDPValue     string
	DefaultRealmValue   string
	SREMCSPGroupsToken  string
	GlobalRealmValue    string
	GlobalAccountIDP    string
	GlobalAccountAud    string
	UserValidationAPIV2 string
	IAMHOSTURL          string
	AccountIAMURL       string
	AccountIAMNamespace string
}

var BootstrapData BootstrapSecret

type UIBootstrapTemplate struct {
	Hostname                    string
	InstanceManagementHostname  string
	NodeEnv                     string
	CertDir                     string
	ConfigEnv                   string
	RedisHost                   string
	AccountAPI                  string
	ProductAPI                  string
	MeteringAPI                 string
	InstanceAPI                 string
	IssuerBaseURL               string
	SubscriptionAPI             string
	APIOAUTHTokenURL            string
	RedisCA                     string
	ClientID                    string
	ClientSecret                string
	DisableRedis                string
	SessionSecret               string
	DeploymentCloud             string
	IAMGlobalAPIKey             string
	APIOAUTHClientID            string
	APIOAUTHClientSecret        string
	IAMAPI                      string
	MyIBMURL                    string
	AWSProvisioningURL          string
	IBMCloudProvisioningURL     string
	ProductRegistrationUsername string
	ProductRegistrationPassword string
	IMIDMgmt                    string
	CSIDPURL                    string
	OnPremAccount               string
}

var UIBootstrapData UIBootstrapTemplate

//+kubebuilder:rbac:groups=operator.ibm.com,resources=accountiams,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.ibm.com,resources=accountiams/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.ibm.com,resources=accountiams/finalizers,verbs=update
//+kubebuilder:rbac:groups=operators.coreos.com,resources=operatorgroups,verbs=get;list;watch
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=liberty.websphere.ibm.com,resources=webspherelibertyapplications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings;roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=use
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers;certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

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

	klog.Infof("Reconciling AccountIAM using fid image")

	instance := &operatorv1alpha1.AccountIAM{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			klog.Infof("CR instance not found, don't requeue")
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

	// create im integration job
	if err := r.configIM(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileUI(ctx, instance); err != nil {
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

	existEDB, err := r.CheckCRD(resources.EDBAPIGroupVersion, resources.EDBClusterKind)
	if err != nil {
		return err
	}
	if !existEDB {
		return errors.New("missing EDB prereq")
	}

	existWebsphere, err := r.CheckCRD(resources.WebSphereAPIGroupVersion, resources.WebSphereKind)
	if err != nil {
		return err
	}
	if !existWebsphere {
		return errors.New("missing Websphere Liberty prereq")
	}

	// Generate PG password
	pgPassword, err := generatePassword()

	if err != nil {
		return err
	}
	klog.Infof("Generated PG password: %s", pgPassword)

	// Get cp-console route
	host, err := r.getHost(ctx, "cp-console", instance.Namespace)
	if err != nil {
		return err
	}
	klog.Infof("cp-console route host: %s", host)

	// Create bootstrap secret
	bootstrapsecret, err := r.initBootstrapData(ctx, instance.Namespace, pgPassword, host)
	if err != nil {
		return err
	}

	// Read the values from bootstrap secret and store in BootstrapData struct
	bootstrapConverter, err := yaml.Marshal(bootstrapsecret.Data)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(bootstrapConverter, &BootstrapData); err != nil {
		return err
	}

	if err := r.cleanJob(ctx, instance.Namespace); err != nil {
		return err
	}

	return nil
}

// Initialize BootstrapData with default values
func (r *AccountIAMReconciler) initBootstrapData(ctx context.Context, ns string, pg []byte, host string) (*corev1.Secret, error) {

	bootstrapsecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Name: "user-mgmt-bootstrap", Namespace: ns}, bootstrapsecret); err != nil {
		if k8serrors.IsNotFound(err) {

			klog.Info("Creating bootstrap secret with PG password")
			newsecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "user-mgmt-bootstrap",
					Namespace: ns,
				},
				Data: map[string][]byte{
					"Realm":               []byte("PrimaryRealm"),
					"ClientID":            []byte("mcsp-id"),
					"ClientSecret":        []byte("mcsp-secret"),
					"DiscoveryEndpoint":   []byte("https://" + host + "/idprovider/v1/auth/.well-known/openid-configuration"),
					"UserValidationAPIV2": []byte("https://openshift.default.svc/apis/user.openshift.io/v1/users/~"),
					"DefaultAUDValue":     []byte("mcsp-id"),
					"DefaultIDPValue":     []byte("https://" + host + "/idprovider/v1/auth"),
					"DefaultRealmValue":   []byte("PrimaryRealm"),
					"SREMCSPGroupsToken":  []byte("mcsp-im-integration-admin"),
					"GlobalRealmValue":    []byte("PrimaryRealm"),
					"GlobalAccountIDP":    []byte("https://" + host + "/idprovider/v1/auth"),
					"GlobalAccountAud":    []byte("mcsp-id"),
					"AccountIAMNamespace": []byte(ns),
					"PGPassword":          pg,
					"IAMHOSTURL":          []byte("https://" + host),
				},
				Type: corev1.SecretTypeOpaque,
			}

			if err := r.Create(ctx, newsecret); err != nil {
				if !k8serrors.IsAlreadyExists(err) {
					return nil, err
				}
			}
			return newsecret, nil
		} else {
			return nil, err
		}
	}
	return bootstrapsecret, nil
}

// Get the host of the route
func (r *AccountIAMReconciler) getHost(ctx context.Context, name string, ns string) (string, error) {
	// config := &corev1.ConfigMap{}
	// if err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, config); err != nil {
	// 	klog.Errorf("Failed to get route %s in namespace %s", name, ns)
	// 	return "", err
	// }
	// return config.Data["cluster_endpoint"], nil

	sourceRoute := &ocproute.Route{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, sourceRoute); err != nil {
		klog.Errorf("Failed to get route %s in namespace %s", name, ns)
		return "", err
	}
	return sourceRoute.Spec.Host, nil
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
	klog.Infof("Cleaning up job %s", object.GetName())
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

func (r *AccountIAMReconciler) initUIBootstrapData(ctx context.Context, instance *operatorv1alpha1.AccountIAM) error {
	clusterInfo := &corev1.ConfigMap{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: "ibmcloud-cluster-info"}, clusterInfo); err != nil {
		return err
	}
	if _, ok := clusterInfo.Data["cluster_kube_apiserver_host"]; !ok {
		return errors.New("configmap ibmcloud-cluster-info missing field 'cluster_kube_apiserver_host'")
	}
	cpconsole, ok := clusterInfo.Data["cluster_endpoint"]
	if !ok {
		return errors.New("configmap ibmcloud-cluster-info missing field 'cluster_endpoint'")
	}
	parsing := strings.Split(clusterInfo.Data["cluster_kube_apiserver_host"], ".")
	domain := strings.Join(parsing[1:], ".")
	log.Log.Info("", "domain: ", domain)

	UIBootstrapData = UIBootstrapTemplate{
		Hostname:                   concat("account-iam-ui-inst-main-", instance.Namespace, ".apps.", domain),
		InstanceManagementHostname: concat("account-iam-ui-inst-", instance.Namespace, ".apps.", domain),
		ClientID:                   BootstrapData.ClientID,
		ClientSecret:               BootstrapData.ClientSecret,
		IAMGlobalAPIKey:            "placeholder value until logic is added to retrieve from IM job",
		RedisHost:                  "placeholder value until onprem Redis integration done",
		RedisCA:                    "placeholder value until onprem Redis integration done",
		SessionSecret:              "placeholder value because we do not know what this is for yet",
		DeploymentCloud:            "IBM_CLOUD",
		IAMAPI:                     concat("https://account-iam-", instance.Namespace, ".apps.", domain),
		NodeEnv:                    "production",
		CertDir:                    "../../security",
		ConfigEnv:                  "dev",
		IssuerBaseURL:              concat(cpconsole, "/idprovider/v1/auth"),
		IMIDMgmt:                   cpconsole,
		CSIDPURL:                   concat(cpconsole, "/common-nav/identity-access/realms"),
		OnPremAccount:              "mcsp-im-intgn-account",
	}

	return nil
}

func (r *AccountIAMReconciler) reconcileOperandResources(ctx context.Context, instance *operatorv1alpha1.AccountIAM) error {

	// TODO: will need to find a better place to initialize the database
	klog.Infof("Creating DB Bootstrap Job")
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

	// Manifests which need data injected before creation
	klog.Infof("Creating MCSP secrets")
	if err := r.InjectData(ctx, instance, res.APP_SECRETS, BootstrapData); err != nil {
		return err
	}

	klog.Infof("Creating MCSP ConfigMaps")
	decodedData, err := r.decodeData(BootstrapData)
	if err != nil {
		return err
	}

	//print decodedData
	// reflectValue := reflect.ValueOf(decodedData)
	// reflectType := reflect.TypeOf(decodedData)

	// for i := 0; i < reflectType.NumField(); i++ {
	// 	fieldName := reflectType.Field(i).Name
	// 	fieldValue := reflectValue.Field(i).String()
	// 	klog.Infof("Field Name: %s, Field Value: %s", fieldName, fieldValue)
	// }

	if err := r.InjectData(ctx, instance, res.APP_CONFIGS, decodedData); err != nil {
		return err
	}

	// static manifests which do not change
	klog.Infof("Creating MCSP static yamls")
	staticYamls := append(res.APP_STATIC_YAMLS, res.CertRotationYamls...)
	staticYamls = append(staticYamls, res.StaticYamlsUI...)
	for _, v := range staticYamls {
		object := &unstructured.Unstructured{}
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

	// Temporary update issuer in platform-auth-idp configmap
	klog.Infof("Updating platform-auth-idp configmap")
	idpconfig := &corev1.ConfigMap{}
	if err := r.Get(ctx, client.ObjectKey{Name: "platform-auth-idp", Namespace: instance.Namespace}, idpconfig); err != nil {
		klog.Errorf("Failed to get configmap platform-auth-idp in namespace %s", instance.Namespace)
		return err
	}
	currentIssuer := idpconfig.Data["OIDC_ISSUER_URL"]
	idpValue := decodedData.DefaultIDPValue

	if currentIssuer == idpValue {
		klog.Infof("ConfigMap platform-auth-idp already has the desired value for OIDC_ISSUER_URL: %s", currentIssuer)
		return nil // Skip the update as the value is already set
	}

	idpconfig.Data["OIDC_ISSUER_URL"] = decodedData.DefaultIDPValue
	if err := r.Update(ctx, idpconfig); err != nil {
		klog.Errorf("Failed to update ConfigMap platform-auth-idp in namespace %s: %v", instance.Namespace, err)
		return err
	}

	// Delete the platform-auth-service and platform-identity-provider pod to restart it
	if err := r.restartAndCheckPod(ctx, instance.Namespace, "platform-auth-service"); err != nil {
		return err
	}
	klog.Infof(" platform-auth-service pod is ready")

	if err := r.restartAndCheckPod(ctx, instance.Namespace, "platform-identity-provider"); err != nil {
		return err
	}
	klog.Infof(" platform-identity-provider pod is ready")

	klog.Infof("MCSP operand resources created successfully")
	return nil
}

func (r *AccountIAMReconciler) configIM(ctx context.Context, instance *operatorv1alpha1.AccountIAM) error {

	host, err := r.getHost(ctx, "account-iam", instance.Namespace)
	if err != nil {
		return err
	}
	klog.Infof("account-iam route host: %s", host)

	mcspHost := "https://" + host
	encodedURL := base64.StdEncoding.EncodeToString([]byte(mcspHost))
	BootstrapData.AccountIAMURL = encodedURL

	klog.Infof("Creating IM Config Job")
	decodedData, err := r.decodeData(BootstrapData)
	if err != nil {
		return err
	}

	if err := r.InjectData(ctx, instance, res.IMConfigYamls, decodedData); err != nil {
		return err
	}

	return nil
}

func (r *AccountIAMReconciler) reconcileUI(ctx context.Context, instance *operatorv1alpha1.AccountIAM) error {
	if err := r.initUIBootstrapData(ctx, instance); err != nil {
		return err
	}

	// Manifests which need data injected before creation
	object := &unstructured.Unstructured{}
	tmpl := template.New("template for injecting data into YAMLs")
	var tmplWriter bytes.Buffer
	for _, v := range res.TemplateYamlsUI {
		manifest := v
		tmplWriter.Reset()

		tmpl, err := tmpl.Parse(manifest)
		if err != nil {
			return err
		}
		if err := tmpl.Execute(&tmplWriter, UIBootstrapData); err != nil {
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

	for _, v := range res.StaticYamlsUI {
		object := &unstructured.Unstructured{}
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

func (r *AccountIAMReconciler) InjectData(ctx context.Context, instance *operatorv1alpha1.AccountIAM, manifests []string, bootstrapData BootstrapSecret) error {

	var buffer bytes.Buffer

	// Loop through each secret manifest that requires data injection
	for _, manifest := range manifests {
		object := &unstructured.Unstructured{}
		buffer.Reset()

		// Parse the manifest template and execute it with the provided bootstrap data
		t := template.Must(template.New("template resrouces").Parse(manifest))
		if err := t.Execute(&buffer, bootstrapData); err != nil {
			return err
		}

		if err := yaml.Unmarshal(buffer.Bytes(), object); err != nil {
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

func (r *AccountIAMReconciler) decodeData(data BootstrapSecret) (BootstrapSecret, error) {
	val := reflect.ValueOf(&data).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String {
			decoded, err := base64.StdEncoding.DecodeString(field.String())
			if err != nil {
				return data, err
			}
			field.SetString(string(decoded))
		}
	}
	return data, nil
}

// CheckCRD returns true if the given crd is existent
func (r *AccountIAMReconciler) CheckCRD(apiGroupVersion string, kind string) (bool, error) {
	dc := discovery.NewDiscoveryClientForConfigOrDie(r.Config)
	exist, err := r.ResourceExists(dc, apiGroupVersion, kind)
	if err != nil {
		return false, err
	}
	if !exist {
		return false, nil
	}
	return true, nil
}

// ResourceExists returns true if the given resource kind exists
// in the given api groupversion
func (r *AccountIAMReconciler) ResourceExists(dc discovery.DiscoveryInterface, apiGroupVersion, kind string) (bool, error) {
	_, apiLists, err := dc.ServerGroupsAndResources()
	if err != nil {
		return false, err
	}
	for _, apiList := range apiLists {
		if apiList.GroupVersion == apiGroupVersion {
			for _, r := range apiList.APIResources {
				if r.Kind == kind {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// restart and check pod
func (r *AccountIAMReconciler) restartAndCheckPod(ctx context.Context, ns, label string) error {
	// restart platform-auth-service pod and wait for it to be ready

	pod, err := r.getPodName(ctx, ns, label)
	if err != nil {
		return err
	}

	podName := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod,
			Namespace: ns,
		},
	}
	if err := r.Delete(ctx, podName); err != nil {
		klog.Errorf("Failed to delete pod %s in namespace %s", label, ns)
		return err
	}

	time.Sleep(10 * time.Second)

	if err := r.waitForDeploymentReady(ctx, ns, label); err != nil {
		return err
	}

	return nil
}

func (r *AccountIAMReconciler) getPodName(ctx context.Context, namespace, label string) (string, error) {
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labels.Set{"app": label})

	if err := r.Client.List(ctx, podList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector,
	}); err != nil {
		return "", err
	}

	if len(podList.Items) == 0 {
		return "", fmt.Errorf("No pod found with label %s in namespace %s", labelSelector, namespace)
	}
	return podList.Items[0].Name, nil
}

func (r *AccountIAMReconciler) waitForDeploymentReady(ctx context.Context, ns, label string) error {

	return wait.PollImmediate(20*time.Second, 2*time.Minute, func() (bool, error) {
		deployment := &appsv1.Deployment{}
		if err := r.Get(ctx, client.ObjectKey{Name: label, Namespace: ns}, deployment); err != nil {
			klog.Errorf("Failed to get deployment %s in namespace %s: %v", label, ns, err)
			return false, err
		}

		klog.Infof("Waiting for Deployment %s to be ready...", label)

		desiredReplicas := *deployment.Spec.Replicas
		readyReplicas := deployment.Status.ReadyReplicas

		if readyReplicas == desiredReplicas {
			klog.Infof("Deployment %s is ready with %d/%d replicas.", label, readyReplicas, desiredReplicas)
			return true, nil
		}

		return false, nil
	})
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
