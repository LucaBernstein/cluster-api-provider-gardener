package v1alpha1

import (
	gardenercorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

var (
	// TODO(tobschli,LucaBernstein): Fine-grain scheme scopes.
	Scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))

	utilruntime.Must(clusterv1beta1.AddToScheme(Scheme))
	utilruntime.Must(AddToScheme(Scheme))

	utilruntime.Must(gardenercorev1beta1.AddToScheme(Scheme))
	utilruntime.Must(kubernetes.AddGardenSchemeToScheme(Scheme))
	// +kubebuilder:scaffold:scheme
}
