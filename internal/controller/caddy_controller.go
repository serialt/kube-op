/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	webv1 "github.com/krabt/caddy-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// CaddyReconciler reconciles a Caddy object
type CaddyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=web.local.com,resources=caddies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=web.local.com,resources=caddies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=web.local.com,resources=caddies/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

func (r *CaddyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := logf.FromContext(ctx)
	caddy := webv1.Caddy{}

	if err := r.Get(ctx, req.NamespacedName, &caddy); err != nil {
		if apierrors.IsNotFound(err) {
			// 资源已被物理删除，无需处理
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	labels := map[string]string{
		"app": caddy.Name,
	}

	// 创建 Deployment 对象
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      caddy.Name,
			Namespace: caddy.Namespace,
			Labels:    labels,
		}}
	if err := controllerutil.SetControllerReference(&caddy, deploy, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		deploy.Spec = appsv1.DeploymentSpec{
			Replicas: &caddy.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  caddy.Name,
							Image: caddy.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: caddy.Spec.Port,
								},
							},
						},
					},
				},
			},
		}

		// 关键操作：设置 OwnerReference
		// 确保删除 Guestbook CR 时，自动清理关联的 Deployment
		if err := controllerutil.SetControllerReference(&caddy, deploy, r.Scheme); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to create or update Deployment")
		caddy.Status.AvailableReplicas = 0
		return ctrl.Result{}, err
	}
	logger.Info("Deployment操作结果", "result", result)

	// 构建svc
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      caddy.Name,
			Namespace: caddy.Namespace,
		},
	}
	result, err = controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec = corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       caddy.Spec.SvcPort,
					TargetPort: intstr.FromInt(int(caddy.Spec.Port)),
				},
			},
		}
		return controllerutil.SetControllerReference(&caddy, svc, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "Failed to create or update Service")
		return ctrl.Result{}, err
	}
	logger.Info("Service操作结果", "result", result)

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      caddy.Name,
			Namespace: caddy.Namespace,
		},
	}
	if caddy.Spec.Domain == "" {
		if err := r.Get(ctx, client.ObjectKeyFromObject(ingress), ingress); err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		} else if err == nil && metav1.IsControlledBy(ingress, &caddy) {
			if err := r.Delete(ctx, ingress); err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("Ingress deleted because domain is not configured")
		}
		return ctrl.Result{}, nil
	}

	pathType := networkingv1.PathTypePrefix
	result, err = controllerutil.CreateOrUpdate(ctx, r.Client, ingress, func() error {
		ingress.Spec = networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: caddy.Spec.Domain,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: caddy.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: caddy.Spec.SvcPort,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		return controllerutil.SetControllerReference(&caddy, ingress, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "Failed to create or update Ingress")
		return ctrl.Result{}, err
	}
	logger.Info("Ingress操作结果", "result", result)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CaddyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1.Caddy{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
