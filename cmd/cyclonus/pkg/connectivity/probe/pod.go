package probe

import (
	"fmt"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	agnhostImage = "k8s.gcr.io/e2e-test-images/agnhost:2.28"
)

func NewPod(ns string, name string, labels map[string]string, ip string, containers []*Container) *Pod {
	return &Pod{
		Namespace:  ns,
		Name:       name,
		Labels:     labels,
		IP:         ip,
		Containers: containers,
	}
}

func NewDefaultPod(ns string, name string, labels map[string]string, ip string, ports []int, protocols []v1.Protocol) *Pod {
	var containers []*Container
	for _, port := range ports {
		for _, protocol := range protocols {
			containers = append(containers, NewDefaultContainer(port, protocol))
		}
	}
	return &Pod{
		Namespace:  ns,
		Name:       name,
		Labels:     labels,
		IP:         ip,
		Containers: containers,
	}
}

type Pod struct {
	Namespace  string
	Name       string
	Labels     map[string]string
	IP         string
	Containers []*Container
}

func (p *Pod) ServiceName() string {
	return fmt.Sprintf("s-%s-%s", p.Namespace, p.Name)
}

func (p *Pod) KubePod() *v1.Pod {
	zero := int64(0)
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Labels:    p.Labels,
			Namespace: p.Namespace,
		},
		Spec: v1.PodSpec{
			TerminationGracePeriodSeconds: &zero,
			Containers:                    p.KubeContainers(),
		},
	}
}

func (p *Pod) KubeService() *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.ServiceName(),
			Namespace: p.Namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: p.Labels,
		},
	}

	for _, cont := range p.Containers {
		service.Spec.Ports = append(service.Spec.Ports, cont.KubeServicePort())
	}

	return service
}

func (p *Pod) KubeContainers() []v1.Container {
	var containers []v1.Container
	for _, cont := range p.Containers {
		containers = append(containers, cont.KubeContainer())
	}
	return containers
}

func (p *Pod) ResolveNamedPort(port string) (int, error) {
	for _, c := range p.Containers {
		if c.PortName == port {
			return c.Port, nil
		}
	}
	return 0, errors.Errorf("unable to resolve named port %s on pod %s/%s", port, p.Namespace, p.Name)
}

func (p *Pod) ResolveNumberedPort(port int) (string, error) {
	for _, c := range p.Containers {
		if c.Port == port {
			return c.PortName, nil
		}
	}
	return "", errors.Errorf("unable to resolve numbered port %d on pod %s/%s", port, p.Namespace, p.Name)
}

func (p *Pod) IsServingPortProtocol(port int, protocol v1.Protocol) bool {
	for _, cont := range p.Containers {
		if cont.Port == port && cont.Protocol == protocol {
			return true
		}
	}
	return false
}

func (p *Pod) SetLabels(labels map[string]string) *Pod {
	return &Pod{
		Namespace:  p.Namespace,
		Name:       p.Name,
		Labels:     labels,
		IP:         p.IP,
		Containers: p.Containers,
	}
}

func (p *Pod) PodString() PodString {
	return NewPodString(p.Namespace, p.Name)
}

type Container struct {
	Name     string
	Port     int
	Protocol v1.Protocol
	PortName string
}

func NewDefaultContainer(port int, protocol v1.Protocol) *Container {
	return &Container{
		Name:     fmt.Sprintf("cont-%d-%s", port, strings.ToLower(string(protocol))),
		Port:     port,
		Protocol: protocol,
		PortName: fmt.Sprintf("serve-%d-%s", port, strings.ToLower(string(protocol))),
	}
}

func (c *Container) KubeServicePort() v1.ServicePort {
	return v1.ServicePort{
		Name:     fmt.Sprintf("service-port-%s-%d", strings.ToLower(string(c.Protocol)), c.Port),
		Protocol: c.Protocol,
		Port:     int32(c.Port),
	}
}

func (c *Container) KubeContainer() v1.Container {
	var cmd []string
	var env []v1.EnvVar

	switch c.Protocol {
	case v1.ProtocolTCP:
		cmd = []string{"/agnhost", "serve-hostname", "--tcp", "--http=false", "--port", fmt.Sprintf("%d", c.Port)}
	case v1.ProtocolUDP:
		cmd = []string{"/agnhost", "serve-hostname", "--udp", "--http=false", "--port", fmt.Sprintf("%d", c.Port)}
	case v1.ProtocolSCTP:
		//cmd = []string{"/agnhost", "netexec", "--sctp-port", fmt.Sprintf("%d", c.Port)}
		env = append(env, v1.EnvVar{
			Name:  fmt.Sprintf("SERVE_SCTP_PORT_%d", c.Port),
			Value: "foo",
		})
		cmd = []string{"/agnhost", "porter"}
	default:
		panic(errors.Errorf("invalid protocol %s", c.Protocol))
	}
	return v1.Container{
		Name:            c.Name,
		ImagePullPolicy: v1.PullIfNotPresent,
		Image:           agnhostImage,
		Command:         cmd,
		Env:             env,
		SecurityContext: &v1.SecurityContext{},
		Ports: []v1.ContainerPort{
			{
				ContainerPort: int32(c.Port),
				Name:          c.PortName,
				Protocol:      c.Protocol,
			},
		},
	}
}
