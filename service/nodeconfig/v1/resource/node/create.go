package node

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"k8s.io/api/core/v1"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/node-operator/service/nodeconfig/v1/key"
)

const (
	// UnschedulablePatch is the JSON patch structure being applied to nodes using
	// a strategic merge patch in order to drain them.
	UnschedulablePatch = `{"spec":{"unschedulable":true}}`
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	draining, err := r.certsSearcher.SearchDraining(key.ClusterID(customObject))
	if err != nil {
		return microerror.Mask(err)
	}

	var restConfig *rest.Config
	{
		c := k8srestconfig.DefaultConfig()

		c.Logger = r.logger

		c.Address = key.ClusterAPIEndpoint(customObject)
		c.InCluster = false
		c.TLS.CAData = draining.NodeOperator.CA
		c.TLS.CrtData = draining.NodeOperator.Crt
		c.TLS.KeyData = draining.NodeOperator.Key

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	{
		n := key.NodeName(customObject)
		t := types.StrategicMergePatchType
		p := []byte(UnschedulablePatch)

		_, err := k8sClient.CoreV1().Nodes().Patch(n, t, p)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var customPods []v1.Pod
	var systemPods []v1.Pod
	{
		fieldSelector := fields.SelectorFromSet(fields.Set{
			"spec.nodeName": key.NodeName(customObject),
		})
		listOptions := apismetav1.ListOptions{
			FieldSelector: fieldSelector.String(),
		}
		podList, err := k8sClient.CoreV1().Pods(v1.NamespaceAll).List(listOptions)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, p := range podList.Items {
			if p.GetNamespace() == "kube-system" {
				systemPods = append(systemPods, p)
			} else {
				customPods = append(customPods, p)
			}
		}
	}

	for _, p := range customPods {
		err := k8sClient.CoreV1().Pods(p.GetNamespace()).Delete(p.GetName(), &apismetav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	for _, p := range systemPods {
		err := k8sClient.CoreV1().Pods(p.GetNamespace()).Delete(p.GetName(), &apismetav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		n := v1.NamespaceDefault

		c := v1alpha1.NodeConfigStatusCondition{
			Status: "True",
			Type:   "Drained",
		}
		customObject.Status.Conditions = append(customObject.Status.Conditions, c)

		_, err := r.g8sClient.CoreV1alpha1().NodeConfigs(n).Update(&customObject)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
