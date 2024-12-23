package kube

import (
	"bytes"
	"context"

	v1alpha12 "sigs.k8s.io/network-policy-api/apis/v1alpha1"
	"sigs.k8s.io/network-policy-api/pkg/client/clientset/versioned/typed/apis/v1alpha1"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

var ErrNotImplemented = errors.New("Not implemented")

type Kubernetes struct {
	ClientSet      *kubernetes.Clientset
	alphaClientSet *v1alpha1.PolicyV1alpha1Client
	RestConfig     *rest.Config
}

func NewKubernetesForContext(context string) (*Kubernetes, error) {
	logrus.Debugf("instantiating k8s Clientset for context %s", context)
	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to build config")
	}
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to instantiate Clientset")
	}
	alphacClientset, err := v1alpha1.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to instantiate alpha network client set")
	}

	return &Kubernetes{
		ClientSet:      clientset,
		alphaClientSet: alphacClientset,
		RestConfig:     kubeConfig,
	}, nil
}

func (k *Kubernetes) GetNamespace(namespace string) (*v1.Namespace, error) {
	ns, err := k.ClientSet.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	return ns, errors.Wrapf(err, "unable to get namespace %s", namespace)
}

func (k *Kubernetes) GetAllNamespaces() (*v1.NamespaceList, error) {
	nsList, err := k.ClientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	return nsList, errors.Wrapf(err, "unable to list namespaces")
}

func (k *Kubernetes) SetNamespaceLabels(namespace string, labels map[string]string) (*v1.Namespace, error) {
	ns, err := k.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	ns.Labels = labels
	_, err = k.ClientSet.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	return ns, errors.Wrapf(err, "unable to update namespace %s", namespace)
}

func (k *Kubernetes) DeleteNamespace(ns string) error {
	err := k.ClientSet.CoreV1().Namespaces().Delete(context.TODO(), ns, metav1.DeleteOptions{})
	return errors.Wrapf(err, "unable to delete namespace %s", ns)
}

func (k *Kubernetes) CreateNamespace(ns *v1.Namespace) (*v1.Namespace, error) {
	nsr, err := k.ClientSet.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	return nsr, errors.Wrapf(err, "unable to create namespace %s", ns.Name)
}

func (k *Kubernetes) DeleteAllNetworkPoliciesInNamespace(ns string) error {
	logrus.Debugf("deleting all network policies in namespace %s", ns)
	netpols, err := k.ClientSet.NetworkingV1().NetworkPolicies(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to list network policies in ns %s", ns)
	}
	for _, np := range netpols.Items {
		logrus.Debugf("deleting network policy %s/%s", ns, np.Name)
		err = k.DeleteNetworkPolicy(np.Namespace, np.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *Kubernetes) DeleteNetworkPolicy(ns string, name string) error {
	err := k.ClientSet.NetworkingV1().NetworkPolicies(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	return errors.Wrapf(err, "unable to delete network policy %s/%s", ns, name)
}

func (k *Kubernetes) GetNetworkPoliciesInNamespace(ctx context.Context, namespace string) ([]networkingv1.NetworkPolicy, error) {
	netpolList, err := k.ClientSet.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get netpols in namespace %s", namespace)
	}
	return netpolList.Items, nil
}

func (k *Kubernetes) GetAdminNetworkPolicies(ctx context.Context) ([]v1alpha12.AdminNetworkPolicy, error) {
	anps, err := k.alphaClientSet.AdminNetworkPolicies().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return anps.Items, nil
}

func (k *Kubernetes) CreateAdminNetworkPolicy(ctx context.Context, policy *v1alpha12.AdminNetworkPolicy) (*v1alpha12.AdminNetworkPolicy, error) {
	return nil, ErrNotImplemented
}

func (k *Kubernetes) UpdateAdminNetworkPolicy(ctx context.Context, policy *v1alpha12.AdminNetworkPolicy) (*v1alpha12.AdminNetworkPolicy, error) {
	return nil, ErrNotImplemented
}

func (k *Kubernetes) DeleteAdminNetworkPolicy(ctx context.Context, name string) error {
	//TODO: implement
	return ErrNotImplemented
}

func (k *Kubernetes) GetBaselineAdminNetworkPolicy(ctx context.Context) (*v1alpha12.BaselineAdminNetworkPolicy, error) {
	banps, err := k.alphaClientSet.BaselineAdminNetworkPolicies().List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}
	if len(banps.Items) == 1 {
		return &banps.Items[0], nil
	}
	return nil, nil
}

func (k *Kubernetes) CreateBaselineAdminNetworkPolicy(ctx context.Context, policy *v1alpha12.BaselineAdminNetworkPolicy) (*v1alpha12.BaselineAdminNetworkPolicy, error) {
	return nil, ErrNotImplemented
}

func (k *Kubernetes) UpdateBaselineAdminNetworkPolicy(ctx context.Context, policy *v1alpha12.BaselineAdminNetworkPolicy) (*v1alpha12.BaselineAdminNetworkPolicy, error) {
	return nil, ErrNotImplemented
}

func (k *Kubernetes) DeleteBaselineAdminNetworkPolicy(ctx context.Context, name string) error {
	//TODO: implement
	return ErrNotImplemented
}

func (k *Kubernetes) UpdateNetworkPolicy(policy *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	logrus.Debugf("updating network policy %s/%s", policy.Namespace, policy.Name)
	np, err := k.ClientSet.NetworkingV1().NetworkPolicies(policy.Namespace).Update(context.TODO(), policy, metav1.UpdateOptions{})
	return np, errors.Wrapf(err, "unable to update network policy %s/%s", policy.Namespace, policy.Name)
}

func (k *Kubernetes) CreateNetworkPolicy(policy *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	logrus.Debugf("creating network policy %s/%s", policy.Namespace, policy.Name)

	createdPolicy, err := k.ClientSet.NetworkingV1().NetworkPolicies(policy.Namespace).Create(context.TODO(), policy, metav1.CreateOptions{})
	return createdPolicy, errors.Wrapf(err, "unable to create network policy %s/%s", policy.Namespace, policy.Name)
}

func (k *Kubernetes) GetDeploymentsInNamespace(namespace string) ([]appsv1.Deployment, error) {
	deploymentList, err := k.ClientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get deployments in namespace %s", namespace)
	}
	return deploymentList.Items, nil
}

