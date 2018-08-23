package varnishservice

import (
	"context"
	"errors"

	icmv1alpha1 "icm-varnish-k8s-operator/operator/controller/pkg/apis/icm/v1alpha1"
	"icm-varnish-k8s-operator/operator/controller/pkg/config"
	"icm-varnish-k8s-operator/operator/controller/pkg/logger"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new VarnishService Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr, config.GlobalConf))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, globalConf *config.Config) reconcile.Reconciler {
	return &ReconcileVarnishService{Client: mgr.GetClient(), scheme: mgr.GetScheme(), globalConf: globalConf}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("varnishservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to VarnishService
	err = c.Watch(&source.Kind{Type: &icmv1alpha1.VarnishService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by VarnishService - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &icmv1alpha1.VarnishService{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &icmv1alpha1.VarnishService{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileVarnishService{}

// ReconcileVarnishService reconciles a VarnishService object
type ReconcileVarnishService struct {
	client.Client
	scheme     *runtime.Scheme
	globalConf *config.Config
}

// Reconcile reads that state of the cluster for a VarnishService object and makes changes based on the state read
// and what is in the VarnishService.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=icm.ibm.com,resources=varnishservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=endpoints,verbs=list;watch
func (r *ReconcileVarnishService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the VarnishService instance
	instance := &icmv1alpha1.VarnishService{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, logger.RError(err, "could not read VarnishService")
	}

	serviceAccountName, err := r.reconcileServiceAccount(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	roleName, err := r.reconcileRole(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err = r.reconcileRoleBinding(instance, roleName, serviceAccountName); err != nil {
		return reconcile.Result{}, err
	}
	applicationPort, err := getApplicationPort(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err = r.reconcileBackendService(instance, applicationPort); err != nil {
		return reconcile.Result{}, err
	}
	deploymentSelector, err := r.reconcileDeployment(instance, serviceAccountName, applicationPort)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err = r.reconcileFrontendService(instance, applicationPort, deploymentSelector); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func getApplicationPort(instance *icmv1alpha1.VarnishService) (*v1.ServicePort, error) {
	if len(instance.Spec.Service.Ports) != 1 {
		err := errors.New("must specify exactly one port in service spec")
		logger.Error(err, "")
		return nil, err
	}
	return &instance.Spec.Service.Ports[0], nil
}
