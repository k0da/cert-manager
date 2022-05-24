package externaldns

import (
	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	"testing"
)

const fqdn = "_acme-challenge.test.example.com."

func TestFQDNToName(t *testing.T) {
	expected := "_acme-challenge-test-example-com"
	got := fqdnToName(fqdn)
	assert.Equal(t, got, expected)
}

func TestNewDNSProvider(t *testing.T) {
	scheme := runtime.NewScheme()
	client := fake.NewSimpleDynamicClient(scheme)

	p := NewProviderWithKubeClient(client, util.RecursiveNameservers)
	err := p.Present("", fqdn, "foobarbaz")
	assert.NoError(t, err)
	err = p.CleanUp("", fqdn, "")
	assert.NoError(t, err)

}
