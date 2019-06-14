package redis

import (
	"fmt"

	"github.com/lithammer/dedent"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	redisv1alpha1 "github.com/kube-incubator/redis-operator/pkg/apis/redis/v1alpha1"
	"github.com/kube-incubator/redis-operator/pkg/constants"
	"github.com/kube-incubator/redis-operator/pkg/util"
)

func GenerateSentinelService(r *redisv1alpha1.Redis, labels map[string]string) *corev1.Service {
	name := util.GetSentinelName(r)
	namespace := r.Namespace

	sentinelTargetPort := intstr.FromInt(26379)
	labels = util.MergeLabels(labels, util.GetLabels(constants.SentinelRoleName, r.Name))

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "sentinel",
					Port:       26379,
					TargetPort: sentinelTargetPort,
					Protocol:   "TCP",
				},
			},
		},
	}
}

func GenerateRedisService(r *redisv1alpha1.Redis, labels map[string]string) *corev1.Service {
	name := util.GetRedisName(r)
	namespace := r.Namespace

	labels = util.MergeLabels(labels, util.GetLabels(constants.RedisRoleName, r.Name))

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   "http",
				"prometheus.io/path":   "/metrics",
			},
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
			Ports: []corev1.ServicePort{
				{
					Port:     constants.ExporterPort,
					Protocol: corev1.ProtocolTCP,
					Name:     constants.ExporterPortName,
				},
			},
			Selector: labels,
		},
	}
}

func GenerateSentinelConfigMap(r *redisv1alpha1.Redis, labels map[string]string) *corev1.ConfigMap {
	name := util.GetSentinelName(r)
	namespace := r.Namespace

	labels = util.MergeLabels(labels, util.GetLabels(constants.SentinelRoleName, r.Name))
	sentinelConfigFileContent := dedent.Dedent(`
		sentinel monitor master 127.0.0.1 6379 2
		sentinel down-after-milliseconds master 1000
		sentinel failover-timeout master 3000
		sentinel parallel-syncs master 2
	`)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			constants.SentinelConfigFileName: sentinelConfigFileContent,
		},
	}
}

func GenerateRedisConfigMap(r *redisv1alpha1.Redis, labels map[string]string) *corev1.ConfigMap {
	name := util.GetRedisName(r)
	namespace := r.Namespace

	labels = util.MergeLabels(labels, util.GetLabels(constants.RedisRoleName, r.Name))
	redisConfigFileContent := dedent.Dedent(`
		slaveof 127.0.0.1 6379
		tcp-keepalive 60
		save 900 1
		save 300 10
	`)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			constants.RedisConfigFileName: redisConfigFileContent,
		},
	}
}

func GenerateRedisShutdownConfigMap(r *redisv1alpha1.Redis, labels map[string]string) *corev1.ConfigMap {
	name := util.GetRedisShutdownConfigMapName(r)
	namespace := r.Namespace

	labels = util.MergeLabels(labels, util.GetLabels(constants.RedisRoleName, r.Name))
	shutdownContent := dedent.Dedent(`
		master=$(redis-cli -h ${RFS_REDIS_SERVICE_HOST} -p ${RFS_REDIS_SERVICE_PORT_SENTINEL} --csv SENTINEL get-master-addr-by-name master | tr ',' ' ' | tr -d '\"' |cut -d' ' -f1)
		redis-cli SAVE
		if [[ $master ==  $(hostname -i) ]]; then
  			redis-cli -h ${RFS_REDIS_SERVICE_HOST} -p ${RFS_REDIS_SERVICE_PORT_SENTINEL} SENTINEL failover master
		fi
	`)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"shutdown.sh": shutdownContent,
		},
	}
}

func GenerateRedisStatefulSet(r *redisv1alpha1.Redis, labels map[string]string) *appsv1beta2.StatefulSet {
	name := util.GetRedisName(r)
	namespace := r.Namespace

	spec := r.Spec
	redisImage := getRedisImage(r)
	redisCommand := getRedisCommand(r)
	resources := getRedisResources(spec)
	labels = util.MergeLabels(labels, util.GetLabels(constants.RedisRoleName, r.Name))
	volumeMounts := getRedisVolumeMounts(r)
	volumes := getRedisVolumes(r)

	ss := &appsv1beta2.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1beta2.StatefulSetSpec{
			ServiceName: name,
			Replicas:    &spec.Redis.Replicas,
			UpdateStrategy: appsv1beta2.StatefulSetUpdateStrategy{
				Type: "RollingUpdate",
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity:    r.Spec.NodeAffinity,
						PodAntiAffinity: createPodAntiAffinity(r.Spec.HardAntiAffinity, labels),
					},
					Tolerations:     r.Spec.Tolerations,
					SecurityContext: r.Spec.SecurityContext,
					Containers: []corev1.Container{
						{
							Name:            "redis",
							Image:           redisImage,
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								{
									Name:          "redis",
									ContainerPort: 6379,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: volumeMounts,
							Command:      redisCommand,
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: constants.GraceTime,
								TimeoutSeconds:      5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) ping",
										},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: constants.GraceTime,
								TimeoutSeconds:      5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) ping",
										},
									},
								},
							},
							Resources: resources,
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/sh", "-c", "/redis-shutdown/shutdown.sh"},
									},
								},
							},
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	if r.Spec.Redis.Storage.PersistentVolumeClaim != nil {
		if !r.Spec.Redis.Storage.KeepAfterDeletion {
			// Set an owner reference so the persistent volumes are deleted when the Redis is
			r.Spec.Redis.Storage.PersistentVolumeClaim.OwnerReferences = []metav1.OwnerReference{
				*metav1.NewControllerRef(r, redisv1alpha1.SchemeGroupVersion.WithKind("Redis")),
			}
		}
		ss.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			*r.Spec.Redis.Storage.PersistentVolumeClaim,
		}
	}

	if r.Spec.Redis.Exporter {
		exporter := createRedisExporterContainer(r)
		ss.Spec.Template.Spec.Containers = append(ss.Spec.Template.Spec.Containers, exporter)
	}

	return ss
}

