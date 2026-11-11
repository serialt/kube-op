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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	webv1 "github.com/krabt/caddy-operator/api/v1"
)

var _ = Describe("Caddy Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		caddy := &webv1.Caddy{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Caddy")
			err := k8sClient.Get(ctx, typeNamespacedName, caddy)
			if err != nil && errors.IsNotFound(err) {
				resource := &webv1.Caddy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: webv1.CaddySpec{
						Replicas: 1,
						Image:    "caddy:2",
						Port:     80,
						SvcPort:  80,
						Domain:   "caddy.example.com",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &webv1.Caddy{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Caddy")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &CaddyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			ingress := &networkingv1.Ingress{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, ingress)).To(Succeed())
			Expect(ingress.Spec.Rules).To(HaveLen(1))
			Expect(ingress.Spec.Rules[0].Host).To(Equal("caddy.example.com"))
			Expect(ingress.Spec.Rules[0].HTTP.Paths).To(HaveLen(1))
			Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(Equal(resourceName))
			Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number).To(Equal(int32(80)))

			Expect(k8sClient.Get(ctx, typeNamespacedName, caddy)).To(Succeed())
			caddy.Spec.Domain = ""
			Expect(k8sClient.Update(ctx, caddy)).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Get(ctx, typeNamespacedName, ingress)).To(MatchError(errors.IsNotFound, "be a not found error"))

			content := "<h1>Hello from Caddy</h1>"
			caddy.Spec.Content = &content
			Expect(k8sClient.Update(ctx, caddy)).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())

			contentConfigMap := &corev1.ConfigMap{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, contentConfigMap)).To(Succeed())
			Expect(contentConfigMap.Data).To(HaveKeyWithValue("index.html", content))

			deployment := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, deployment)).To(Succeed())
			container := deployment.Spec.Template.Spec.Containers[0]
			Expect(container.Command).To(Equal([]string{"caddy"}))
			Expect(container.Args).To(Equal([]string{
				"file-server", "--root", "/var/www/html", "--listen", ":80",
			}))
			Expect(container.VolumeMounts).To(ContainElement(corev1.VolumeMount{
				Name: "web-content", MountPath: "/var/www/html", ReadOnly: true,
			}))
			Expect(deployment.Spec.Template.Spec.Volumes).To(ContainElement(corev1.Volume{
				Name: "web-content",
				VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: resourceName},
					Items:                []corev1.KeyToPath{{Key: "index.html", Path: "index.html"}},
				}},
			}))

			caddy.Spec.Content = nil
			Expect(k8sClient.Update(ctx, caddy)).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Get(ctx, typeNamespacedName, contentConfigMap)).To(MatchError(errors.IsNotFound, "be a not found error"))
			Expect(k8sClient.Get(ctx, typeNamespacedName, deployment)).To(Succeed())
			Expect(deployment.Spec.Template.Spec.Containers[0].Command).To(BeEmpty())
			Expect(deployment.Spec.Template.Spec.Containers[0].Args).To(BeEmpty())
			Expect(deployment.Spec.Template.Spec.Volumes).To(BeEmpty())
		})
	})
})
