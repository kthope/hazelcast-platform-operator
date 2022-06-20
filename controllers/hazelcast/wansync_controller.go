package hazelcast

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/hazelcast/hazelcast-go-client"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hazelcastcomv1alpha1 "github.com/hazelcast/hazelcast-platform-operator/api/v1alpha1"
	n "github.com/hazelcast/hazelcast-platform-operator/internal/naming"
	"github.com/hazelcast/hazelcast-platform-operator/internal/protocol/codec"
	wanutil "github.com/hazelcast/hazelcast-platform-operator/internal/wan"
)

// WanSyncReconciler reconciles a WanSync object
type WanSyncReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logr.Logger
}

//+kubebuilder:rbac:groups=hazelcast.com,resources=wansyncs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hazelcast.com,resources=wansyncs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hazelcast.com,resources=wansyncs/finalizers,verbs=update
//+kubebuilder:rbac:groups=hazelcast.com,resources=wanconfigurations,verbs=get;list

func (r *WanSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	wanSync := &hazelcastcomv1alpha1.WanSync{}
	if err := r.Get(ctx, req.NamespacedName, wanSync); err != nil {
		if kerrors.IsNotFound(err) {
			logger.V(2).Info("Could not find WanSync, it is probably already deleted")
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	if wanSync.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(wanSync, n.Finalizer) {
			controllerutil.AddFinalizer(wanSync, n.Finalizer)
			logger.Info("Adding finalizer to WanSync")
			if err := r.Update(ctx, wanSync); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(wanSync, n.Finalizer) {
			logger.Info("Deleting WanSync finalizer")
			controllerutil.RemoveFinalizer(wanSync, n.Finalizer)
			if err := r.Update(ctx, wanSync); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	wanConfig := &hazelcastcomv1alpha1.WanConfiguration{}
	if err := r.Get(ctx, types.NamespacedName{Name: wanSync.Spec.WanConfigurationResource, Namespace: req.Namespace}, wanConfig); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get WanConfiguration, %w", err)
	}

	hzClient, err := GetHazelcastClientFromWan(ctx, r.Client, wanConfig)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.applyWanSync(ctx, hzClient, wanConfig); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WanSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hazelcastcomv1alpha1.WanSync{}).
		Complete(r)
}

func (r *WanSyncReconciler) applyWanSync(ctx context.Context, client *hazelcast.Client, wan *hazelcastcomv1alpha1.WanConfiguration) error {
	publisherId, err := wanutil.GetFromStore(ctx, r.Client, wan)
	if err != nil {
		return err
	}

	cliInt := hazelcast.NewClientInternal(client)

	req := codec.EncodeMCWanSyncMapRequest(
		wanutil.HazelcastWanConfigurationName(wan.Spec.MapResourceName),
		publisherId,
		1,
		wan.Spec.MapResourceName,
	)

	for _, member := range cliInt.OrderedMembers() {
		if _, err := cliInt.InvokeOnMember(ctx, req, member.UUID, nil); err != nil {
			return err
		}
	}

	return nil
}
