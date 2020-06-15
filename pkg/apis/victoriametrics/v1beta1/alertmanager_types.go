package v1beta1

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"strings"
)

// VmAlertmanager describes an VmAlertmanager cluster.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version",description="The version of VmAlertmanager"
// +kubebuilder:printcolumn:name="ReplicaCount",type="integer",JSONPath=".spec.ReplicaCount",description="The desired replicas number of Alertmanagers"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:path=vmalertmanagers,scope=Namespaced,shortName=vma,singular=vmalertmanager
type VmAlertmanager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the VmAlertmanager cluster. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Spec VmAlertmanagerSpec `json:"spec"`
	// Most recent observed status of the VmAlertmanager cluster. Read-only. Not
	// included when requesting from the apiserver, only from the Prometheus
	// Operator API itself. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	Status *VmAlertmanagerStatus `json:"status,omitempty"`
}

// VmAlertmanagerSpec is a specification of the desired behavior of the VmAlertmanager cluster. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type VmAlertmanagerSpec struct {
	// PodMetadata configures Labels and Annotations which are propagated to the alertmanager pods.
	PodMetadata *EmbeddedObjectMetadata `json:"podMetadata,omitempty"`
	// Image if specified has precedence over baseImage, tag and sha
	// combinations. Specifying the version is still necessary to ensure the
	// Prometheus Operator knows what version of VmAlertmanager is being
	// configured.
	Image *string `json:"image,omitempty"`
	// Version the cluster should be on.
	Version string `json:"version,omitempty"`
	// Tag of VmAlertmanager container image to be deployed. Defaults to the value of `version`.
	// Version is ignored if Tag is set.
	Tag string `json:"tag,omitempty"`
	// SHA of VmAlertmanager container image to be deployed. Defaults to the value of `version`.
	// Similar to a tag, but the SHA explicitly deploys an immutable container image.
	// Version and Tag are ignored if SHA is set.
	SHA string `json:"sha,omitempty"`
	// BaseImage that is used to deploy pods, without tag.
	BaseImage string `json:"baseImage,omitempty"`
	// ImagePullSecrets An optional list of references to secrets in the same namespace
	// to use for pulling prometheus and alertmanager images from registries
	// see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Secrets is a list of Secrets in the same namespace as the VmAlertmanager
	// object, which shall be mounted into the VmAlertmanager Pods.
	// The Secrets are mounted into /etc/alertmanager/secrets/<secret-name>.
	Secrets []string `json:"secrets,omitempty"`
	// ConfigMaps is a list of ConfigMaps in the same namespace as the VmAlertmanager
	// object, which shall be mounted into the VmAlertmanager Pods.
	// The ConfigMaps are mounted into /etc/alertmanager/configmaps/<configmap-name>.
	ConfigMaps []string `json:"configMaps,omitempty"`
	// ConfigSecret is the name of a Kubernetes Secret in the same namespace as the
	// VmAlertmanager object, which contains configuration for this VmAlertmanager
	// instance. Defaults to 'alertmanager-<alertmanager-name>'
	// The secret is mounted into /etc/alertmanager/config.
	ConfigSecret string `json:"configSecret,omitempty"`
	// Log level for VmAlertmanager to be configured with.
	LogLevel string `json:"logLevel,omitempty"`
	// LogFormat for VmAlertmanager to be configured with.
	LogFormat string `json:"logFormat,omitempty"`
	// ReplicaCount Size is the expected size of the alertmanager cluster. The controller will
	// eventually make the size of the running cluster equal to the expected
	// +kubebuilder:validation:Minimum:=1
	ReplicaCount *int32 `json:"replicaCount,omitempty"`
	// Retention Time duration VmAlertmanager shall retain data for. Default is '120h',
	// and must match the regular expression `[0-9]+(ms|s|m|h)` (milliseconds seconds minutes hours).
	// +kubebuilder:validation:Pattern:="[0-9]+(ms|s|m|h)"
	Retention string `json:"retention,omitempty"`
	// Storage is the definition of how storage will be used by the VmAlertmanager
	// instances.
	Storage *StorageSpec `json:"storage,omitempty"`
	// Volumes allows configuration of additional volumes on the output StatefulSet definition.
	// Volumes specified will be appended to other volumes that are generated as a result of
	// StorageSpec objects.
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition.
	// VolumeMounts specified will be appended to other VolumeMounts in the alertmanager container,
	// that are generated as a result of StorageSpec objects.
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// ExternalURL the VmAlertmanager instances will be available under. This is
	// necessary to generate correct URLs. This is necessary if VmAlertmanager is not
	// served from root of a DNS name.
	ExternalURL string `json:"externalURL,omitempty"`
	// RoutePrefix VmAlertmanager registers HTTP handlers for. This is useful,
	// if using ExternalURL and a proxy is rewriting HTTP routes of a request,
	// and the actual ExternalURL is still true, but the server serves requests
	// under a different route prefix. For example for use with `kubectl proxy`.
	RoutePrefix string `json:"routePrefix,omitempty"`
	// Paused If set to true all actions on the underlaying managed objects are not
	// goint to be performed, except for delete actions.
	Paused bool `json:"paused,omitempty"`
	// NodeSelector Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Resources container resource request and limits,
	// https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity If specified, the pod's scheduling constraints.
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// Tolerations If specified, the pod's tolerations.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// SecurityContext holds pod-level security attributes and common container settings.
	// This defaults to the default PodSecurityContext.
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// Prometheus Pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// ListenLocal makes the VmAlertmanager server listen on loopback, so that it
	// does not bind against the Pod IP. Note this is only for the VmAlertmanager
	// UI, not the gossip communication.
	ListenLocal bool `json:"listenLocal,omitempty"`
	// Containers allows injecting additional containers. This is meant to
	// allow adding an authentication proxy to an VmAlertmanager pod.
	Containers []v1.Container `json:"containers,omitempty"`
	// InitContainers allows adding initContainers to the pod definition. Those can be used to e.g.
	// fetch secrets for injection into the VmAlertmanager configuration from external sources. Any
	// errors during the execution of an initContainer will lead to a restart of the Pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// Using initContainers for any use case other then secret fetching is entirely outside the scope
	// of what the maintainers will support and by doing so, you accept that this behaviour may break
	// at any time without notice.
	InitContainers []v1.Container `json:"initContainers,omitempty"`
	// PriorityClassName class assigned to the Pods
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// AdditionalPeers allows injecting a set of additional Alertmanagers to peer with to form a highly available cluster.
	AdditionalPeers []string `json:"additionalPeers,omitempty"`
	// ClusterAdvertiseAddress is the explicit address to advertise in cluster.
	// Needs to be provided for non RFC1918 [1] (public) addresses.
	// [1] RFC1918: https://tools.ietf.org/html/rfc1918
	ClusterAdvertiseAddress string `json:"clusterAdvertiseAddress,omitempty"`
	// PortName used for the pods and governing service.
	// This defaults to web
	PortName string `json:"portName,omitempty"`
}

