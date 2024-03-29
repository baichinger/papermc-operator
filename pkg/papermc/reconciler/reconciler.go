package reconciler

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	papermciov1 "github.com/baichinger/papermc-operator/api/v1"
	papermc "github.com/baichinger/papermc-operator/pkg/papermc/client"
)

const (
	imageDownloader = "docker.io/busybox:latest"
	imageServer     = "gcr.io/distroless/java17-debian11:nonroot"

	labelName     = "app.kubernetes.io/name"
	labelInstance = "app.kubernetes.io/instance"
	labelVersion  = "app.kubernetes.io/version"

	objectName = "PaperMC"

	conditionTypeAvailable = "Available"
	conditionTypeDegraded  = "Degraded"

	runAsUserId = 1000

	serverPort = 25565

	desiredVersionUpdateInterval = 2 * time.Hour
)

type Reconciler struct {
	client client.Client
	scheme *runtime.Scheme
	ctx    context.Context
	paper  *papermciov1.Paper
}

func NewPaperReconciler(client client.Client, scheme *runtime.Scheme, ctx context.Context, paper *papermciov1.Paper) *Reconciler {
	return &Reconciler{
		client: client,
		scheme: scheme,
		ctx:    ctx,
		paper:  paper,
	}
}

func (r *Reconciler) InitializeConditions() Result {
	if r.paper.Status.Conditions == nil || len(r.paper.Status.Conditions) == 0 {
		meta.SetStatusCondition(&r.paper.Status.Conditions, metav1.Condition{
			Type:    conditionTypeAvailable,
			Status:  metav1.ConditionUnknown,
			Reason:  "Reconciling",
			Message: "Starting reconciliation",
		})
		now := metav1.Now()
		r.paper.Status.UpdatedTimestamp = &now

		if err := r.client.Status().Update(r.ctx, r.paper); err != nil {
			return newFailedResult(err)
		}

		return newUpdatedResult()
	}

	return newSkippedResult()
}

func (r *Reconciler) ReconcileDesiredVersion() Result {
	now := metav1.Now()
	if r.paper.Status.DesiredState != nil && r.paper.Status.DesiredState.Version.Version == r.paper.Spec.Version &&
		r.paper.Status.DesiredState.UpdatedTimestamp.Time.Add(desiredVersionUpdateInterval).After(now.Time) {
		return newSkippedResult()
	}

	pmcClient := papermc.NewPapermcClient(r.ctx)

	if build, err := pmcClient.GetBuildForVersion(r.paper.Spec.Version); err != nil {
		return newFailedResult(err)
	} else if r.paper.Status.DesiredState == nil || r.paper.Status.DesiredState.Version.Version != r.paper.Spec.Version || r.paper.Status.DesiredState.Version.Build != build {
		url, err := pmcClient.GetUrlForVersionBuildDownload(r.paper.Spec.Version, build)
		if err != nil {
			return newFailedResult(err)
		}

		r.paper.Status.DesiredState = &papermciov1.DesiredState{
			Version: papermciov1.Version{
				Version: r.paper.Spec.Version,
				Build:   build,
			},
			Url: url,
		}

		meta.SetStatusCondition(&r.paper.Status.Conditions, metav1.Condition{
			Type:    conditionTypeAvailable,
			Status:  metav1.ConditionFalse,
			Reason:  "Reconciling",
			Message: "Version, build, and url available",
		})
	}

	r.paper.Status.DesiredState.UpdatedTimestamp = now
	r.paper.Status.UpdatedTimestamp = &now

	if err := r.client.Status().Update(r.ctx, r.paper); err != nil {
		return newFailedResult(err)
	}

	return newUpdatedResult()
}

func (r *Reconciler) ReconcileStatus() Result {
	if r.paper.Status.ActualState != nil && r.paper.Status.ActualState.Version == r.paper.Status.DesiredState.Version {
		return newSkippedResult()
	}

	r.paper.Status.ActualState = &papermciov1.ActualState{
		Version: r.paper.Status.DesiredState.Version,
	}

	meta.SetStatusCondition(&r.paper.Status.Conditions, metav1.Condition{
		Type:    conditionTypeAvailable,
		Status:  metav1.ConditionTrue,
		Reason:  "Reconciling",
		Message: "Done",
	})

	now := metav1.Now()
	r.paper.Status.UpdatedTimestamp = &now

	if err := r.client.Status().Update(r.ctx, r.paper); err != nil {
		return newFailedResult(err)
	}

	return newUpdatedResult()
}

