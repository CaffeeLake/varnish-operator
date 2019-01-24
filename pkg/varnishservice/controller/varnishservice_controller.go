package controller

import (
	"context"
	"errors"
	icmv1alpha1 "icm-varnish-k8s-operator/pkg/apis/icm/v1alpha1"
	"icm-varnish-k8s-operator/pkg/logger"
	"icm-varnish-k8s-operator/pkg/varnishservice/compare"
	"icm-varnish-k8s-operator/pkg/varnishservice/config"
	"icm-varnish-k8s-operator/pkg/varnishservice/pods"
	"icm-varnish-k8s-operator/pkg/varnishservice/webhooks"

	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"

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
func Add(mgr manager.Manager, cfg *config.Config, logr *logger.Logger) error {
	r := &ReconcileVarnishService{
		Client: mgr.GetClient(),
		config: cfg,
		logger: logr,
		scheme: mgr.GetScheme(),
		events: NewEventHandler(mgr.GetRecorder(EventRecorderNameVarnishService)),
	}

	if err := webhooks.InstallWebhooks(mgr, cfg, logr); err != nil {
		return err
	}

	// Create a new controller
	c, err := controller.New("varnishservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to VarnishService
	logr.Infow("Starting watch loop for VarnishService objects")
	err = c.Watch(&source.Kind{Type: &icmv1alpha1.VarnishService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes in the below resources:
	// Pods
	// ConfigMap
	// Deployment
	// Service
	// Role
	// RoleBinding
	// ServiceAccount

	podPredicate, err := pods.NewAnnotationsPredicate(map[string]string{labelVarnishComponent: componentNameVarnishes})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &v1.Pod{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(
			func(a handler.MapObject) []reconcile.Request {
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{
						Name:      a.Meta.GetLabels()[labelVarnishOwner],
						Namespace: a.Meta.GetNamespace(),
					}},
				}
			}),
	}, podPredicate)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &icmv1alpha1.VarnishService{},
	})
	if err != nil {
		return err
	}

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

	err = c.Watch(&source.Kind{Type: &rbacv1beta1.Role{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &icmv1alpha1.VarnishService{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1beta1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &icmv1alpha1.VarnishService{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1beta1.ClusterRole{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &icmv1alpha1.VarnishService{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &rbacv1beta1.ClusterRoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &icmv1alpha1.VarnishService{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
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
	config *config.Config
	logger *logger.Logger
	scheme *runtime.Scheme
	events *EventHandler
}

// Reconcile reads that state of the cluster for a VarnishService object and makes changes based on the state read
// and what is in the VarnishService.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=icm.ibm.com,resources=varnishservices,verbs=list;watch;create;update;delete
// +kubebuilder:rbac:groups=icm.ibm.com,resources=varnishservices/status,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups="",resources=services;serviceaccounts,verbs=list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=endpoints,verbs=list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=pods,verbs=list;get;watch;update
// +kubebuilder:rbac:groups="",resources=nodes,verbs=watch;list
// +kubebuilder:rbac:groups="admissionregistration.k8s.io",resources="validatingwebhookconfigurations;mutatingwebhookconfigurations",verbs=list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=list;watch;create;update;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=list;watch;create;update;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings;clusterroles;clusterrolebindings,verbs=list;watch;create;update;delete
func (r *ReconcileVarnishService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logr := r.logger.With("name", request.Name, "namespace", request.Namespace)

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
		return reconcile.Result{}, logr.RErrorw(err, "could not read VarnishService")
	}

	r.scheme.Default(instance)

	// For some reason, sometimes Kubernetes returns the object without apiVersion and kind
	// Since the code below relies on that values we set them manually if they are empty
	if instance.APIVersion == "" {
		instance.APIVersion = "icm.ibm.com/v1alpha1"
	}
	if instance.Kind == "" {
		instance.Kind = "VarnishService"
	}

	instanceStatus := &icmv1alpha1.VarnishService{}
	instance.ObjectMeta.DeepCopyInto(&instanceStatus.ObjectMeta)
	instance.Status.DeepCopyInto(&instanceStatus.Status)

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
	clusterRoleName, err := r.reconcileClusterRole(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err = r.reconcileClusterRoleBinding(instance, clusterRoleName, serviceAccountName); err != nil {
		return reconcile.Result{}, err
	}
	applicationPort, err := r.getApplicationPort(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	endpointSelector, err := r.reconcileNoCachedService(instance, instanceStatus, applicationPort)
	if err != nil {
		return reconcile.Result{}, err
	}
	varnishSelector, err := r.reconcileDeployment(instance, instanceStatus, serviceAccountName, applicationPort, endpointSelector)
	if err != nil {
		return reconcile.Result{}, err
	}
	_, err = r.reconcileConfigMap(varnishSelector, instance, instanceStatus)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err = r.reconcilePodDisruptionBudget(instance, varnishSelector); err != nil {
		return reconcile.Result{}, err
	}
	if err = r.reconcileCachedService(instance, instanceStatus, applicationPort, varnishSelector); err != nil {
		return reconcile.Result{}, err
	}

	if !compare.EqualVarnishServiceStatus(&instance.Status, &instanceStatus.Status) {
		logr.Infoc("Updating VarnishService Status", "diff", compare.DiffVarnishServiceStatus(&instance.Status, &instanceStatus.Status))
		if err = r.Status().Update(context.TODO(), instanceStatus); err != nil {
			return reconcile.Result{}, logr.RErrorw(err, "could not update VarnishService Status", "name", instance.Name, "namespace", instance.Namespace)
		}
	} else {
		logr.Debugw("No updates for VarnishService status")
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileVarnishService) getApplicationPort(instance *icmv1alpha1.VarnishService) (*v1.ServicePort, error) {
	logr := r.logger.With("name", instance.Name, "namespace", instance.Namespace)

	if len(instance.Spec.Service.Ports) != 1 {
		err := errors.New("must specify exactly one port in service spec")
		return nil, logr.RErrorw(err, "")
	}
	port := instance.Spec.Service.Ports[0]
	if port.TargetPort == (intstr.IntOrString{}) {
		port.TargetPort = intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: port.Port,
		}
	}
	return &port, nil
}
