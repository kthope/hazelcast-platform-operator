package hazelcast

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hazelcastcomv1alpha1 "github.com/hazelcast/hazelcast-platform-operator/api/v1alpha1"
	"github.com/hazelcast/hazelcast-platform-operator/controllers/protocol/codec"
)

// WanConfigurationReconciler reconciles a WanConfiguration object
type WanConfigurationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=hazelcast.com,resources=wanconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hazelcast.com,resources=wanconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hazelcast.com,resources=wanconfigurations/finalizers,verbs=update

func (r *WanConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	wan := &hazelcastcomv1alpha1.WanConfiguration{}
	if err := r.Client.Get(ctx, req.NamespacedName, wan); err != nil {
		return ctrl.Result{}, err
	}

	cli, err := r.getHazelcastClient(ctx, wan)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.applyWanConfiguration(ctx, cli, wan); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WanConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hazelcastcomv1alpha1.WanConfiguration{}).
		Complete(r)
}

func (r *WanConfigurationReconciler) getHazelcastClient(ctx context.Context, wan *hazelcastcomv1alpha1.WanConfiguration) (*hazelcast.Client, error) {
	m := &hazelcastcomv1alpha1.Map{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: wan.Spec.MapResourceName, Namespace: wan.Namespace}, m); err != nil {
		return nil, fmt.Errorf("failed to get Map CR from WanConfiguration: %w", err)
	}
	return GetHazelcastClient(m)
}

func (r *WanConfigurationReconciler) applyWanConfiguration(ctx context.Context, client *hazelcast.Client, wan *hazelcastcomv1alpha1.WanConfiguration) error {
	cliInt := hazelcast.NewClientInternal(client)

	req := codec.EncodeMCAddWanBatchPublisherConfigRequest(
		wan.Spec.MapResourceName+"-default",
		wan.Spec.TargetClusterName,
		wan.Name,
		wan.Spec.Endpoints,
		1024,
		128,
		200,
		10000,
		0,
		0,
	)
	for _, member := range cliInt.OrderedMembers() {
		_, err := cliInt.InvokeOnMember(ctx, req, member.UUID, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
