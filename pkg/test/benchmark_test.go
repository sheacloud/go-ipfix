//Can be run with go test -bench=. -benchmem -cpuprofile cpu.out -memprofile mem.out while in this directory

package main

import (
	"net"
	"testing"
	"time"

	"github.com/vmware/go-ipfix/pkg/entities"
	"github.com/vmware/go-ipfix/pkg/registry"
)

var flow flow7Tuple
var dataSet entities.Set

var sourceIPv4AddressElement *entities.InfoElement
var destinationIPv4AddressElement *entities.InfoElement
var sourceTransportPortElement *entities.InfoElement
var destinationTransportPortElement *entities.InfoElement
var protocolIdentifierElement *entities.InfoElement
var flowStartMillisecondsElement *entities.InfoElement
var flowEndMillisecondsElement *entities.InfoElement
var octetTotalCountElement *entities.InfoElement
var packetTotalCountElement *entities.InfoElement

func init() {
	registry.LoadRegistry()

	flow = flow7Tuple{
		SrcAddr:               net.ParseIP("192.168.0.1"),
		DstAddr:               net.ParseIP("192.168.1.1"),
		SrcPort:               50000,
		DstPort:               443,
		Protocol:              6,
		FlowStartMilliseconds: uint64(time.Now().UnixNano() / 1000000),
		FlowEndMilliseconds:   uint64(time.Now().UnixNano()/1000000) + 100,
		OctetCount:            4000,
		PacketCount:           5,
	}

	dataSet = entities.NewSet(entities.Data, 1, false)

	sourceIPv4AddressElement, _ = registry.GetInfoElement("sourceIPv4Address", registry.IANAEnterpriseID)
	destinationIPv4AddressElement, _ = registry.GetInfoElement("destinationIPv4Address", registry.IANAEnterpriseID)
	sourceTransportPortElement, _ = registry.GetInfoElement("sourceTransportPort", registry.IANAEnterpriseID)
	destinationTransportPortElement, _ = registry.GetInfoElement("destinationTransportPort", registry.IANAEnterpriseID)
	protocolIdentifierElement, _ = registry.GetInfoElement("protocolIdentifier", registry.IANAEnterpriseID)
	flowStartMillisecondsElement, _ = registry.GetInfoElement("flowStartMilliseconds", registry.IANAEnterpriseID)
	flowEndMillisecondsElement, _ = registry.GetInfoElement("flowEndMilliseconds", registry.IANAEnterpriseID)
	octetTotalCountElement, _ = registry.GetInfoElement("octetTotalCount", registry.IANAEnterpriseID)
	packetTotalCountElement, _ = registry.GetInfoElement("packetTotalCount", registry.IANAEnterpriseID)
}

type flow7Tuple struct {
	SrcAddr, DstAddr                           net.IP
	SrcPort, DstPort                           uint16
	Protocol                                   uint8
	FlowStartMilliseconds, FlowEndMilliseconds uint64
	OctetCount, PacketCount                    uint64
}

func generateFlowMessage(flow flow7Tuple) {
	elements := make([]*entities.InfoElementWithValue, 9)

	ie := entities.NewInfoElementWithValue(sourceIPv4AddressElement, flow.SrcAddr)
	elements[0] = ie

	ie = entities.NewInfoElementWithValue(destinationIPv4AddressElement, flow.DstAddr)
	elements[1] = ie

	ie = entities.NewInfoElementWithValue(sourceTransportPortElement, flow.SrcPort)
	elements[2] = ie

	ie = entities.NewInfoElementWithValue(destinationTransportPortElement, flow.DstPort)
	elements[3] = ie

	ie = entities.NewInfoElementWithValue(protocolIdentifierElement, flow.Protocol)
	elements[4] = ie

	ie = entities.NewInfoElementWithValue(flowStartMillisecondsElement, flow.FlowStartMilliseconds)
	elements[5] = ie

	ie = entities.NewInfoElementWithValue(flowEndMillisecondsElement, flow.FlowEndMilliseconds)
	elements[6] = ie

	ie = entities.NewInfoElementWithValue(octetTotalCountElement, flow.OctetCount)
	elements[7] = ie

	ie = entities.NewInfoElementWithValue(packetTotalCountElement, flow.PacketCount)
	elements[8] = ie

	dataSet.AddRecord(elements, 1)
}

func BenchmarkFlowRecordCreation(b *testing.B) {
	for n := 0; n < b.N; n++ {
		generateFlowMessage(flow)
	}
}
