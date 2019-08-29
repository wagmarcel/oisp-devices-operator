package oispdevicesmanager

import (
	"context"
	//"fmt"
	//"encoding/json"
	generror "errors"

	oispv1alpha1 "github.com/oisp-devices-operator/pkg/apis/oisp/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/labels"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	//"github.com/operator-framework/operator-sdk/pkg/k8sutil"
)

var log = logf.Log.WithName("controller_oispdevicesmanager")

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
		labelNodes: make(map[string] *labelNode)}
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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner OispDevicesManager
	err = c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileOispDevicesManager implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileOispDevicesManager{}

// labelNodes: list of nodes with a specific label
type labelNode struct{
	value string
	nodes *corev1.NodeList
}

// ReconcileOispDevicesManager reconciles a OispDevicesManager object
type ReconcileOispDevicesManager struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	labelNodes map[string]*labelNode
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

	// if not a node (i.e. has namespace)
	// fetch the OispDevicesManager instance
	if (request.Namespace != "") {
		instance := &oispv1alpha1.OispDevicesManager{}
		reqLogger.Info("Marcel: before instance ", "instance", instance)
		err := r.client.Get(context.TODO(), request.NamespacedName, instance)
		reqLogger.Info("Marcel: process devicemanager ", "instance", instance, "Error", err)
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
		if (instance.Status.Phase == "") {
			instance.Status.Phase = oispv1alpha1.PhasePending
		}

		// Initialize and set to RUNNING if possible
		// If there is not label key, go in error state.
		if (instance.Spec.WatchLabelKey == "") {
			instance.Status.Phase = oispv1alpha1.PhaseError
			_ = r.client.Status().Update(context.TODO(), instance)
			return reconcile.Result{}, generror.New("No label key given")
		}
		// if there is label key and value, get initial list of interesting nodes
		if (instance.Spec.WatchLabelValue != "") {
			reqLogger.Info("Marcel912: adding watchLabel", "watchLabelKey", instance.Spec.WatchLabelKey, "watchLabelValue", instance.Spec.WatchLabelValue)
			nodes, err := r.getNodesWithLabel(instance.Spec.WatchLabelKey, instance.Spec.WatchLabelValue)
			reqLogger.Info("Nodes found", "nodes", nodes.Items, "err", err)
			if err != nil { // if fetching nodes was not successful, try it later again
				return reconcile.Result{}, err
			}
			r.labelNodes[instance.Spec.WatchLabelKey] = &labelNode{value: instance.Spec.WatchLabelValue, nodes: nodes}
			instance.Status.Phase = oispv1alpha1.PhaseRunning
		} else {
			instance.Status.Phase = oispv1alpha1.PhaseError
			_ = r.client.Status().Update(context.TODO(), instance)
			return reconcile.Result{}, generror.New("No label value given")
		}

		reqLogger.Info("Marcel123 now updating state")
		// Update State
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else { //node given, update all nodes from the labelNodes list
		//getNodesWithLabel()
	}
	reqLogger.Info("Job done.")
	return reconcile.Result{}, nil
}

func (r *ReconcileOispDevicesManager) getNodesWithLabel(key string, value string) (*corev1.NodeList, error) {
	//opts := &client.ListOptions{Namespace: "kube-system"}
	//opts.SetLabelSelector(fmt.Sprintf("app=%s", "name"))
	//opts.InNamespace("kube-system")
	//opts := &client.ListOptions{}
	sel := labels.Set{key: value};
	opts := &client.ListOptions{LabelSelector: sel.AsSelector()}
	nodes := &corev1.NodeList{}
	//oispDevicesManagerList := &oispv1alpha1.OispDevicesManagerList{}
	err := r.client.List(context.TODO(), opts, nodes)
	//nodes := &corev1.NodeList{}
	//sel := labels.Set{"kubernetes.io/arch": "amd64"};
	//options := client.ListOptions{LabelSelector: sel.AsSelector()}
	//options := client.ListOptions{}
	//err := r.client.List(context.TODO(), &options, nodes)
	//log.Info("Received", "nodes", nodes.Items)
	return nodes, err
}
