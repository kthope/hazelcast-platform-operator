package hazelcast

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/hazelcast/hazelcast-go-client"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	hazelcastcomv1alpha1 "github.com/hazelcast/hazelcast-platform-operator/api/v1alpha1"
	n "github.com/hazelcast/hazelcast-platform-operator/controllers/naming"
	"github.com/hazelcast/hazelcast-platform-operator/controllers/protocol/codec"
	codecTypes "github.com/hazelcast/hazelcast-platform-operator/controllers/protocol/types"
)

// WanConfigurationReconciler reconciles a WanConfiguration object
type WanConfigurationReconciler struct {
	client.Client
	logr.Logger
}

func NewWanConfigurationReconciler(client client.Client, log logr.Logger) *WanConfigurationReconciler {
	return &WanConfigurationReconciler{
		Client: client,
		Logger: log,
	}
}

//+kubebuilder:rbac:groups=hazelcast.com,resources=wanconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hazelcast.com,resources=wanconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hazelcast.com,resources=wanconfigurations/finalizers,verbs=update

func (r *WanConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.WithValues("name", req.Name, "namespace", req.NamespacedName)

	logger.Info("Fetching WAN configuration")
	wan := &hazelcastcomv1alpha1.WanConfiguration{}
	if err := r.Get(ctx, req.NamespacedName, wan); err != nil {
		if kerrors.IsNotFound(err) {
			logger.V(2).Info("Could not find WanConfiguration, it is probably already deleted")
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, err
		}
	}
	ctx = context.WithValue(ctx, "logger", logger)

	logger.Info("Getting Hazelcast client")
	cli, err := r.getHazelcastClient(ctx, wan)
	if err != nil {
		return ctrl.Result{}, err
	}

	if wan.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(wan, n.Finalizer) {
			controllerutil.AddFinalizer(wan, n.Finalizer)
			logger.Info("Adding finalizer")
			if err := r.Update(ctx, wan); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(wan, n.Finalizer) {
			logger.Info("Deleting WAN configuration")
			if err := r.stopWanConfiguration(ctx, cli, wan); err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("Deleting WAN configuration finalizer")
			controllerutil.RemoveFinalizer(wan, n.Finalizer)
			if err := r.Update(ctx, wan); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	logger.Info("Applying WAN configuration")
	if err := r.applyWanConfiguration(ctx, cli, wan); err != nil {
		if errors.Is(err, &noAdded{}) {
			logger.Info("WAN configuration is previously deleted, resuming it")
			if err := r.resumeWanConfiguration(ctx, cli, wan); err != nil {
				logger.Info("Failed to resume WAN configuration")
				return ctrl.Result{}, err
			}
			logger.Info("WAN configuration is resumed")
			return ctrl.Result{}, nil
		}
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
	req := &addBatchPublisherRequest{
		hazelcastWanConfigurationName(wan.Spec.MapResourceName),
		wan.Spec.TargetClusterName,
		wan.Name,
		wan.Spec.Endpoints,
		1024,
		128,
		200,
		10000,
		0,
		0,
	}
	resp, err := addBatchPublisherConfig(ctx, client, req)
	if err != nil {
		return fmt.Errorf("failed to apply WAN configuration: %w", err)
	}
	if len(resp.added) == 0 {
		return &noAdded{}
	}

	return nil
}

func (r *WanConfigurationReconciler) stopWanConfiguration(ctx context.Context, client *hazelcast.Client, wan *hazelcastcomv1alpha1.WanConfiguration) error {
	req := &changeWanStateRequest{
		name:        hazelcastWanConfigurationName(wan.Spec.MapResourceName),
		publisherId: wan.Name,
		state:       codecTypes.WanReplicationStatePaused,
	}
	return changeWanState(ctx, client, req)
}

func (r *WanConfigurationReconciler) resumeWanConfiguration(ctx context.Context, client *hazelcast.Client, wan *hazelcastcomv1alpha1.WanConfiguration) error {
	req := &changeWanStateRequest{
		name:        hazelcastWanConfigurationName(wan.Spec.MapResourceName),
		publisherId: wan.Name,
		state:       codecTypes.WanReplicationStateReplicating,
	}
	return changeWanState(ctx, client, req)
}

func hazelcastWanConfigurationName(mapName string) string {
	return mapName + "-default"
}

func defaultWanReference(mapName string) codecTypes.WanReplicationRef {
	return codecTypes.WanReplicationRef{
		Name:                 mapName + "-default",
		MergePolicyClassName: "PassThroughMergePolicy",
		Filters:              []string{},
	}
}

type addBatchPublisherRequest struct {
	name                  string
	targetCluster         string
	publisherId           string
	endpoints             string
	queueCapacity         int32
	batchSize             int32
	batchMaxDelayMillis   int32
	responseTimeoutMillis int32
	ackType               int32
	queueFullBehavior     int32
}

type addBatchPublisherResponse struct {
	added   []string
	ignored []string
}

func addBatchPublisherConfig(
	ctx context.Context,
	client *hazelcast.Client,
	request *addBatchPublisherRequest,
) (*addBatchPublisherResponse, error) {
	cliInt := hazelcast.NewClientInternal(client)

	logger := ctx.Value("logger").(logr.Logger)
	req := codec.EncodeMCAddWanBatchPublisherConfigRequest(
		request.name,
		request.targetCluster,
		request.publisherId,
		request.endpoints,
		1024,
		128,
		200,
		10000,
		0,
		0,
	)

	respsAdded := make([][]string, 0)
	respsIgnored := make([][]string, 0)
	for _, member := range cliInt.OrderedMembers() {
		resp, err := cliInt.InvokeOnMember(ctx, req, member.UUID, nil)
		if err != nil {
			return nil, err
		}
		added, ignored := codec.DecodeMCAddWanBatchPublisherConfigResponse(resp)
		logger.Info("addWanBatchPublisher decoded", "added", added, "ignored", ignored)
		respsAdded, respsIgnored = append(respsAdded, added), append(respsIgnored, ignored)
	}

	added, ignored := inferValidResponse(respsAdded, respsIgnored)
	return &addBatchPublisherResponse{
		added:   added,
		ignored: ignored,
	}, nil
}

func inferValidResponse(added [][]string, ignored [][]string) ([]string, []string) {
	mostAddedIndex := -1
	mostAddedLen := -1
	for i := 0; i < len(added); i++ {
		if len(added[i]) > mostAddedLen {
			mostAddedIndex = i
			mostAddedLen = len(added[i])
		}
	}
	return added[mostAddedIndex], ignored[mostAddedIndex]
}

type changeWanStateRequest struct {
	name        string
	publisherId string
	state       codecTypes.WanReplicationState
}

func changeWanState(ctx context.Context, client *hazelcast.Client, request *changeWanStateRequest) error {
	cliInt := hazelcast.NewClientInternal(client)

	req := codec.EncodeMCChangeWanReplicationStateRequest(
		request.name,
		request.publisherId,
		request.state,
	)

	for _, member := range cliInt.OrderedMembers() {
		_, err := cliInt.InvokeOnMember(ctx, req, member.UUID, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkResponsesEqual(resps [][]string) bool {
	if len(resps) <= 1 {
		return true
	}
	for i := 0; i < len(resps)-1; i++ {
		if len(resps[i]) != len(resps[i+1]) {
			return false
		}
		for j := 0; j < len(resps[i]); j++ {
			if resps[i][j] != resps[i+1][j] {
				return false
			}
		}
	}
	return true
}

type noAdded struct {
}

func (*noAdded) Error() string {
	return "no publisher is added"
}
