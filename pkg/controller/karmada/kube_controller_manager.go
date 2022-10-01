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
	clientutil "github.com/carlory/firefly/pkg/util/client"
	maputil "github.com/carlory/firefly/pkg/util/map"
)

func (ctrl *KarmadaController) EnsureKubeControllerManager(karmada *installv1alpha1.Karmada) error {
	return ctrl.EnsureKubeControllerManagerDeployment(karmada)
}

func (ctrl *KarmadaController) EnsureKubeControllerManagerDeployment(karmada *installv1alpha1.Karmada) error {
	componentName := util.ComponentName(constants.KarmadaComponentKubeControllerManager, karmada.Name)
	kcm := karmada.Spec.ControllerManager.KubeControllerManager
	repository := karmada.Spec.ImageRepository
	tag := karmada.Spec.KubernetesVersion
	if kcm.ImageRepository != "" {
		repository = kcm.ImageRepository
	}
	if kcm.ImageTag != "" {
		tag = kcm.ImageTag
	}

	defaultArgs := map[string]string{
		"allocate-node-cidrs":              "true",
		"authentication-kubeconfig":        "/etc/kubeconfig",
		"authorization-kubeconfig":         "/etc/kubeconfig",
		"bind-address":                     "0.0.0.0",
		"client-ca-file":                   "/etc/kubernetes/pki/ca.crt",
		"cluster-cidr":                     "10.244.0.0/16",
		"cluster-name":                     "kubernetes",
		"cluster-signing-cert-file":        "/etc/kubernetes/pki/ca.crt",
		"cluster-signing-key-file":         "/etc/kubernetes/pki/ca.key",
		"kubeconfig":                       "/etc/kubeconfig",
		"leader-elect":                     "true",
		"node-cidr-mask-size":              "24",
		"port":                             "0",
		"root-ca-file":                     "/etc/kubernetes/pki/ca.crt",
		"service-account-private-key-file": "/etc/kubernetes/pki/karmada.key",
		"service-cluster-ip-range":         karmada.Spec.Networking.ServiceSubnet,
		"use-service-account-credentials":  "true",
		"v":                                "4",
	}
	if kcm.Controllers != nil {
		defaultArgs["controllers"] = strings.Join(kcm.Controllers, ",")
	}
	computedArgs := maputil.MergeStringMaps(defaultArgs, kcm.ExtraArgs)
	args := maputil.ConvertToCommandOrArgs(computedArgs)

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
							Image:           util.ComponentImageName(repository, "kube-controller-manager", tag),
							ImagePullPolicy: "IfNotPresent",
							Command:         []string{"kube-controller-manager"},
							Args:            args,
							Resources:       kcm.Resources,
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
	return clientutil.CreateOrUpdateDeployment(ctrl.client, deployment)
}
