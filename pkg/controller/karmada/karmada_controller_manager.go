package karmada

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	installv1alpha1 "github.com/carlory/firefly/pkg/apis/install/v1alpha1"
	"github.com/carlory/firefly/pkg/constants"
	"github.com/carlory/firefly/pkg/util"
)

func (ctrl *KarmadaController) EnsureKarmadaControllerManager(karmada *installv1alpha1.Karmada) error {
	return ctrl.EnsureKarmadaControllerManagerDeployment(karmada)
}

func (ctrl *KarmadaController) EnsureKarmadaControllerManagerDeployment(karmada *installv1alpha1.Karmada) error {
	componentName := util.ComponentName(constants.KarmadaComponentControllerManager, karmada.Name)
	repository := karmada.Spec.ImageRepository
	version := karmada.Spec.KarmadaVersion

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      componentName,
			Namespace: karmada.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": componentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": componentName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "karmada-controller-manager",
							Image:           util.ComponentImageName(repository, constants.KarmadaComponentControllerManager, version),
							ImagePullPolicy: "IfNotPresent",
							Command: []string{
								"/bin/karmada-controller-manager",
								"--bind-address=0.0.0.0",
								"--kubeconfig=/etc/kubeconfig",
								"--cluster-status-update-frequency=10s",
								"--secure-port=10357",
								"--feature-gates=PropagateDeps=true",
								"--v=4",
							},
							Resources: karmada.Spec.ControllerManager.KarmadaControllerManager.Resources,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "kubeconfig",
									MountPath: "/etc/kubeconfig",
									SubPath:   "kubeconfig",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "kubeconfig",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: fmt.Sprintf("%s-kubeconfig", karmada.Name),
								},
							},
						},
					},
				},
			},
		},
	}
	controllerutil.SetOwnerReference(karmada, deployment, scheme.Scheme)
	return CreateOrUpdateDeployment(ctrl.client, deployment)
}
