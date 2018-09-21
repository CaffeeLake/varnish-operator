package varnishservice

import (
	icmapiv1alpha1 "icm-varnish-k8s-operator/pkg/apis/icm/v1alpha1"
)

func combinedLabels(instance *icmapiv1alpha1.VarnishService, componentName string) (m map[string]string) {
	inherited := inheritLabels(instance)
	generated := generateLabels(instance, componentName)

	m = make(map[string]string, len(inherited)+len(generated))
	for k, v := range inherited {
		m[k] = v
	}
	for k, v := range generated {
		m[k] = v
	}
	return
}

func inheritLabels(instance *icmapiv1alpha1.VarnishService) (m map[string]string) {
	m = make(map[string]string, len(instance.Labels))
	for k, v := range instance.Labels {
		m[k] = v
	}
	return
}

func generateLabels(instance *icmapiv1alpha1.VarnishService, componentName string) map[string]string {
	return map[string]string{
		"varnish-owner":     instance.Name,
		"varnish-component": componentName,
		"varnish-uid":       string(instance.UID),
	}
}
