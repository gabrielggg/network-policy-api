package matcher

import (
	"fmt"
	"strings"

	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/olekukonko/tablewriter"
	"github.com/mattfenwick/cyclonus/pkg/kube"
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

//type TrafficPeer struct {
//	Internal *InternalPeer
//	IP       string
//}

func (p *TrafficPeer) Namespace() string {
	if p.Internal == nil {
		return ""
	}
	return p.Internal.Namespace
}

func (p *TrafficPeer) IsExternal() bool {
	return p.Internal == nil
}

func (p *TrafficPeer) HasWorkload() bool {
	return p.Workload != nil
}

func (p *TrafficPeer) Translate() TrafficPeer {
	fmt.Printf(p.Internal)

	//crear una lista de estos objectos
	podNetworking := PodNetworking{
                IP: //extract ip from pod
	      	// don't worry about populating below fields right now
	        IsHostNetworking: nil
	        NodeLabels: nil
        }
	//esta es la idea
	//podsNetworking.append(podNetworking)
	//iterar sobre esto para popular la lista dependiendo de la cantidad de pods del workload
	var podsNetworking []PodNetworking
	podsNetworking = append(podsNetworking, podNetworking)

	
	kubeClient, err := kube.NewKubernetesForContext("")
	utils.DoOrDie(err)
	if err != nil {
		logrus.Errorf("unable to read ReplicaSet from kube, ns '%s': %+v", "default", err)
	}
	ns, err := kubeClient.GetNamespace(p.Internal.Namespace)
	utils.DoOrDie(err)
	kubePods, err := kube.GetPodsInNamespaces(kubeClient, []string{p.Internal.Namespace})
	if err != nil {
		logrus.Errorf("unable to read pods from kube, ns '%s': %+v", "default", err)
	}
	for _, pod := range kubePods {
		workloadOwner := ""
		if k == "daemonset" || k == "statefulset" {
			workloadOwner = pod.OwnerReferences[0].Name
		} else {
			kubeReplicaSets, err := kubeClient.GetReplicaSet(trafficPeer.Internal.Namespace, pod.OwnerReferences[0].Name)
			if err != nil {
				logrus.Errorf("unable to read Replicaset from kube, ns '%s': %+v", "default", err)
			}
			workloadOwner = kubeReplicaSets.OwnerReferences[0].Name
		}
		if workloadOwner == v {
			trafficPeer.Internal.PodLabels = pod.Labels
			trafficPeer.Internal.NamespaceLabels = ns.Labels
			trafficPeer.IP = pod.Status.PodIP
		}
	}
	
	InternalPeer := InternalPeer{
		Workload: p.Internal.Workload,
                PodLabels: nil,
                NamespaceLabels: nil,
                Namespace: "tmpns",
		Pods: //generar data para pods lista de tipo podNetworking
        }

	

	
	TranslatedPeer := TrafficPeer{
		Internal: InternalPeer
	        // keep this field for backwards-compatibility or for IPs without internalPeer
		IP: nil
        }
	return TranslatedPeer
}



//type InternalPeer struct {
//	PodLabels       map[string]string
//	NamespaceLabels map[string]string
//	Namespace       string
//	NodeLabels      map[string]string
//	Node            string
//}

//////////

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
