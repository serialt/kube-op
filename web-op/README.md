
## kube-operator开发示例

mkdir web-op && cd web-op

### 1、初始化
```
kubebuilder init --domain=web.imau.cc  --repo=github.com/serialt/web-operator  --owner=serialt



#创建api
kubebuilder create api --group webapp --version v1 --kind Guestbook
 
# 下载依赖
go mod tidy




# 安装CRD资源到集群内
make install
[serialt@Krab web-op]🐳 make install
"/Users/serialt/github/kube-op/web-op/bin/controller-gen" rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
Downloading sigs.k8s.io/kustomize/kustomize/v5@v5.7.1
customresourcedefinition.apiextensions.k8s.io/redis.myapp.redis.imau.cc created
[serialt@Krab web-op]🐳 k get crd
NAME                                       CREATED AT
redis.myapp.redis.imau.cc                  


```

### 2、自定义CRD
```
// api/v1/guestbook_types.go
type GuestbookSpec struct {
	Replicas int32  `json:"replicas"` // 对应 YAML 的 `spec.replicas`
	Image    string `json:"image"`    // 对应 YAML 的 `spec.image`
	Port     int32  `json:"port"`     // 应用监听的端口
}

type GuestbookStatus struct {
    // +listType=map
	// +listMapKey=type
	// +optional
	Conditions        []metav1.Condition `json:"conditions,omitempty"`
	AvailableReplicas int32  
}
```

生成文件

```
[serialt@Krab web-op]🐳 make
```


### 3、实现业务逻辑

```go
import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "github.com/serialt/web-operator/api/v1"
)

func (r *GuestbookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	// 获取 CR 实例
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

	return ctrl.Result{}, nil
}
```

### 4、部署本地测试

```bash
# 生成CRD YAML
[serialt@Krab web-op]🐳 make manifests

# 部署本地集群
[serialt@Krab web-op]🐳 make install
"/Users/serialt/github/kube-op/web-op/bin/controller-gen" rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
Downloading sigs.k8s.io/kustomize/kustomize/v5@v5.7.1
customresourcedefinition.apiextensions.k8s.io/guestbooks.webapp.web.imau.cc created

# 本地
[serialt@Krab web-op]🐳 make install



# 测试的yaml
# config/samples/webapp_v1_guestbook.yaml
apiVersion: webapp.web.imau.cc/v1
kind: Guestbook
metadata:
  labels:
    app.kubernetes.io/name: web-op
    app.kubernetes.io/managed-by: kustomize
  name: guestbook-sample
spec:
  # TODO(user): Add fields here
  replicas: 3
  image: nginx:1.28-alpine
  port: 8080
  
# 执行后可以看到3个pod
[serialt@Krab web-op]🐳 kubectl apply -f   config/samples/webapp_v1_guestbook.yaml






```







