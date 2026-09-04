/*
Copyright 2025 serialt.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "github.com/serialt/web-operator/api/v1"
)

// GuestbookReconciler reconciles a Guestbook object
type GuestbookReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.web.imau.cc,resources=guestbooks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.web.imau.cc,resources=guestbooks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webapp.web.imau.cc,resources=guestbooks/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Guestbook object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *GuestbookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	guestbook := webappv1.Guestbook{}
	if err := r.Get(ctx, req.NamespacedName, &guestbook); err != nil {
		if apierrors.IsNotFound(err) {
			// 资源已被物理删除，无需处理
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	labels := map[string]string{
		"app": guestbook.Name,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      guestbook.Name,
			Namespace: guestbook.Namespace,
		},
	}

	if err := controllerutil.SetControllerReference(&guestbook, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// 关键操作 ：CreateOrUpdate
	// 如果 Deployment 不存在则创建，存在则更新到最新状态
	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		// 示例：创建或更新一个 Deployment
		// 在此函数中更新 Deployment 的期望状态
		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: &guestbook.Spec.Replicas,
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
							Name:  guestbook.Name,
							Image: guestbook.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: guestbook.Spec.Port,
								},
							},
						},
					},
				},
			},
		}

		// 关键操作：设置 OwnerReference
		// 确保删除 Guestbook CR 时，自动清理关联的 Deployment
		if err := controllerutil.SetControllerReference(&guestbook, deployment, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error(err, "Failed to create or update Deployment")
		guestbook.Status.AvailableReplicas = 0
		return ctrl.Result{}, err
	}

	logger.Info("Deployment操作结果", "result", result)

	// 构建svc
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      guestbook.Name,
			Namespace: guestbook.Namespace,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec = corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       guestbook.Spec.Port,
					TargetPort: intstr.FromInt(int(guestbook.Spec.Port)),
				},
			},
		}

		return controllerutil.SetControllerReference(&guestbook, svc, r.Scheme)

	})
	if err != nil {
		logger.Error(err, "Failed to create or update Service")
		return ctrl.Result{}, err
	}
	logger.Info("Service操作结果", "result", result)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GuestbookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Guestbook{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
