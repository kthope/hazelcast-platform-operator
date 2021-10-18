package naming

// Labels and label values
const (
	// Finalizer name used by operator
	Finalizer = "hazelcast.com/finalizer"
	// LicenseDataKey is a key used in k8s secret that holds the Hazelcast license
	LicenseDataKey = "license-key"
	// ServicePerPodLabelName set to true when the service is a Service per pod
	ServicePerPodLabelName       = "hazelcast.com/service-per-pod"
	ServicePerPodCountAnnotation = "hazelcast.com/service-per-pod-count"
	ExposeExternallyAnnotation   = "hazelcast.com/expose-externally-member-access"

	// PodNameLabel label that represents the name of the pod in the StatefulSet
	PodNameLabel = "statefulset.kubernetes.io/pod-name"
	// ApplicationNameLabel label for the name of the application
	ApplicationNameLabel = "app.kubernetes.io/name"
	// ApplicationInstanceNameLabel label for a unique name identifying the instance of an application
	ApplicationInstanceNameLabel = "app.kubernetes.io/instance"
	// ApplicationManagedByLabel label for the tool being used to manage the operation of an application
	ApplicationManagedByLabel = "app.kubernetes.io/managed-by"

	LabelValueTrue  = "true"
	LabelValueFalse = "false"

	OperatorName      = "hazelcast-enterprise-operator"
	Hazelcast         = "hazelcast"
	HazelcastPortName = "hazelcast-port"

	// ManagementCenter MC name
	ManagementCenter = "management-center"
	// Mancenter MC short name
	Mancenter = "mancenter"
	// MancenterStorageName storage name for MC
	MancenterStorageName = Mancenter + "-storage"
)

// Environment variables used for Hazelcast cluster configuration
const (
	// KubernetesEnabled enable Kubernetes discovery
	KubernetesEnabled = "HZ_NETWORK_JOIN_KUBERNETES_ENABLED"
	// KubernetesServiceName used to scan only PODs connected to the given service
	KubernetesServiceName = "HZ_NETWORK_JOIN_KUBERNETES_SERVICENAME"
	// KubernetesNodeNameAsExternalAddress uses the node name to connect to a NodePort service instead of looking up the external IP using the API
	KubernetesNodeNameAsExternalAddress = "HZ_NETWORK_JOIN_KUBERNETES_USENODENAMEASEXTERNALADDRESS"
	// KubernetesServicePerPodLabel label name used to tag services that should form the Hazelcast cluster together
	KubernetesServicePerPodLabel = "HZ_NETWORK_JOIN_KUBERNETES_SERVICEPERPODLABELNAME"
	// KubernetesServicePerPodLabelValue label value used to tag services that should form the Hazelcast cluster together
	KubernetesServicePerPodLabelValue = "HZ_NETWORK_JOIN_KUBERNETES_SERVICEPERPODLABELVALUE"

	RESTEnabled            = "HZ_NETWORK_RESTAPI_ENABLED"
	RESTHealthCheckEnabled = "HZ_NETWORK_RESTAPI_ENDPOINTGROUPS_HEALTHCHECK_ENABLED"

	// HzLicenseKey License key for Hazelcast cluster
	HzLicenseKey = "HZ_LICENSEKEY"
	ClusterName  = "HZ_CLUSTERNAME"
)

// Environment variables used for Management Center configuration
const (
	// McLicenseKey License key for Management Center
	McLicenseKey = "MC_LICENSE_KEY"
	// McInitCmd init command for Management Center
	McInitCmd = "MC_INIT_CMD"
	JavaOpts  = "JAVA_OPTS"
)

// Hazelcast default configurations
const (
	// DefaultHzPort Hazelcast default port
	DefaultHzPort = 5701
	// DefaultClusterSize default number of members of Hazelcast cluster
	DefaultClusterSize = 3
	// HazelcastRepo image repository for Hazelcast
	HazelcastRepo = "hazelcast/hazelcast-enterprise"
	// HazelcastVersion version of Hazelcast image
	HazelcastVersion = "5.0-BETA-1"
)

// Management Center default configurations
const (
	// MCRepo image repository for Management Center
	MCRepo = "hazelcast/management-center"
	// MCVersion version of Management Center image
	MCVersion = "5.0-BETA-2"
)
