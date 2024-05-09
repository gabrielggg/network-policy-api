package matcher

import (
	"fmt"
	"strings"

	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/olekukonko/tablewriter"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/utils"
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
	
	fmt.Printf(p.Internal.Workload)
	workloadMetadata := strings.Split(p.Internal.Workload, "/")
	fmt.Printf(workloadMetadata[0])
	kubeClient, err := kube.NewKubernetesForContext("")
	utils.DoOrDie(err)
	if err != nil {
		logrus.Errorf("unable to read ReplicaSet from kube, ns '%s': %+v", "default", err)
	}
	ns, err := kubeClient.GetNamespace(workloadMetadata[0])
	utils.DoOrDie(err)
	kubePods, err := kube.GetPodsInNamespaces(kubeClient, []string{workloadMetadata[0]})
	if err != nil {
		logrus.Errorf("unable to read pods from kube, ns '%s': %+v", workloadMetadata[0], err)
	}
	for _, pod := range kubePods {
		workloadOwner := ""
		if workloadMetadata[1] == "daemonset" || workloadMetadata[1] == "statefulset" {
			workloadOwner = pod.OwnerReferences[0].Name
		} else {
			kubeReplicaSets, err := kubeClient.GetReplicaSet(workloadMetadata[0], pod.OwnerReferences[0].Name)
			if err != nil {
				logrus.Errorf("unable to read Replicaset from kube, rs '%s': %+v", pod.OwnerReferences[0].Name, err)
			}
			workloadOwner = kubeReplicaSets.OwnerReferences[0].Name
		}
		if workloadOwner == workloadMetadata[2] {
			podLabels = pod.Labels
			namespaceLabels = ns.Labels
			podNetworking := PodNetworking{
		                IP: pod.Status.PodIP,
		        }
			podsNetworking = append(podsNetworking, podNetworking)
			
		} else {
			
		}
		
	}

	InternalPeer := InternalPeer{
		Workload: p.Internal.Workload,
		PodLabels: podLabels,
		NamespaceLabels: namespaceLabels,
		Namespace: workloadMetadata[0],
		Pods: podsNetworking,
	}
		
	TranslatedPeer := TrafficPeer{
		Internal: &InternalPeer,
        }
	return TranslatedPeer
}

type TrafficPeer struct {
	Internal *InternalPeer
        // IP external to cluster
	IP          string
}

// Internal to cluster
type InternalPeer struct {
        // optional: if set, will override remaining values with information from cluster
        Workload string

	PodLabels       map[string]string
	NamespaceLabels map[string]string
	Namespace       string
        // optional
        Pods      []*PodNetworking
}

type PodNetworking struct {
       IP string
      // don't worry about populating below fields right now
       IsHostNetworking bool
       NodeLabels []string
}
