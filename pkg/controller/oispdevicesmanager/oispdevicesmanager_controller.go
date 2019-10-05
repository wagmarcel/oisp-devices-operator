package oispdevicesmanager

import (
	"context"
	"io/ioutil"
	"reflect"
	//"fmt"
	//"encoding/json"
	generror "errors"

	oispv1alpha1 "github.com/oisp-devices-operator/pkg/apis/oisp/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	//"k8s.io/apimachinery/pkg/runtime/serializer/json"
	//"k8s.io/client-go/kubernetes/scheme"
	//"k8s.io/apimachinery/pkg/runtime/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/yaml"
	//"github.com/operator-framework/operator-sdk/pkg/k8sutil"
)

var log = logf.Log.WithName("controller_oispdevicesmanager")
const oispDevicesManagerFinalizer = "finalizer.oisp.net"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new OispDevicesManager Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileOispDevicesManager{client: mgr.GetClient(), scheme: mgr.GetScheme(),
		deviceManagerNodes: make(map[string] *deviceManagerNodes)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("oispdevicesmanager-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource OispDevicesManager
	err = c.Watch(&source.Kind{Type: &oispv1alpha1.OispDevicesManager{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Nodes are watched to find out which secondary resources should be scheduled
	err = c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileOispDevicesManager implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileOispDevicesManager{}

// labelNodes: list of nodes with a specific label
type deviceManagerNodes struct{
	deviceManager *oispv1alpha1.OispDevicesManager
	nodes map[string]*corev1.Node
}

// ReconcileOispDevicesManager reconciles a OispDevicesManager object
type ReconcileOispDevicesManager struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	deviceManagerNodes map[string]*deviceManagerNodes
}

// Reconcile reads that state of the cluster for a OispDevicesManager object and makes changes based on the state read
// and what is in the OispDevicesManager.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileOispDevicesManager) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling OispDevicesManager")

	// There could be two Kinds: (1) OispDeviceMananger and (2) Nodes. Nodes are detetctetd by missing namespace
	// fetch the OispDevicesManager instance
	if (request.Namespace != "") {
		// OispDeviceManager Status has 3 phases:
		// (1) PENDING - after initialization. Will go to RUNNING when the necessary WatchLabelKey/Value and Annotation Key is found
		// (2) RUNNING - a valid OispDeviceManager has been found and registered with operator
		// (3) ERROR   - error condition met
		instance := &oispv1alpha1.OispDevicesManager{}
		err := r.client.Get(context.TODO(), request.NamespacedName, instance)
		if err != nil {
			if errors.IsNotFound(err) {
				// Request object not found, could have been deleted after reconcile request.
				// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
				// Return and don't requeue
				return reconcile.Result{}, nil
			}
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}

		// Phase transition: When nothing defined go to PENDING
		if (instance.Status.Phase == "") {
			instance.Status.Phase = oispv1alpha1.PhasePending
		}

		// If there is not LabelKey => ERROR
		if (instance.Spec.WatchLabelKey == "") {
			instance.Status.Phase = oispv1alpha1.PhaseError
			_ = r.client.Status().Update(context.TODO(), instance)
			return reconcile.Result{}, generror.New("No label key given")
		}
		// If there is no Annotation Key => ERROR
		if (instance.Spec.WatchAnnotationKey == "") {
			instance.Status.Phase = oispv1alpha1.PhaseError
			_ = r.client.Status().Update(context.TODO(), instance)
			return reconcile.Result{}, generror.New("No Annotation key given")
		}

		//check whether resource is supposed to be deleted.
		isOispDevicesManagerMarkedToBeDeleted := instance.GetDeletionTimestamp() != nil
		if isOispDevicesManagerMarkedToBeDeleted {
			log.Info("Detected finalized oispDeviceManager")
			if contains(instance.GetFinalizers(), oispDevicesManagerFinalizer) {
				if err := r.finalizeOispDevicesManager(instance); err != nil {
					return reconcile.Result{}, err
				}

				// Remove OispDevicesManagerFinalizer. Once all finalizers have been
				// removed, the object will be deleted.
				instance.SetFinalizers(remove(instance.GetFinalizers(), oispDevicesManagerFinalizer))
				err := r.client.Update(context.TODO(), instance)
				if err != nil {
					return reconcile.Result{}, err
				}
			}
			return reconcile.Result{}, nil
		}
		// Add finalizer for this CR
		if !contains(instance.GetFinalizers(), oispDevicesManagerFinalizer) {
			if err := r.addFinalizer(instance); err != nil {
				return reconcile.Result{}, err
			}
		}

		//
		// At this point, labelkey and annkey should be defined properly
		// LabelValue can be empty
		// Get all Nodes which provid the right labels and annotations
		nodes, err := r.getNodesWithLabelAndSensorAnnotation(instance.Spec.WatchLabelKey, instance.Spec.WatchLabelValue, instance.Spec.WatchAnnotationKey)
		reqLogger.Info("Nodes found", "number of nodes", len(nodes), "err", err)
		if err != nil { // if fetching nodes was not successful, try it later again
			return reconcile.Result{}, err
		}
		// Register OispManager with Operator
		// This is needed to check later events from Nodes (which can apply to different operators)
		// global key is needed because the oispManager needs to be identified uniquely in a map
		r.deviceManagerNodes[getGlobalKey(instance.Spec.WatchLabelKey, instance.Spec.WatchLabelValue)] = &deviceManagerNodes{deviceManager: instance.DeepCopy(), nodes: nodes}
		if (len(nodes) > 0) { // check the deployments and create new when not existing
			rec, err := r.createDevicePluginDeployments(instance, nodes)
			if (err != nil) {
				return rec, err
			}
		}
		instance.Status.Phase = oispv1alpha1.PhaseRunning
		// Update State
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else { //node given, check whether node was updated
		node := corev1.Node{}
		err := r.client.Get(context.TODO(), request.NamespacedName, &node)
		log.Info("node update request", "node", node.Name, "err", err)
		// Check for every devices manager whether node is listed
		// If it is known, check whether relevant labels or annotations changed. If yes, create/update deployment and update node
		// If it is unknown, check whether relevant info is given. If yes, add node to deviceManager and create deployment
		for lk, _ := range r.deviceManagerNodes {
			oldNode := r.deviceManagerNodes[lk].nodes[node.Name]
			watchLabelKey := r.deviceManagerNodes[lk].deviceManager.Spec.WatchLabelKey
			watchAnnKey := r.deviceManagerNodes[lk].deviceManager.Spec.WatchAnnotationKey
			if (oldNode != nil) {
				labels := r.deviceManagerNodes[lk].nodes[node.Name].GetObjectMeta().GetLabels()
				annotations := r.deviceManagerNodes[lk].nodes[node.Name].GetObjectMeta().GetAnnotations()
				log.Info("Marcel: Node exists already. Checking updates for labels and annotations", "deviceManager", lk)
				if (reflect.DeepEqual(node.GetObjectMeta().GetLabels(), labels) &&
				reflect.DeepEqual(node.GetObjectMeta().GetAnnotations(), annotations)) {
					log.Info("Node labels and annotations did not change.")
					return reconcile.Result{}, nil //done, node did not change (from label and ann pov)
				}
				// Either Labels or Annotations changed. Investigate the changes and register node if appropriate
				if (node.GetObjectMeta().GetLabels()[watchLabelKey] != oldNode.GetObjectMeta().GetLabels()[watchLabelKey] ||
				node.GetObjectMeta().GetAnnotations()[watchAnnKey] != oldNode.GetObjectMeta().GetAnnotations()[watchAnnKey]) {
					// oldNode should be updated
					r.deviceManagerNodes[lk].nodes[node.Name] = node.DeepCopy()
					r.createOrUpdateDevicePluginDeployment(&node, r.deviceManagerNodes[lk].deviceManager)
				}
			} else {
				// Node is not known.
				if (node.GetObjectMeta().GetLabels()[watchLabelKey] != "" && node.GetObjectMeta().GetAnnotations()[watchAnnKey] != "") {
					r.deviceManagerNodes[lk].nodes[node.Name] = node.DeepCopy()
					r.createOrUpdateDevicePluginDeployment(&node, r.deviceManagerNodes[lk].deviceManager)
				}
			}
		}
	}
	reqLogger.Info("Job done.")
	return reconcile.Result{}, nil
}

func (r *ReconcileOispDevicesManager) getNodesWithLabelAndSensorAnnotation(key string, value string, annotationKey string) (map[string]*corev1.Node, error) {
	log.Info("getNodesWithLabelAndSensorAnnotation", "AnnotationKey", annotationKey)
	result := map[string]*corev1.Node{}
	sel := labels.Set{key: value};
	opts := &client.ListOptions{LabelSelector: sel.AsSelector()}
	nodes := &corev1.NodeList{}
	err := r.client.List(context.TODO(), opts, nodes)
	for _, element := range nodes.Items {
		name := element.GetObjectMeta().GetName()
		annotations := element.GetObjectMeta().GetAnnotations()
		for annk, _ := range annotations {
			if annk == annotationKey {
				log.Info("Adding current node to result", "name", name)
				result[name] = element.DeepCopy()
			}
		}
	}
	return result, err
}

func createDevicePluginDeploymentTemplate(node *corev1.Node, nameSpace string,
	nodeSelector map[string]string, template *corev1.PodTemplateSpec, annKey string) *appsv1.Deployment {
	log.Info("start createDEvicePluginDeploymentTemplate")
	temp := template.DeepCopy()
	basename := node.GetObjectMeta().GetName()
	labels := labelsForDevicePlugin(nodeSelector, basename)
	replicas := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      basename + "-devicePlugin-deployment",
			Namespace: nameSpace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: *temp,
		},
	}

	//name := node.GetObjectMeta().GetName() + "-oispdevices-deployment"
	//dep.GetObjectMeta().SetName(name)
	//dep.GetObjectMeta().SetNamespace(nameSpace)
	dep.Spec.Template.Spec.NodeSelector = nodeSelector
	config := node.GetObjectMeta().GetAnnotations()[annKey]
	configEnv := corev1.EnvVar{Name: "K8S_PLUGIN_CONFIG", Value: config}
	dep.Spec.Template.Spec.Containers[0].Env = append(dep.Spec.Template.Spec.Containers[0].Env, configEnv)
	return dep
}

func (r *ReconcileOispDevicesManager) createOrUpdateDevicePluginDeployment(node *corev1.Node, deviceManager *oispv1alpha1.OispDevicesManager) (reconcile.Result, error) {
	log.Info("start createOrUpdateDevicePluginDeployment")
	nameSpace := deviceManager.GetObjectMeta().GetNamespace()
	nodeSelector := map[string]string{deviceManager.Spec.WatchLabelKey: deviceManager.Spec.WatchLabelValue}
	//template := deserializeDeployment("deploy/templates/oisp-iot-plugin-deployment.yaml")
	dep := createDevicePluginDeploymentTemplate(node, nameSpace, nodeSelector, &deviceManager.Spec.PodTemplateSpec, deviceManager.Spec.WatchAnnotationKey)
	log.Info("Create Deployment", "dep", dep.Spec)
	if err := controllerutil.SetControllerReference(deviceManager, dep, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Name", dep.Name)
		if err := r.client.Create(context.TODO(), dep); err != nil {
			return reconcile.Result{}, err
		}
	} else { //deployment exists, update it
		log.Info("Updating Deployment", "Deployment.Name", dep.Name)
		if err := r.client.Update(context.TODO(), dep); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileOispDevicesManager) createDevicePluginDeployments(deviceManager *oispv1alpha1.OispDevicesManager,
	nodes map[string]*corev1.Node) (reconcile.Result, error) {
	log.Info("start createDevicePluginDeployment", "nodes", nodes)

	for _, node := range nodes {
		rec, err := r.createOrUpdateDevicePluginDeployment(node, deviceManager)
		if (err != nil) {
			return rec, err
		}
	}
	log.Info("end createDevicePluginDeployment")
	//controllerutil.SetControllerReference(deviceManager, dep, r.scheme)
	return reconcile.Result{}, nil
}

func (r *ReconcileOispDevicesManager) finalizeOispDevicesManager(m *oispv1alpha1.OispDevicesManager) error {
	//delete cached deviceManangerNodes struct
	key := getGlobalKey(m.Spec.WatchLabelKey, m.Spec.WatchLabelValue)
	delete(r.deviceManagerNodes, key)
	log.Info("Successfully finalized OispDevicesManager")
	return nil
}

func (r *ReconcileOispDevicesManager) addFinalizer(m *oispv1alpha1.OispDevicesManager) error {
	log.Info("Adding Finalizer for oispDevicesManager")
	m.SetFinalizers(append(m.GetFinalizers(), oispDevicesManagerFinalizer))

	// Update CR
	err := r.client.Update(context.TODO(), m)
	if err != nil {
		log.Error(err, "Failed to update OispDevicesMananger with finalizer")
		return err
	}
	return nil
}

func labelsForDevicePlugin(ns map[string]string, basename string) map[string]string {
	ns["app"] = basename + "-oisp-device-plugin"
	return ns
}

func deserializeDeployment(filename string) *appsv1.Deployment {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}
	dep := appsv1.Deployment{}
	err = yaml.Unmarshal(dat, &dep)

	if err != nil {
		return nil
	}
	return &dep
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func remove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}

func getGlobalKey(key string, value string) string {
	return key + "." + value
}