func (r *Reconciler) ReconcilePersistentVolumeClaimForDesiredVersion() Result {
	name := buildObjectNameForVersion(r.paper.Name, r.paper.Status.DesiredState.Version)

	if err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.paper.Namespace, Name: name}, &corev1.PersistentVolumeClaim{}); err != nil {
		if !apierrors.IsNotFound(err) {
			return newFailedResult(err)
		}
	} else {
		// nothing to do, PVC exists
		return newSkippedResult()
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.paper.Namespace,
			Labels:    labelsForDesiredVersion(r.paper),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *resource.NewScaledQuantity(50, resource.Mega),
				},
			},
		},
	}

	err := ctrl.SetControllerReference(r.paper, pvc, r.scheme)
	if err != nil {
		return newFailedResult(err)
	}

	if err := r.client.Create(r.ctx, pvc); err != nil {
		return newFailedResult(err)
	}

	return newUpdatedResult()
}

func (r *Reconciler) ReconcilePersistentVolumeClaimForPaperInstance() Result {
	if err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.paper.Namespace, Name: r.paper.Name}, &corev1.PersistentVolumeClaim{}); err != nil {
		if !apierrors.IsNotFound(err) {
			return newFailedResult(err)
		}
	} else {
		// nothing to do, PVC exists
		return newSkippedResult()
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.paper.Name,
			Namespace: r.paper.Namespace,
			Labels:    labelsForPaperInstance(r.paper),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *resource.NewScaledQuantity(1, resource.Giga),
				},
			},
		},
	}

	err := ctrl.SetControllerReference(r.paper, pvc, r.scheme)
	if err != nil {
		return newFailedResult(err)
	}

	if err := r.client.Create(r.ctx, pvc); err != nil {
		return newFailedResult(err)
	}

	return newUpdatedResult()
}

func (r *Reconciler) ReconcileProvisionerForDesiredVersion() Result {
	if r.paper.Status.DesiredState == nil {
		return newFailedResult(fmt.Errorf("desired state undefined"))
	}

	name := buildObjectNameForVersion(r.paper.Name, r.paper.Status.DesiredState.Version)

	if r.paper.Status.ActualState != nil && r.paper.Status.DesiredState.Version == r.paper.Status.ActualState.Version {
		// nothing to do, provisioner already did its job
		return newSkippedResult()
	}

	existingPod := corev1.Pod{}
	if err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.paper.Namespace, Name: name}, &existingPod); err != nil {
		if !apierrors.IsNotFound(err) {
			return newFailedResult(err)
		}
	} else if existingPod.Status.Phase == corev1.PodFailed {
		// delete and try again
		err := r.client.Delete(r.ctx, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: r.paper.Namespace, Name: name}})
		if err != nil && !(apierrors.IsNotFound(err) || apierrors.IsGone(err)) {
			return newFailedResult(err)
		}
		return newUpdatedResult()
	} else if existingPod.Status.Phase == corev1.PodSucceeded {
		// move to next step, provisioner finished
		return newSkippedResult()
	} else {
		// nothing to do, provisioner Pod exists
		return newUpdatedResult()
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.paper.Namespace,
			Labels:    labelsForDesiredVersion(r.paper),
		},
		Spec: corev1.PodSpec{
			AutomountServiceAccountToken: pointer.Bool(false),
			Containers: []corev1.Container{{
				Name:       "paper",
				Image:      r.imageForPaperDownloader(r.paper),
				Command:    []string{"wget", "-O", "paper.jar", r.paper.Status.DesiredState.Url},
				WorkingDir: "/data",
				VolumeMounts: []corev1.VolumeMount{{
					Name:      "data",
					MountPath: "/data",
				}},
				SecurityContext: secureContainerSecurityContext(),
			}},
			// ServiceAccountName: p.Name,
			RestartPolicy:   corev1.RestartPolicyNever,
			SecurityContext: securePodSecurityContext(),
			Volumes: []corev1.Volume{{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: name,
					},
				},
			}},
		},
	}

	err := ctrl.SetControllerReference(r.paper, pod, r.scheme)
	if err != nil {
		return newFailedResult(err)
	}

	if err := r.client.Create(r.ctx, pod); err != nil {
		return newFailedResult(err)
	}

	return newUpdatedResult()
}

