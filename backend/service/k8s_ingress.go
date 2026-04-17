package service

import (
	"context"
	"time"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"devops-backend/utils"
)

type IngressInfo struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	IngressClass   string            `json:"ingress_class"`
	GatewayAddress string            `json:"gateway_address"`
	Rules          []IngressRule     `json:"rules"`
	Labels         map[string]string `json:"labels"`
	CreatedAt      time.Time         `json:"created_at"`
	RuleCount      int               `json:"rule_count"`
	Hosts          []string          `json:"hosts"`
}

type IngressRule struct {
	Host  string        `json:"host"`
	Paths []IngressPath `json:"paths"`
}

type IngressPath struct {
	Path        string `json:"path"`
	PathType    string `json:"path_type"`
	ServiceName string `json:"service_name"`
	ServicePort int    `json:"service_port"`
}

type K8sIngress struct{}

func (s *K8sIngress) GetIngresses(clusterID uint, namespace string) ([]IngressInfo, error) {
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

	ingressList, err := client.Clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	ingresses := make([]IngressInfo, 0, len(ingressList.Items))
	for _, ing := range ingressList.Items {
		rules := make([]IngressRule, 0)
		hosts := make([]string, 0)

		for _, rule := range ing.Spec.Rules {
			paths := make([]IngressPath, 0)
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					port := 0
					if path.Backend.Service.Port.Number != 0 {
						port = int(path.Backend.Service.Port.Number)
					}

					pathType := ""
					if path.PathType != nil {
						pathType = string(*path.PathType)
					}

					paths = append(paths, IngressPath{
						Path:        path.Path,
						PathType:    pathType,
						ServiceName: path.Backend.Service.Name,
						ServicePort: port,
					})
				}
			}

			rules = append(rules, IngressRule{
				Host:  rule.Host,
				Paths: paths,
			})

			if rule.Host != "" {
				hosts = append(hosts, rule.Host)
			}
		}

		ingressClass := ""
		if ing.Spec.IngressClassName != nil {
			ingressClass = *ing.Spec.IngressClassName
		}

		gatewayAddress := ""
		if len(ing.Status.LoadBalancer.Ingress) > 0 {
			lbIngress := ing.Status.LoadBalancer.Ingress[0]
			if lbIngress.IP != "" {
				gatewayAddress = lbIngress.IP
			} else if lbIngress.Hostname != "" {
				gatewayAddress = lbIngress.Hostname
			}
		}
		if gatewayAddress == "" {
			gatewayAddress = "-"
		}

		ingresses = append(ingresses, IngressInfo{
			Name:           ing.Name,
			Namespace:      ing.Namespace,
			IngressClass:   ingressClass,
			GatewayAddress: gatewayAddress,
			Rules:          rules,
			Labels:         ing.Labels,
			CreatedAt:      ing.CreationTimestamp.Time,
			RuleCount:      len(ing.Spec.Rules),
			Hosts:          hosts,
		})
	}

	return ingresses, nil
}

func (s *K8sIngress) GetIngress(clusterID uint, namespace, name string) (*networkingv1.Ingress, error) {
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

	ingress, err := client.Clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ingress, nil
}

func (s *K8sIngress) DeleteIngress(clusterID uint, namespace, name string) error {
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

	return client.Clientset.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *K8sIngress) CreateIngress(clusterID uint, namespace string, ingress *networkingv1.Ingress) error {
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

	_, err = client.Clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	return err
}

func (s *K8sIngress) UpdateIngress(clusterID uint, namespace, name string, ingress *networkingv1.Ingress) error {
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

	_, err = client.Clientset.NetworkingV1().Ingresses(namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	return err
}

func (s *K8sIngress) GetIngressYAML(clusterID uint, namespace, name string) (string, error) {
	ingress, err := s.GetIngress(clusterID, namespace, name)
	if err != nil {
		return "", err
	}

	ingress.APIVersion = "networking.k8s.io/v1"
	ingress.Kind = "Ingress"

	yamlBytes, err := yaml.Marshal(ingress)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

func (s *K8sIngress) UpdateIngressYAML(clusterID uint, namespace, name string, yamlStr string) error {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return err
	}

	var ingress networkingv1.Ingress
	if err := yaml.Unmarshal([]byte(yamlStr), &ingress); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.Clientset.NetworkingV1().Ingresses(namespace).Update(ctx, &ingress, metav1.UpdateOptions{})
	return err
}

type IngressEventInfo struct {
	Type    string    `json:"type"`
	Reason  string    `json:"reason"`
	Time    time.Time `json:"time"`
	Source  string    `json:"source"`
	Message string    `json:"message"`
}

func (s *K8sIngress) GetIngressEvents(clusterID uint, namespace, name string) ([]IngressEventInfo, error) {
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
		FieldSelector: "involvedObject.name=" + name + ",involvedObject.kind=Ingress",
	})
	if err != nil {
		return nil, err
	}

	eventInfos := make([]IngressEventInfo, 0, len(events.Items))
	for _, event := range events.Items {
		source := ""
		if event.Source.Component != "" {
			source = event.Source.Component
		}
		if event.Source.Host != "" {
			source += "/" + event.Source.Host
		}

		eventInfos = append(eventInfos, IngressEventInfo{
			Type:    event.Type,
			Reason:  event.Reason,
			Time:    event.FirstTimestamp.Time,
			Source:  source,
			Message: event.Message,
		})
	}

	return eventInfos, nil
}
