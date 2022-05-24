package externaldns

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/external-dns/endpoint"
	//	"k8s.io/client-go/kubernetes/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

type DNSProvider struct {
	client           dynamic.Interface
	dns01Nameservers []string
	log              logr.Logger
}

const GroupName = "externaldns.k8s.io"
const GroupVersion = "v1alpha1"

var gvk = schema.GroupVersionKind{
	Group:   GroupName,
	Version: GroupVersion,
	Kind:    "DNSEndpoint",
}
var gvr = schema.GroupVersionResource{
	Group:    GroupName,
	Version:  GroupVersion,
	Resource: "dnsendpoint",
}

//func addKnownTypes(scheme *runtime.Scheme, groupVersion schema.GroupVersion) error {
//	scheme.AddKnownTypes(groupVersion,
//		&endpoint.DNSEndpoint{},
//		&endpoint.DNSEndpointList{},
//	)
//	metav1.AddToGroupVersion(scheme, groupVersion)
//	return nil
//}

func NewDNSProvider(dns01Nameservers []string) (*DNSProvider, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return NewProviderWithKubeClient(client, dns01Nameservers), nil
}

func fqdnToName(fqdn string) string {
	fqdn = strings.TrimSuffix(fqdn, ".")
	return strings.ReplaceAll(fqdn, ".", "-")
}

func NewProviderWithKubeClient(client dynamic.Interface, ns []string) *DNSProvider {
	return &DNSProvider{
		client:           client,
		dns01Nameservers: ns,
	}
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (c *DNSProvider) Present(_, fqdn, value string) error {
	obj := unstructured.Unstructured{}
	endpointSpec := endpoint.DNSEndpointSpec{}
	endpointSpec.Endpoints = append(endpointSpec.Endpoints, endpoint.NewEndpoint(fqdn, endpoint.RecordTypeTXT, value))
	obj.SetName(fqdnToName(fqdn))
	obj.SetGroupVersionKind(gvk)

	epSpec, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&endpointSpec)
	if err != nil {
		return err
	}
	if err := unstructured.SetNestedField(obj.Object, epSpec, "spec"); err != nil {
		return err
	}
	if _, err := c.client.Resource(gvr).Namespace("default").Create(context.TODO(), &obj, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil

}

// CleanUp removes the record matching the specified parameters.
func (c *DNSProvider) CleanUp(_, fqdn, _ string) error {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if err := c.client.Resource(gvr).Namespace("default").Delete(context.TODO(), fqdnToName(fqdn), deleteOptions); err != nil {
		return err
	}
	return nil
}
