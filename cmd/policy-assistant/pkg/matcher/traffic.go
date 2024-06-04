package matcher

import (
	"fmt"
	"strings"

	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
)

type Traffic struct {
	Source      *TrafficPeer
	Destination *TrafficPeer

	ResolvedPort     int
	ResolvedPortName string
	Protocol         v1.Protocol
}

func (t *Traffic) Table() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)

	pp := fmt.Sprintf("%d (%s) on %s", t.ResolvedPort, t.ResolvedPortName, t.Protocol)
	table.SetHeader([]string{"Port/Protocol", "Source/Dest", "Pod IP", "Namespace", "NS Labels", "Pod Labels"})
	
	source := []string{pp, "source", t.Source.IP}
	if t.Source.Internal != nil {
		i := t.Source.Internal
		source = append(source, i.Namespace, labelsToString(i.NamespaceLabels), labelsToString(i.PodLabels))
	} else {
		source = append(source, "", "", "")
	}
	table.Append(source)

	dest := []string{pp, "destination", t.Destination.IP}
	if t.Destination.Internal != nil {
		i := t.Destination.Internal
		dest = append(dest, i.Namespace, labelsToString(i.NamespaceLabels), labelsToString(i.PodLabels))
	} else {
		dest = append(dest, "", "", "")
	}
	table.Append(dest)

	table.Render()
	return tableString.String()
}

func labelsToString(labels map[string]string) string {
	format := func(k string) string { return fmt.Sprintf("%s: %s", k, labels[k]) }
	return strings.Join(slice.Map(format, slice.Sort(maps.Keys(labels))), "\n")
}

type TrafficPeer struct {
	Internal *InternalPeer
	// IP external to cluster
	IP string
}

func (p *TrafficPeer) Namespace() string {
	if p.Internal == nil {
		return ""
	}
	return p.Internal.Namespace
}

func (p *TrafficPeer) IsExternal() bool {
	return p.Internal == nil
}

func (p *TrafficPeer) Translate() TrafficPeer {
	var podsNetworking []*PodNetworking
	var podLabels map[string]string
	var namespaceLabels map[string]string
	var workloadOwner string
	workloadOwnerExists := false
	fmt.Println(p.Internal.Workload)
	workloadMetadata := strings.Split(strings.ToLower(p.Internal.Workload), "/")
	if len(workloadMetadata) != 3 || (workloadMetadata[0] == "" || workloadMetadata[1] == "" || workloadMetadata[2] == "") || (workloadMetadata[1] != "daemonset" && workloadMetadata[1] != "statefulset" && workloadMetadata[1] != "replicaset" && workloadMetadata[1] != "deployment" && workloadMetadata[1] != "pod") {
		logrus.Fatalf("Bad Workload structure: Types supported are pod, replicaset, deployment, daemonset, statefulset, and 3 fields are required with this structure, <namespace>/<workloadType>/<workloadName>")
	}
	kubeClient, err := kube.NewKubernetesForContext("")
	utils.DoOrDie(err)
	ns, err := kubeClient.GetNamespace(workloadMetadata[0])
	utils.DoOrDie(err)
	kubePods, err := kube.GetPodsInNamespaces(kubeClient, []string{workloadMetadata[0]})
	if err != nil {
		logrus.Fatalf("unable to read pods from kube, ns '%s': %+v", workloadMetadata[0], err)
	}

	for _, pod := range kubePods {
		if workloadMetadata[1] == "daemonset" || workloadMetadata[1] == "statefulset" || workloadMetadata[1] == "replicaset" {
			workloadOwner = pod.OwnerReferences[0].Name
		} else if workloadMetadata[1] == "deployment" {
			kubeReplicaSets, err := kubeClient.GetReplicaSet(workloadMetadata[0], pod.OwnerReferences[0].Name)
			if err != nil {
				logrus.Fatalf("unable to read Replicaset from kube, rs '%s': %+v", pod.OwnerReferences[0].Name, err)
			}
			workloadOwner = kubeReplicaSets.OwnerReferences[0].Name
		} else if workloadMetadata[1] == "pod" {
			workloadOwner = pod.Name
		}
		if workloadOwner == workloadMetadata[2] {
			podLabels = pod.Labels
			namespaceLabels = ns.Labels
			podNetworking := PodNetworking{
				IP: pod.Status.PodIP,
			}
			podsNetworking = append(podsNetworking, &podNetworking)
			workloadOwnerExists = true

		}
	}

	if !workloadOwnerExists {
		logrus.Fatalf("workload not found on the cluster")
	}

	InternalPeer := InternalPeer{
		Workload:        p.Internal.Workload,
		PodLabels:       podLabels,
		NamespaceLabels: namespaceLabels,
		Namespace:       workloadMetadata[0],
		Pods:            podsNetworking,
	}

	TranslatedPeer := TrafficPeer{
		Internal: &InternalPeer,
	}
	return TranslatedPeer
}

func DeploymentsToTrafficPeers() []TrafficPeer {
	var deploymentPeers []TrafficPeer
	kubeClient, err := kube.NewKubernetesForContext("")
	utils.DoOrDie(err)
	kubeNamespaces, err := kubeClient.GetAllNamespaces()
	if err != nil {
		logrus.Fatalf("unable to read namespaces from kube: %+v", err)
	}

	for _, namespace := range kubeNamespaces.Items {
		kubeDeployments, err := kubeClient.GetDeploymentsInNamespace(namespace.Name)
		if err != nil {
			logrus.Fatalf("unable to read deployments from kube, ns '%s': %+v", namespace.Name, err)
		}
		for _, deployment := range kubeDeployments {
			fmt.Println(namespace.Name+"/deployment/"+deployment.Name)
			TmpInternalPeer := InternalPeer{
				Workload: namespace.Name+"/deployment/"+deployment.Name,
			}
			TmpPeer := TrafficPeer{
				Internal: &TmpInternalPeer,
			}
			deploymentPeers = append(deploymentPeers, TmpPeer.Translate())
		}

	}

	return deploymentPeers
}

func DaemonSetsToTrafficPeers() []TrafficPeer {
	var daemonSetPeers []TrafficPeer
	kubeClient, err := kube.NewKubernetesForContext("")
	utils.DoOrDie(err)
	kubeNamespaces, err := kubeClient.GetAllNamespaces()
	if err != nil {
		logrus.Fatalf("unable to read namespaces from kube: %+v", err)
	}

	for _, namespace := range kubeNamespaces.Items {
		kubeDaemonSets, err := kubeClient.GetDaemonSetsInNamespace(namespace.Name)
		if err != nil {
			logrus.Fatalf("unable to read daemonSets from kube, ns '%s': %+v", namespace.Name, err)
		}
		for _, daemonSet := range kubeDaemonSets {
			TmpInternalPeer := InternalPeer{
				Workload: namespace.Name+"/daemonset/"+daemonSet.Name,
			}
			TmpPeer := TrafficPeer{
				Internal: &TmpInternalPeer,
			}
			daemonSetPeers = append(daemonSetPeers, TmpPeer.Translate())
		}

	}

	return daemonSetPeers
}

// Internal to cluster
type InternalPeer struct {
	// optional: if set, will override remaining values with information from cluster
	Workload        string
	PodLabels       map[string]string
	NamespaceLabels map[string]string
	Namespace       string
	// optional
	Pods []*PodNetworking
}

type PodNetworking struct {
	IP string
	// don't worry about populating below fields right now
	IsHostNetworking bool
	NodeLabels       []string
}
