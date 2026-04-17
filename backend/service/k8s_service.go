package service

import (
	"context"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"

	"devops-backend/utils"
)

type ServiceInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"cluster_ip"`
	Ports     []ServicePort     `json:"ports"`
	Selector  map[string]string `json:"selector"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
}

type ServicePort struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
	Protocol   string `json:"protocol"`
}

type K8sService struct{}

func (s *K8sService) GetServices(clusterID uint, namespace string) ([]ServiceInfo, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	svcList, err := client.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	services := make([]ServiceInfo, 0, len(svcList.Items))
	for _, svc := range svcList.Items {
		ports := make([]ServicePort, 0)
		for _, p := range svc.Spec.Ports {
			targetPort := ""
			if p.TargetPort.IntVal != 0 {
				targetPort = strconv.Itoa(int(p.TargetPort.IntVal))
			} else if p.TargetPort.StrVal != "" {
				targetPort = p.TargetPort.StrVal
			}

			ports = append(ports, ServicePort{
				Name:       p.Name,
				Port:       p.Port,
				TargetPort: targetPort,
				Protocol:   string(p.Protocol),
			})
		}

		services = append(services, ServiceInfo{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Type:      string(svc.Spec.Type),
			ClusterIP: svc.Spec.ClusterIP,
			Ports:     ports,
			Selector:  svc.Spec.Selector,
			Labels:    svc.Labels,
			CreatedAt: svc.CreationTimestamp.Time,
		})
	}

	return services, nil
}

func (s *K8sService) GetService(clusterID uint, namespace, name string) (*corev1.Service, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	svc, err := client.Clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *K8sService) DeleteService(clusterID uint, namespace, name string) error {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.Clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *K8sService) CreateService(clusterID uint, namespace string, svc *corev1.Service) error {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.Clientset.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
	return err
}

func (s *K8sService) GetServiceDetail(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	svc, err := s.GetService(clusterID, namespace, name)
	if err != nil {
		return nil, err
	}

	detail := map[string]interface{}{
		"name":             svc.Name,
		"namespace":        svc.Namespace,
		"type":             string(svc.Spec.Type),
		"cluster_ip":       svc.Spec.ClusterIP,
		"external_ips":     svc.Spec.ExternalIPs,
		"selector":         svc.Spec.Selector,
		"labels":           svc.Labels,
		"annotations":      svc.Annotations,
		"created_at":       svc.CreationTimestamp.Time,
		"ports":            s.parsePorts(svc.Spec.Ports),
		"session_affinity": string(svc.Spec.SessionAffinity),
	}

	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		externalIPs := make([]string, 0)
		for _, ing := range svc.Status.LoadBalancer.Ingress {
			if ing.IP != "" {
				externalIPs = append(externalIPs, ing.IP)
			} else if ing.Hostname != "" {
				externalIPs = append(externalIPs, ing.Hostname)
			}
		}
		detail["external_ip"] = externalIPs
	}

	if svc.Spec.Type == corev1.ServiceTypeNodePort {
		nodePorts := make([]map[string]interface{}, 0)
		for _, port := range svc.Spec.Ports {
			nodePorts = append(nodePorts, map[string]interface{}{
				"name":        port.Name,
				"port":        port.Port,
				"target_port": s.parseTargetPort(port.TargetPort),
				"node_port":   port.NodePort,
				"protocol":    string(port.Protocol),
			})
		}
		detail["node_ports"] = nodePorts
	}

	return detail, nil
}

func (s *K8sService) GetServicePods(clusterID uint, namespace string, selector map[string]string) ([]map[string]interface{}, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{
		MatchLabels: selector,
	})

	podList, err := client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	pods := make([]map[string]interface{}, 0)
	for _, pod := range podList.Items {
		podIP := pod.Status.PodIP
		if podIP == "" {
			podIP = "-"
		}

		pods = append(pods, map[string]interface{}{
			"name":       pod.Name,
			"namespace":  pod.Namespace,
			"status":     string(pod.Status.Phase),
			"pod_ip":     podIP,
			"node_name":  pod.Spec.NodeName,
			"created_at": pod.CreationTimestamp.Time,
			"labels":     pod.Labels,
		})
	}

	return pods, nil
}

func (s *K8sService) GetServiceEvents(clusterID uint, namespace, name string) ([]map[string]interface{}, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	events, err := client.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: "involvedObject.name=" + name + ",involvedObject.kind=Service",
	})
	if err != nil {
		return nil, err
	}

	eventList := make([]map[string]interface{}, 0)
	for _, event := range events.Items {
		eventList = append(eventList, map[string]interface{}{
			"type":      event.Type,
			"reason":    event.Reason,
			"message":   event.Message,
			"count":     event.Count,
			"timestamp": event.LastTimestamp.Time,
			"source":    event.Source.Component,
		})
	}

	return eventList, nil
}

func (s *K8sService) parsePorts(ports []corev1.ServicePort) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, port := range ports {
		result = append(result, map[string]interface{}{
			"name":        port.Name,
			"port":        port.Port,
			"target_port": s.parseTargetPort(port.TargetPort),
			"protocol":    string(port.Protocol),
			"node_port":   port.NodePort,
		})
	}
	return result
}

func (s *K8sService) parseTargetPort(targetPort intstr.IntOrString) string {
	if targetPort.IntVal != 0 {
		return strconv.Itoa(int(targetPort.IntVal))
	}
	return targetPort.StrVal
}

func (s *K8sService) GetServiceDeployments(clusterID uint, namespace string, selector map[string]string) ([]map[string]interface{}, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var deployList *appsv1.DeploymentList
	deployList, err = client.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	deployments := make([]map[string]interface{}, 0)
	for _, deploy := range deployList.Items {
		if s.matchSelector(deploy.Spec.Template.Labels, selector) {
			status := "未知"
			if deploy.Status.Replicas == deploy.Status.ReadyReplicas {
				status = "运行中"
			} else if deploy.Status.Replicas > deploy.Status.ReadyReplicas {
				status = "更新中"
			} else if deploy.Status.ReadyReplicas == 0 && deploy.Status.Replicas > 0 {
				status = "异常"
			}

			images := make([]string, 0)
			for _, container := range deploy.Spec.Template.Spec.Containers {
				images = append(images, container.Image)
			}

			deployments = append(deployments, map[string]interface{}{
				"name":           deploy.Name,
				"namespace":      deploy.Namespace,
				"status":         status,
				"replicas":       deploy.Status.Replicas,
				"ready_replicas": deploy.Status.ReadyReplicas,
				"images":         images,
				"labels":         deploy.Labels,
				"created_at":     deploy.CreationTimestamp.Time,
			})
		}
	}

	return deployments, nil
}

func (s *K8sService) matchSelector(labels map[string]string, selector map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for key, value := range selector {
		if labels[key] != value {
			return false
		}
	}
	return true
}

func (s *K8sService) GetServiceYAML(clusterID uint, namespace, name string) (string, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return "", err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	svc, err := client.Clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	svc.TypeMeta = metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	}

	yamlData, err := yaml.Marshal(svc)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

func (s *K8sService) UpdateServiceYAML(clusterID uint, namespace, name string, yamlStr string) error {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var svc corev1.Service
	if err := yaml.Unmarshal([]byte(yamlStr), &svc); err != nil {
		return err
	}

	_, err = client.Clientset.CoreV1().Services(namespace).Update(ctx, &svc, metav1.UpdateOptions{})
	return err
}
