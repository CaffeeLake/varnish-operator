package controller

import (
	"context"
	"sort"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ReconcileVarnish) getPodInfo(ctx context.Context, namespace string, labels labels.Selector, validPort intstr.IntOrString) ([]PodInfo, int32, error) {
	found := &v1.EndpointsList{}
	err := r.List(ctx, found, client.MatchingLabelsSelector{Selector: labels}, client.InNamespace(namespace))
	if err != nil {
		return nil, 0, errors.Wrapf(err, "could not retrieve backends from namespace %s with labels %s", namespace, labels.String())
	}

	if len(found.Items) == 0 {
		return nil, 0, errors.Errorf("no endpoints from namespace %s matching labels %s", namespace, labels.String())
	}

	var backendList []PodInfo
	var portNumber int32
	for _, endpoints := range found.Items {
		for _, endpoint := range endpoints.Subsets {
			for _, address := range append(endpoint.Addresses, endpoint.NotReadyAddresses...) {
				for _, port := range endpoint.Ports {
					if port.Port == validPort.IntVal || port.Name == validPort.StrVal {
						portNumber = port.Port
						nodeLabels, err := r.getNodeLabels(ctx, *address.NodeName)
						if err != nil {
							return nil, 0, errors.WithStack(err)
						}
						b := PodInfo{IP: address.IP, NodeLabels: nodeLabels, PodName: address.TargetRef.Name}
						backendList = append(backendList, b)
						break
					}
				}
			}
		}
	}

	// sort slices so they also look the same for the code using it
	// prevents cases when generated VCL files change only because of order changes in the slice
	sort.SliceStable(backendList, func(i, j int) bool {
		return backendList[i].IP < backendList[j].IP
	})
	return backendList, portNumber, nil
}
