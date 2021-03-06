package serviceipcidr

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

// GetServiceIPCIDR 从api-server组件对象中获得service的网段范围并返回.
// kuber并没有提供接口获得这样的信息, 这里借鉴了`kubectl cluster-info`的做法.
func GetServiceIPCIDR() (serviceIPCIDR string, err error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("failed to get incluster config: %s", err)
		return "", err
	}
	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("failed to get kubernetes client: %s", err)
		return "", err
	}

	// 不能简单地用Get("api-server"), 因为Pod名称中还会加一些额外字符串, 比如hostname.
	// 之前用的是 metav1.LabelSelector{}, 然后转换成String(), 但那只是简单的Marshal(),
	// 实际上应该使用 labels.Set{} 结构
	labelSet := labels.Set{
		"component": "kube-apiserver",
	}
	podListOpts := metav1.ListOptions{
		LabelSelector: labelSet.String(),
	}
	podList, err := client.CoreV1().Pods("kube-system").List(podListOpts)
	if err != nil {
		klog.Errorf("failed to list pod: %s", err)
		return "", err
	}
	pod := podList.Items[0]
	cmds := pod.Spec.Containers[0].Command
	for _, item := range cmds {
		if strings.Contains(item, "service-cluster-ip-range") {
			serviceIPCIDR = strings.Split(item, "=")[1]
			break
		}
	}

	return
}
