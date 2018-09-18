package varnishservice

import (
	"context"
	icmapiv1alpha1 "icm-varnish-k8s-operator/pkg/apis/icm/v1alpha1"
	"icm-varnish-k8s-operator/pkg/compare"
	"icm-varnish-k8s-operator/pkg/logger"

	"k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileVarnishService) reconcileNoCachedService(instance, instanceStatus *icmapiv1alpha1.VarnishService, applicationPort *v1.ServicePort) (map[string]string, error) {
	selector := make(map[string]string, len(instance.Spec.Service.Selector))
	labels := make(map[string]string, len(instance.Spec.Service.Selector)+1)
	for k, v := range instance.Spec.Service.Selector {
		selector[k] = v
		labels[k] = v
	}
	labels["varnish-component"] = "nocached"
	noCachedService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-nocached",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Selector: selector,
			Ports:    []v1.ServicePort{*applicationPort},
		},
	}

	if err := r.reconcileServiceGeneric(instance, &instanceStatus.Status.Service.NoCached, noCachedService); err != nil {
		return labels, err
	}
	return labels, nil
}

func (r *ReconcileVarnishService) reconcileCachedService(instance, instanceStatus *icmapiv1alpha1.VarnishService, applicationPort *v1.ServicePort, varnishSelector map[string]string) error {
	cachedService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-cached",
			Namespace: instance.Namespace,
			Labels: map[string]string{
				"varnish-component": "cached",
			},
		},
	}
	instance.Spec.Service.DeepCopyInto(&cachedService.Spec)

	cachedService.Spec.Ports = []v1.ServicePort{
		{
			Name:       "http",
			Port:       applicationPort.Port,
			Protocol:   v1.ProtocolTCP,
			TargetPort: intstr.FromInt(r.globalConf.VarnishTargetPort),
		},
	}

	cachedService.Spec.Selector = varnishSelector

	if err := r.reconcileServiceGeneric(instance, &instanceStatus.Status.Service.Cached, cachedService); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileVarnishService) reconcileServiceGeneric(instance *icmapiv1alpha1.VarnishService, instanceServiceStatus *icmapiv1alpha1.VarnishServiceSingleServiceStatus, desired *v1.Service) error {
	// Set controller reference for desired object
	if err := controllerutil.SetControllerReference(instance, desired, r.scheme); err != nil {
		return logger.RError(err, "Cannot set controller reference for desired", "namespace", desired.Namespace, "name", desired.Name)
	}

	found := &v1.Service{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, found)
	// If the desired does not exist, create it
	// Else if there was a problem doing the GET, just return an error
	// Else if the desired exists, and it is different, update
	// Else no changes, do nothing
	if err != nil && kerrors.IsNotFound(err) {
		logger.Info("Creating Service", "config", desired)
		if err = r.Create(context.TODO(), desired); err != nil {
			return logger.RError(err, "Unable to create service", "name", desired.Name, "namespace", desired.Namespace)
		}
	} else if err != nil {
		return logger.RError(err, "Could not Get desired")
	} else if !compare.EqualService(desired, found) {
		logger.Info("Updating Service", "diff", compare.DiffService(desired, found), "name", desired.Name, "namespace", desired.Namespace)
		desired.Spec.ClusterIP = found.Spec.ClusterIP
		found.Spec = desired.Spec
		err = r.Update(context.TODO(), found)
		if err != nil {
			return logger.RError(err, "Unable to update desired")
		}
	} else {
		logger.V5Info("No updates for Service", "name", desired.Name)
	}

	*instanceServiceStatus = icmapiv1alpha1.VarnishServiceSingleServiceStatus{
		ServiceStatus: found.Status,
		IP:            found.Spec.ClusterIP,
	}
	return nil
}