func GenerateSentinelDeployment(r *redisv1alpha1.Redis, labels map[string]string) *appsv1beta2.Deployment {
	name := util.GetSentinelName(r)
	configMapName := util.GetSentinelName(r)
	namespace := r.Namespace

	spec := r.Spec
	redisImage := getRedisImage(r)
	sentinelCommand := getSentinelCommand(r)
	resources := getSentinelResources(spec)
	labels = util.MergeLabels(labels, util.GetLabels(constants.SentinelRoleName, r.Name))

	return &appsv1beta2.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1beta2.DeploymentSpec{
			Replicas: &spec.Sentinel.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity:    r.Spec.NodeAffinity,
						PodAntiAffinity: createPodAntiAffinity(r.Spec.HardAntiAffinity, labels),
					},
					Tolerations:     r.Spec.Tolerations,
					SecurityContext: r.Spec.SecurityContext,
					InitContainers: []corev1.Container{
						{
							Name:            "sentinel-config-copy",
							Image:           redisImage,
							ImagePullPolicy: "IfNotPresent",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "sentinel-config",
									MountPath: "/redis",
								},
								{
									Name:      "sentinel-config-writable",
									MountPath: "/redis-writable",
								},
							},
							Command: []string{
								"cp",
								fmt.Sprintf("/redis/%s", constants.SentinelConfigFileName),
								fmt.Sprintf("/redis-writable/%s", constants.SentinelConfigFileName),
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("16Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("16Mi"),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "sentinel",
							Image:           redisImage,
							ImagePullPolicy: "Always",
							Ports: []corev1.ContainerPort{
								{
									Name:          "sentinel",
									ContainerPort: 26379,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "sentinel-config-writable",
									MountPath: "/redis",
								},
							},
							Command: sentinelCommand,
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: constants.GraceTime,
								TimeoutSeconds:      5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) -p 26379 ping",
										},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: constants.GraceTime,
								TimeoutSeconds:      5,
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"sh",
											"-c",
											"redis-cli -h $(hostname) -p 26379 ping",
										},
									},
								},
							},
							Resources: resources,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "sentinel-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
									},
								},
							},
						},
						{
							Name: "sentinel-config-writable",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

func GeneratePodDisruptionBudget(name string, namespace string, labels map[string]string, minAvailable intstr.IntOrString) *policyv1beta1.PodDisruptionBudget {
	return &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
	}
}

func getSentinelResources(spec redisv1alpha1.RedisSpec) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: getRequests(spec.Sentinel.Resources),
		Limits:   getLimits(spec.Sentinel.Resources),
	}
}

func getRedisResources(spec redisv1alpha1.RedisSpec) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: getRequests(spec.Redis.Resources),
		Limits:   getLimits(spec.Redis.Resources),
	}
}

func getLimits(resources redisv1alpha1.RedisResources) corev1.ResourceList {
	return generateResourceList(resources.Limits.CPU, resources.Limits.Memory)
}

func getRequests(resources redisv1alpha1.RedisResources) corev1.ResourceList {
	return generateResourceList(resources.Requests.CPU, resources.Requests.Memory)
}

func generateResourceList(cpu string, memory string) corev1.ResourceList {
	resources := corev1.ResourceList{}
	if cpu != "" {
		resources[corev1.ResourceCPU], _ = resource.ParseQuantity(cpu)
	}
	if memory != "" {
		resources[corev1.ResourceMemory], _ = resource.ParseQuantity(memory)
	}
	return resources
}

