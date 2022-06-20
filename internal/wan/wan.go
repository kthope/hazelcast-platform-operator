package wan

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hazelcastcomv1alpha1 "github.com/hazelcast/hazelcast-platform-operator/api/v1alpha1"
	"github.com/hazelcast/hazelcast-platform-operator/internal/util"
)

func HazelcastWanConfigurationName(mapName string) string {
	return mapName + "-default"
}

func GetConfigMap(ctx context.Context, cli client.Client, wan *hazelcastcomv1alpha1.WanConfiguration) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "wan-target-" + wan.Spec.TargetClusterName + "-map-" + wan.Spec.MapResourceName,
			Namespace: wan.GetNamespace(),
		},
	}
	if err := util.CreateOrGet(ctx, cli, client.ObjectKeyFromObject(cm), cm); err != nil {
		return nil, err
	}
	return cm, nil
}

func IsInStore(ctx context.Context, cli client.Client, wan *hazelcastcomv1alpha1.WanConfiguration) (bool, error) {
	cm, err := GetConfigMap(ctx, cli, wan)
	if err != nil {
		return false, err
	}
	if cm.Data == nil {
		return false, nil
	}
	_, ok := cm.Data[wan.Name]
	return ok, nil
}

func GetFromStore(ctx context.Context, cli client.Client, wan *hazelcastcomv1alpha1.WanConfiguration) (string, error) {
	cm, err := GetConfigMap(ctx, cli, wan)
	if err != nil {
		return "", err
	}
	if cm.Data == nil {
		return "", fmt.Errorf("publisherId is not found: %w", err)
	}
	publisherId, ok := cm.Data[wan.Name]
	if !ok {
		return "", fmt.Errorf("publisherId is not found: %w", err)
	}
	return publisherId, nil
}

func SaveToStore(ctx context.Context, cli client.Client, wan *hazelcastcomv1alpha1.WanConfiguration) (string, error) {
	cm, err := GetConfigMap(ctx, cli, wan)
	if err != nil {
		return "", err
	}

	publisherName := wan.Name + "-" + rand.String(16)
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	cm.Data[wan.Name] = publisherName
	err = cli.Update(ctx, cm)
	if err != nil {
		return "", err
	}
	return publisherName, nil
}

func DeleteFromStore(ctx context.Context, cli client.Client, wan *hazelcastcomv1alpha1.WanConfiguration) (string, error) {
	cm, err := GetConfigMap(ctx, cli, wan)
	if err != nil {
		return "", err
	}

	name, ok := cm.Data[wan.Name]
	if !ok {
		return "", fmt.Errorf("failed to find publisher name in configmap")
	}

	delete(cm.Data, wan.Name)

	err = cli.Update(ctx, cm)
	if err != nil {
		return "", err
	}

	return name, nil
}

func GetHazelcastClientFromWan(ctx context.Context, cli client.Client, wan *hazelcastcomv1alpha1.WanConfiguration) (*hazelcast.Client, error) {
	m := &hazelcastcomv1alpha1.Map{}
	if err := cli.Get(ctx, types.NamespacedName{Name: wan.Spec.MapResourceName, Namespace: wan.Namespace}, m); err != nil {
		return nil, fmt.Errorf("failed to get Map CR from WanConfiguration: %w", err)
	}
	return GetHazelcastClient(m)
}

func GetHazelcastClient(m *hazelcastcomv1alpha1.Map) (*hazelcast.Client, error) {
	hzcl, ok := GetClient(types.NamespacedName{Name: m.Spec.HazelcastResourceName, Namespace: m.Namespace})
	if !ok {
		return nil, errors.NewInternalError(fmt.Errorf("cannot connect to the cluster for %s", m.Spec.HazelcastResourceName))
	}
	if hzcl.client == nil || !hzcl.client.Running() {
		return nil, fmt.Errorf("trying to connect to the cluster %s", m.Spec.HazelcastResourceName)
	}

	return hzcl.client, nil
}

func GetClient(ns types.NamespacedName) (client *hazelcast.Client, ok bool) {
	if v, ok := clients.Load(ns); ok {
		return v.(*Client), true
	}
	return nil, false
}