func (r *Reconciler) ReconcileConfigurationForPaperInstance() Result {
	if err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.paper.Namespace, Name: r.paper.Name}, &corev1.ConfigMap{}); err != nil {
		if !apierrors.IsNotFound(err) {
			return newFailedResult(err)
		}
	} else {
		// nothing to do, configuration exists
		return newSkippedResult()
	}

	cfg := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.paper.Name,
			Namespace: r.paper.Namespace,
			Labels:    labelsForPaperInstance(r.paper),
		},
		Data: map[string]string{
			"eula.txt": "eula=true",
		},
	}

	err := ctrl.SetControllerReference(r.paper, cfg, r.scheme)
	if err != nil {
		return newFailedResult(err)
	}

	if err := r.client.Create(r.ctx, cfg); err != nil {
		return newFailedResult(err)
	}

	return newUpdatedResult()
}

func (r *Reconciler) ReconcilePaperInstance() Result {
	// todo: recreate pod if unhealthy
	existingPod := corev1.Pod{}
	if err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.paper.Namespace, Name: r.paper.Name}, &existingPod); err != nil {
		if !apierrors.IsNotFound(err) {
			return newFailedResult(err)
		}
	} else if existingPod.Status.Phase == corev1.PodFailed {
		// failure, recreate paper pod
		return r.deletePaperInstance()
	} else if existingPod.Status.Phase == corev1.PodRunning && !labels.Equals(existingPod.Labels, labelsForDesiredVersion(r.paper)) {
		// upgrade, replace paper pod
		return r.deletePaperInstance()
	} else if existingPod.Status.Phase != corev1.PodRunning {
		// give it a moment
		return newUpdatedResult()
	} else {
		// nothing to do, paper instance Pod exists
		return newSkippedResult()
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.paper.Name,
			Namespace: r.paper.Namespace,
			Labels:    labelsForDesiredVersion(r.paper),
		},
		Spec: corev1.PodSpec{
			AutomountServiceAccountToken: pointer.Bool(false),
			Containers: []corev1.Container{{
				Name:       "paper",
				Image:      r.imageForPaperInstance(r.paper),
				Args:       []string{"/app/paper/paper.jar"},
				WorkingDir: "/app/data",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "app-paper",
						MountPath: "/app/paper",
						ReadOnly:  true,
					},
					{
						Name:      "app-data",
						MountPath: "/app/data",
					},
					{
						Name:      "configuration",
						MountPath: "/app/data/eula.txt",
						SubPath:   "eula.txt",
					},
					{
						Name:      "tmp",
						MountPath: "/tmp",
					},
				},
				//Resources: corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceMemory: resource.}},
				StartupProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						TCPSocket: &corev1.TCPSocketAction{
							Port: intstr.FromInt(serverPort),
						},
					},
					InitialDelaySeconds: 15,
					PeriodSeconds:       5,
					FailureThreshold:    60,
				},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						TCPSocket: &corev1.TCPSocketAction{
							Port: intstr.FromInt(serverPort),
						},
					},
					PeriodSeconds:    3,
					FailureThreshold: 1,
				},
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						TCPSocket: &corev1.TCPSocketAction{
							Port: intstr.FromInt(serverPort),
						},
					},
					PeriodSeconds: 5,
				},
				SecurityContext: secureContainerSecurityContext(),
			}},
			// ServiceAccountName: p.Name,
			RestartPolicy:   corev1.RestartPolicyAlways,
			SecurityContext: securePodSecurityContext(),
			Volumes: []corev1.Volume{
				{
					Name: "app-paper",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: buildObjectNameForVersion(r.paper.Name, r.paper.Status.DesiredState.Version),
						},
					},
				},
				{
					Name: "app-data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: r.paper.Name,
						},
					},
				},
				{
					Name: "configuration",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: r.paper.Name,
							},
							DefaultMode: pointer.Int32(0444),
							Optional:    pointer.Bool(false),
						},
					},
				},
				{
					Name: "tmp",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							SizeLimit: resource.NewScaledQuantity(100, resource.Mega),
						},
					},
				},
			},
		},
	}

	err := ctrl.SetControllerReference(r.paper, pod, r.scheme)
	if err != nil {
		return newFailedResult(err)
	}

	if err := r.client.Create(r.ctx, pod); err != nil {
		return newFailedResult(err)
	}

	return newUpdatedResult()
}

