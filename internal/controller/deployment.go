package controller

import (
	"fmt"

	kaimeraaiv1 "github.com/kaimera-ai/kaimera/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type deploymentConfig struct {
	image       string
	tolerations []corev1.Toleration
	limits      corev1.ResourceList
	command     []string
	env         []corev1.EnvVar
}

func getDeploymentConfiguration(provider, runtime, modelName string, maxModelLength int) *deploymentConfig {
	runtimeProviderMap := map[string]map[string]deploymentConfig{
		"Ollama": {
			"cpu": {
				image: "patnaikshekhar/kaimera-ollama:v0.0.1",
				env: []corev1.EnvVar{
					{
						Name:  "MODEL_NAME",
						Value: modelName,
					},
				},
			},
			"gpu": {
				image: "patnaikshekhar/kaimera-ollama:v0.0.1",
				env: []corev1.EnvVar{
					{
						Name:  "MODEL_NAME",
						Value: modelName,
					},
				},
				tolerations: []corev1.Toleration{
					{
						Key:      "nvidia.com/gpu",
						Operator: "Exists",
						Effect:   "NoSchedule",
					},
				},
				limits: corev1.ResourceList{
					"nvidia.com/gpu": resource.MustParse("1"),
				},
			},
			"default": {
				image: "patnaikshekhar/kaimera-ollama:v0.0.1",
				env: []corev1.EnvVar{
					{
						Name:  "MODEL_NAME",
						Value: modelName,
					},
				},
			},
		},
		"vLLM": {
			"cpu": {
				image: "patnaikshekhar/vllm-cpu:1",
				command: []string{
					"vllm",
					"serve",
					"--dtype",
					"auto",
					"--max-model-len",
					fmt.Sprintf("%d", maxModelLength),
					modelName,
				},
			},
			"gpu": {
				image: "vllm/vllm-openai:latest",
				tolerations: []corev1.Toleration{
					{
						Key:      "nvidia.com/gpu",
						Operator: "Exists",
						Effect:   "NoSchedule",
					},
				},
				limits: corev1.ResourceList{
					"nvidia.com/gpu": resource.MustParse("1"),
				},
				command: []string{
					"vllm",
					"serve",
					"--dtype",
					"auto",
					"--max-model-len",
					fmt.Sprintf("%d", maxModelLength),
					modelName,
				},
			},
			"default": {
				image: "patnaikshekhar/vllm-cpu:1",
				command: []string{
					"vllm",
					"serve",
					"--dtype",
					"auto",
					"--max-model-len",
					fmt.Sprintf("%d", maxModelLength),
					modelName,
				},
			},
		},
	}

	runtimeMap, exists := runtimeProviderMap[provider]
	if !exists {
		runtimeMap = runtimeProviderMap["vLLM"]
	}

	deploymentConfiguration, exists := runtimeMap[runtime]
	if !exists {
		deploymentConfiguration = runtimeMap["default"]
	}

	return &deploymentConfiguration
}

func (r *ModelDeploymentReconciler) generateDeployment(md *kaimeraaiv1.ModelDeployment) (*appsv1.Deployment, error) {

	if md.Spec.Replicas == 0 {
		md.Spec.Replicas = 1
	}
	maxModelLength := 512
	if md.Spec.MaxModelLength > 0 {
		maxModelLength = int(md.Spec.MaxModelLength)
	}

	config := getDeploymentConfiguration(md.Spec.Provider, md.Spec.Runtime, md.Spec.ModelName, maxModelLength)

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      md.Name,
			Namespace: md.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &md.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": md.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": md.Name,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector: md.Spec.NodeSelectorLabels,
					Containers: []corev1.Container{
						{
							Name:            "app",
							Image:           config.image,
							ImagePullPolicy: "IfNotPresent",
							Command:         config.command,
							Resources: corev1.ResourceRequirements{
								Limits: config.limits,
							},
							Env: config.env,
						},
					},
					Tolerations: config.tolerations,
				},
			},
		},
	}

	err := ctrl.SetControllerReference(md, deploy, r.Scheme)
	if err != nil {
		return nil, err
	}

	return deploy, nil
}
