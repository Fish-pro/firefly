package karmada

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	installv1alpha1 "github.com/carlory/firefly/pkg/apis/install/v1alpha1"
	"github.com/carlory/firefly/pkg/constants"
	"github.com/carlory/firefly/pkg/util"
)

func (ctrl *KarmadaController) EnsureKubeControllerManager(karmada *installv1alpha1.Karmada) error {
	return ctrl.EnsureKubeControllerManagerDeployment(karmada)
}

func (ctrl *KarmadaController) EnsureKubeControllerManagerDeployment(karmada *installv1alpha1.Karmada) error {
	componentName := util.ComponentName(constants.KarmadaComponentKubeControllerManager, karmada.Name)
	repository := karmada.Spec.ImageRepository
	version := karmada.Spec.KubernetesVersion

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
							Name:            "kube-controller-manager",
							Image:           util.ComponentImageName(repository, "kube-controller-manager", version),
							ImagePullPolicy: "IfNotPresent",
							Command: []string{
								"kube-controller-manager",
								"--allocate-node-cidrs=true",
								"--authentication-kubeconfig=/etc/kubeconfig",
								"--authorization-kubeconfig=/etc/kubeconfig",
								"--bind-address=0.0.0.0",
								"--client-ca-file=/etc/kubernetes/pki/ca.crt",
								"--cluster-cidr=10.244.0.0/16",
								"--cluster-name=kubernetes",
								"--cluster-signing-cert-file=/etc/kubernetes/pki/ca.crt",
								"--cluster-signing-key-file=/etc/kubernetes/pki/ca.key",
								fmt.Sprintf("--controllers=%s", strings.Join(karmada.Spec.ControllerManager.KubeControllerManager.Controllers, ",")),
								"--kubeconfig=/etc/kubeconfig",
								"--leader-elect=true",
								"--node-cidr-mask-size=24",
								"--port=0",
								"--root-ca-file=/etc/kubernetes/pki/ca.crt",
								"--service-account-private-key-file=/etc/kubernetes/pki/karmada.key",
								fmt.Sprintf("--service-cluster-ip-range=%s", karmada.Spec.Networking.ServiceSubnet),
								"--use-service-account-credentials=true",
								"--v=4",
							},
							Resources: karmada.Spec.ControllerManager.KubeControllerManager.Resources,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "k8s-certs",
									MountPath: "/etc/kubernetes/pki",
									ReadOnly:  true,
								},
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
							Name: "k8s-certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: fmt.Sprintf("%s-cert", util.ComponentName("karmada", karmada.Name)),
								},
							},
						},
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