func (r *Reconciler) ReconcilePaperService() Result {
	existingService := corev1.Service{}
	if err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.paper.Namespace, Name: r.paper.Name}, &existingService); err != nil {
		if !apierrors.IsNotFound(err) {
			return newFailedResult(err)
		}
	} else {
		// nothing to do, paper instance Service exists
		return newSkippedResult()
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.paper.Name,
			Namespace: r.paper.Namespace,
			Labels:    labelsForPaperInstance(r.paper),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       serverPort,
					TargetPort: intstr.FromInt(serverPort),
				},
			},
			Selector: labelsForPaperInstance(r.paper),
			Type:     corev1.ServiceTypeNodePort,
		},
	}

	err := ctrl.SetControllerReference(r.paper, service, r.scheme)
	if err != nil {
		return newFailedResult(err)
	}

	if err := r.client.Create(r.ctx, service); err != nil {
		return newFailedResult(err)
	}

	return newUpdatedResult()
}

func (r *Reconciler) ReconcileOrphanObjects() Result {
	logger := log.FromContext(r.ctx)

	selectorString := fmt.Sprintf("app.kubernetes.io/instance=%s,app.kubernetes.io/version,app.kubernetes.io/version!=%s", r.paper.Name, r.paper.Status.DesiredState.Version.String())
	selector, err := labels.Parse(selectorString)
	if err != nil {
		logger.Info("failed to parse selector", "string", selectorString, "err", err)
		return newSkippedResult()
	}
	options := &client.ListOptions{
		LabelSelector: selector,
		Namespace:     r.paper.Namespace,
	}

	podList := &corev1.PodList{}
	if err := r.client.List(r.ctx, podList, options); err != nil {
		return newFailedResult(err)
	}

	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := r.client.List(r.ctx, pvcList, options); err != nil {
		return newFailedResult(err)
	}

	if len(podList.Items) == 0 && len(pvcList.Items) == 0 {
		return newSkippedResult()
	}

	for _, pod := range podList.Items {
		err := r.client.Delete(r.ctx, &pod)
		logger.Info("orphan object deleted", "kind", pod.TypeMeta.Kind, "name", pod.Name, "err", err)
	}

	for _, pvc := range pvcList.Items {
		err := r.client.Delete(r.ctx, &pvc)
		logger.Info("orphan object deleted", "kind", pvc.TypeMeta.Kind, "name", pvc.Name, "err", err)
	}

	return newUpdatedResult()
}

func (r *Reconciler) deletePaperInstance() Result {
	err := r.client.Delete(r.ctx, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: r.paper.Namespace, Name: r.paper.Name}})
	if err != nil && !(apierrors.IsNotFound(err) || apierrors.IsGone(err)) {
		return newFailedResult(err)
	}
	return newUpdatedResult()
}

func secureContainerSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"ALL",
			},
		},
		Privileged:               pointer.Bool(false),
		AllowPrivilegeEscalation: pointer.Bool(false),
		ReadOnlyRootFilesystem:   pointer.Bool(true),
		RunAsNonRoot:             pointer.Bool(true),
		RunAsUser:                pointer.Int64(runAsUserId),
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
}

func securePodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		FSGroup: pointer.Int64(runAsUserId),
	}
}

func (r *Reconciler) imageForPaperDownloader(_ *papermciov1.Paper) string {
	return imageDownloader
}

func (r *Reconciler) imageForPaperInstance(_ *papermciov1.Paper) string {
	return imageServer
}

func buildObjectNameForVersion(name string, version papermciov1.Version) string {
	return fmt.Sprintf("%s-%s", name, version.String())
}

func labelsForDesiredVersion(p *papermciov1.Paper) map[string]string {
	return map[string]string{
		labelName:     objectName,
		labelInstance: p.Name,
		labelVersion:  p.Status.DesiredState.Version.String(),
	}
}

func labelsForPaperInstance(p *papermciov1.Paper) map[string]string {
	return map[string]string{
		labelName:     objectName,
		labelInstance: p.Name,
	}
}