func createRedisExporterContainer(r *redisv1alpha1.Redis) corev1.Container {
	exporterImage := getRedisExporterImage(r)

	// Define readiness and liveness probes only if config option to disable isn't set
	var readinessProbe, livenessProbe *corev1.Probe
	if !r.Spec.Redis.DisableExporterProbes {
		readinessProbe = &corev1.Probe{
			InitialDelaySeconds: 10,
			TimeoutSeconds:      3,
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/",
					Port: intstr.FromString("metrics"),
				},
			},
		}

		livenessProbe = &corev1.Probe{
			TimeoutSeconds: 3,
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/",
					Port: intstr.FromString("metrics"),
				},
			},
		}
	}

	return corev1.Container{
		Name:            constants.ExporterContainerName,
		Image:           exporterImage,
		ImagePullPolicy: "Always",
		Env: []corev1.EnvVar{
			{
				Name: "REDIS_ALIAS",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: constants.ExporterPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		ReadinessProbe: readinessProbe,
		LivenessProbe:  livenessProbe,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(constants.ExporterDefaultLimitCPU),
				corev1.ResourceMemory: resource.MustParse(constants.ExporterDefaultLimitMemory),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(constants.ExporterDefaultRequestCPU),
				corev1.ResourceMemory: resource.MustParse(constants.ExporterDefaultRequestMemory),
			},
		},
	}
}

func createPodAntiAffinity(hard bool, labels map[string]string) *corev1.PodAntiAffinity {
	if hard {
		// Return a HARD anti-affinity (no same pods on one node)
		return &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					TopologyKey: constants.HostnameTopologyKey,
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: labels,
					},
				},
			},
		}
	}

	// Return a SOFT anti-affinity
	return &corev1.PodAntiAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
			{
				Weight: 100,
				PodAffinityTerm: corev1.PodAffinityTerm{
					TopologyKey: constants.HostnameTopologyKey,
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: labels,
					},
				},
			},
		},
	}
}

func GetQuorum(r *redisv1alpha1.Redis) int32 {
	return getQuorum(r)
}

func getQuorum(r *redisv1alpha1.Redis) int32 {
	return r.Spec.Sentinel.Replicas/2 + 1
}

func getRedisImage(r *redisv1alpha1.Redis) string {
	return fmt.Sprintf("%s:%s", r.Spec.Redis.Image, r.Spec.Redis.Version)
}

func getRedisExporterImage(r *redisv1alpha1.Redis) string {
	return fmt.Sprintf("%s:%s", r.Spec.Redis.ExporterImage, r.Spec.Redis.ExporterVersion)
}

func getRedisVolumeMounts(r *redisv1alpha1.Redis) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      constants.RedisConfigurationVolumeName,
			MountPath: "/redis",
		},
		{
			Name:      constants.RedisShutdownConfigurationVolumeName,
			MountPath: "/redis-shutdown",
		},
		{
			Name:      getRedisDataVolumeName(r),
			MountPath: "/data",
		},
	}

	return volumeMounts
}

func getRedisVolumes(r *redisv1alpha1.Redis) []corev1.Volume {
	configMapName := util.GetRedisName(r)
	shutdownConfigMapName := util.GetRedisShutdownConfigMapName(r)

	executeMode := int32(0744)
	volumes := []corev1.Volume{
		{
			Name: constants.RedisConfigurationVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
				},
			},
		},
		{
			Name: constants.RedisShutdownConfigurationVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: shutdownConfigMapName,
					},
					DefaultMode: &executeMode,
				},
			},
		},
	}

	dataVolume := getRedisDataVolume(r)
	if dataVolume != nil {
		volumes = append(volumes, *dataVolume)
	}

	return volumes
}

func getRedisDataVolume(r *redisv1alpha1.Redis) *corev1.Volume {
	// This will find the volumed desired by the user. If no volume defined
	// an EmptyDir will be used by default
	switch {
	case r.Spec.Redis.Storage.PersistentVolumeClaim != nil:
		return nil
	case r.Spec.Redis.Storage.EmptyDir != nil:
		return &corev1.Volume{
			Name: constants.RedisStorageVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: r.Spec.Redis.Storage.EmptyDir,
			},
		}
	default:
		return &corev1.Volume{
			Name: constants.RedisStorageVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
	}
}

func getRedisDataVolumeName(r *redisv1alpha1.Redis) string {
	switch {
	case r.Spec.Redis.Storage.PersistentVolumeClaim != nil:
		return r.Spec.Redis.Storage.PersistentVolumeClaim.Name
	case r.Spec.Redis.Storage.EmptyDir != nil:
		return constants.RedisStorageVolumeName
	default:
		return constants.RedisStorageVolumeName
	}
}

func getRedisCommand(r *redisv1alpha1.Redis) []string {
	if len(r.Spec.Redis.Command) > 0 {
		return r.Spec.Redis.Command
	}
	return []string{
		"redis-server",
		fmt.Sprintf("/redis/%s", constants.RedisConfigFileName),
	}
}

func getSentinelCommand(r *redisv1alpha1.Redis) []string {
	if len(r.Spec.Sentinel.Command) > 0 {
		return r.Spec.Sentinel.Command
	}
	return []string{
		"redis-server",
		fmt.Sprintf("/redis/%s", constants.SentinelConfigFileName),
		"--sentinel",
	}
}