// VmAlertmanagerList is a list of Alertmanagers.
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VmAlertmanagerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	// List of Alertmanagers
	Items []VmAlertmanager `json:"items"`
}

// VmAlertmanagerStatus is the most recent observed status of the VmAlertmanager cluster. Read-only. Not
// included when requesting from the apiserver, only from the Prometheus
// Operator API itself. More info:
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
// +k8s:openapi-gen=true
type VmAlertmanagerStatus struct {
	// Paused Represents whether any actions on the underlaying managed objects are
	// being performed. Only delete actions will be performed.
	Paused bool `json:"paused"`
	// ReplicaCount Total number of non-terminated pods targeted by this VmAlertmanager
	// cluster (their labels match the selector).
	Replicas int32 `json:"replicas"`
	// UpdatedReplicas Total number of non-terminated pods targeted by this VmAlertmanager
	// cluster that have the desired version spec.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// AvailableReplicas Total number of available pods (ready for at least minReadySeconds)
	// targeted by this VmAlertmanager cluster.
	AvailableReplicas int32 `json:"availableReplicas"`
	// UnavailableReplicas Total number of unavailable pods targeted by this VmAlertmanager cluster.
	UnavailableReplicas int32 `json:"unavailableReplicas"`
}

func (cr VmAlertmanager) Name() string {
	return cr.ObjectMeta.Name
}

func (cr *VmAlertmanager) AsOwner() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         cr.APIVersion,
			Kind:               cr.Kind,
			Name:               cr.Name(),
			UID:                cr.UID,
			Controller:         pointer.BoolPtr(true),
			BlockOwnerDeletion: pointer.BoolPtr(true),
		},
	}
}

func (cr VmAlertmanager) PodAnnotations() map[string]string {
	annotations := map[string]string{}
	if cr.Spec.PodMetadata != nil {
		for annotation, value := range cr.Spec.PodMetadata.Annotations {
			annotations[annotation] = value
		}
	}
	return annotations
}

func (cr VmAlertmanager) Annotations() map[string]string {
	annotations := make(map[string]string)
	for annotation, value := range cr.ObjectMeta.Annotations {
		if !strings.HasPrefix(annotation, "kubectl.kubernetes.io/") {
			annotations[annotation] = value
		}
	}
	return annotations
}

func (cr VmAlertmanager) SelectorLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      "vmalertmanager",
		"app.kubernetes.io/instance":  cr.Name(),
		"app.kubernetes.io/component": "monitoring",
		"managed-by":                  "vm-operator",
	}
}

func (cr VmAlertmanager) PodLabels() map[string]string {
	labels := cr.SelectorLabels()
	if cr.Spec.PodMetadata != nil {
		for label, value := range cr.Spec.PodMetadata.Labels {
			labels[label] = value
		}
	}
	return labels
}

func (cr VmAlertmanager) FinalLabels() map[string]string {
	labels := cr.SelectorLabels()
	if cr.ObjectMeta.Labels != nil {
		for label, value := range cr.ObjectMeta.Labels {
			labels[label] = value
		}
	}
	return labels
}

func (cr VmAlertmanager) PrefixedName() string {
	return fmt.Sprintf("vmalertmanager-%s", cr.Name())
}

func init() {
	SchemeBuilder.Register(&VmAlertmanager{}, &VmAlertmanagerList{})
}