func (k *Kubernetes) GetDaemonSetsInNamespace(namespace string) ([]appsv1.DaemonSet, error) {
	daemonSetList, err := k.ClientSet.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get daemonSets in namespace %s", namespace)
	}
	return daemonSetList.Items, nil
}

func (k *Kubernetes) GetStatefulSetsInNamespace(namespace string) ([]appsv1.StatefulSet, error) {
	statefulSetList, err := k.ClientSet.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get StatefulSets in namespace %s", namespace)
	}
	return statefulSetList.Items, nil
}

func (k *Kubernetes) GetReplicaSetsInNamespace(namespace string) ([]appsv1.ReplicaSet, error) {
	replicaSetList, err := k.ClientSet.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get ReplicaSets in namespace %s", namespace)
	}
	return replicaSetList.Items, nil
}

func (k *Kubernetes) GetReplicaSet(namespace string, name string) (*appsv1.ReplicaSet, error) {
	replicaSet, err := k.ClientSet.AppsV1().ReplicaSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return replicaSet, errors.Wrapf(err, "unable to get replicaSet %s/%s", namespace, name)
}

func (k *Kubernetes) GetService(namespace string, name string) (*v1.Service, error) {
	service, err := k.ClientSet.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return service, errors.Wrapf(err, "unable to get service %s/%s", namespace, name)
}

func (k *Kubernetes) CreateService(svc *v1.Service) (*v1.Service, error) {
	ns := svc.Namespace
	logrus.Debugf("creating service %s/%s", ns, svc.Name)
	createdService, err := k.ClientSet.CoreV1().Services(ns).Create(context.TODO(), svc, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create service %s/%s", ns, svc.Name)
	}
	return createdService, nil
}

func (k *Kubernetes) DeleteService(namespace string, name string) error {
	logrus.Debugf("deleting service %s/%s", namespace, name)
	err := k.ClientSet.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	return errors.Wrapf(err, "unable to delete service %s/%s", namespace, name)
}

func (k *Kubernetes) GetServicesInNamespace(namespace string) ([]v1.Service, error) {
	serviceList, err := k.ClientSet.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get services in namespace %s", namespace)
	}
	return serviceList.Items, nil
}

func (k *Kubernetes) GetPodsInNamespace(namespace string) ([]v1.Pod, error) {
	podList, err := k.ClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get pods in namespace %s", namespace)
	}
	return podList.Items, nil
}

func (k *Kubernetes) GetPod(namespace string, podName string) (*v1.Pod, error) {
	pod, err := k.ClientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	return pod, errors.Wrapf(err, "unable to get pod %s/%s", namespace, podName)
}

func (k *Kubernetes) SetPodLabels(namespace string, podName string, labels map[string]string) (*v1.Pod, error) {
	pod, err := k.GetPod(namespace, podName)
	if err != nil {
		return nil, err
	}
	pod.Labels = labels
	updatedPod, err := k.ClientSet.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	return updatedPod, errors.Wrapf(err, "unable to update pod %s/%s", namespace, podName)
}

func (k *Kubernetes) CreatePod(pod *v1.Pod) (*v1.Pod, error) {
	ns := pod.Namespace
	logrus.Debugf("creating pod %s/%s", ns, pod.Name)

	createdPod, err := k.ClientSet.CoreV1().Pods(ns).Create(context.TODO(), pod, metav1.CreateOptions{})
	return createdPod, errors.Wrapf(err, "unable to create pod %s/%s", ns, pod.Name)
}

func (k *Kubernetes) DeletePod(namespace string, podName string) error {
	logrus.Debugf("deleting pod %s/%s", namespace, podName)
	err := k.ClientSet.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	return errors.Wrapf(err, "unable to delete pod %s/%s", namespace, podName)
}

// ExecuteRemoteCommand executes a remote shell command on the given pod
// returns the output from stdout and stderr
func (k *Kubernetes) ExecuteRemoteCommand(namespace string, pod string, container string, command []string) (string, string, error, error) {
	request := k.ClientSet.
		CoreV1().
		RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(pod).
		SubResource("exec").
		Param("container", container).
		//Timeout(5*time.Second). // TODO this seems to not do anything ... why ?
		VersionedParams(
			&v1.PodExecOptions{
				Container: container,
				Command:   command,
				Stdin:     false,
				Stdout:    true,
				Stderr:    true,
				TTY:       true,
			},
			scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.RestConfig, "POST", request.URL())
	if err != nil {
		return "", "", nil, errors.Wrapf(err, "unable to instantiate SPDYExecutor")
	}

	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	})

	out, errOut := buf.String(), errBuf.String()
	return out, errOut, errors.Wrapf(err, "unable to stream command"), nil
}
