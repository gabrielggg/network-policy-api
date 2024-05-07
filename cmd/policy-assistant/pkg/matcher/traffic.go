package matcher

import (
	"fmt"
	"strings"

	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/olekukonko/tablewriter"
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

func (t *Traffic) Translate() string {

	pp := fmt.Sprintf("%d (%s) on %s", t.ResolvedPort, t.ResolvedPortName, t.Protocol)

	source := []string{pp, "source", t.Source.IP}
	if t.Source.Workload != nil {
		i := t.Source.Workload
		source = append(source, i.Namespace, labelsToString(i.NamespaceLabels), labelsToString(i.PodLabels))
	} else {
		source = append(source, "", "", "")
	}

	dest := []string{pp, "destination", t.Destination.IP}
	if t.Destination.Workload != nil {
		i := t.Destination.Workload
		dest = append(dest, i.Namespace, labelsToString(i.NamespaceLabels), labelsToString(i.PodLabels))
	} else {
		dest = append(dest, "", "", "")
	}
	
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

func (p *TrafficPeer) Translate() string {
	fmt.Printf(p.Workload.fullName)
	fmt.Printf(p.Internal)
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
       // keep this field for backwards-compatibility or for IPs without internalPeer
	IP          string
       // use this for pod IPs
       *Workload
}

type InternalPeer struct {
	PodLabels       map[string]string
	NamespaceLabels map[string]string
	Namespace       string
       // I believe I added these node pieces. We can remove
}

type Workload struct {
      // format: namespace/kind/name
      	fullName string
	pods []PodNetworking
}

type PodNetworking struct {
       IP string
      // don't worry about populating below fields right now
       IsHostNetworking bool
       NodeLabels []string
}
