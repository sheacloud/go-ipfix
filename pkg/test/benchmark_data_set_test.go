//Can be run with go test -bench=. -benchmem -cpuprofile cpu.out -memprofile mem.out while in this directory

package test

import (
	"net"
	"testing"
	"time"

	"github.com/vmware/go-ipfix/pkg/entities"
	"github.com/vmware/go-ipfix/pkg/registry"
)

var flow flowTuple
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

var tcpBitsElement *entities.InfoElement
var ingressInterfaceElement *entities.InfoElement
var egressInterfaceElement *entities.InfoElement
var nextHopIPv4AddressElement *entities.InfoElement
var vlanIdElement *entities.InfoElement
var sourceMacAddrElement *entities.InfoElement
var dstMacAddrElement *entities.InfoElement

func init() {
	registry.LoadRegistry()

	srcMac, _ := net.ParseMAC("00:11:22:33:44:55")
	dstMac, _ := net.ParseMAC("66:77:88:99:aa:bb")

	flow = flowTuple{
		srcAddr:               net.ParseIP("192.168.0.1"),
		dstAddr:               net.ParseIP("192.168.1.1"),
		srcPort:               50000,
		dstPort:               443,
		protocol:              6,
		tcpBits:               0,
		ingressInterface:      1,
		egressInterface:       2,
		nextHop:               net.ParseIP("192.168.1.2"),
		vlanId:                5,
		sourceMacAddr:         srcMac,
		dstMacAddr:            dstMac,
		flowStartMilliseconds: uint64(time.Now().UnixNano() / 1000000),
		flowEndMilliseconds:   uint64(time.Now().UnixNano()/1000000) + 100,
		octetCount:            4000,
		packetCount:           5,
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

	tcpBitsElement, _ = registry.GetInfoElement("tcpControlBits", registry.IANAEnterpriseID)
	ingressInterfaceElement, _ = registry.GetInfoElement("ingressInterface", registry.IANAEnterpriseID)
	egressInterfaceElement, _ = registry.GetInfoElement("egressInterface", registry.IANAEnterpriseID)
	nextHopIPv4AddressElement, _ = registry.GetInfoElement("ipNextHopIPv4Address", registry.IANAEnterpriseID)
	vlanIdElement, _ = registry.GetInfoElement("vlanId", registry.IANAEnterpriseID)
	sourceMacAddrElement, _ = registry.GetInfoElement("sourceMacAddress", registry.IANAEnterpriseID)
	dstMacAddrElement, _ = registry.GetInfoElement("destinationMacAddress", registry.IANAEnterpriseID)
}

type flowTuple struct {
	srcAddr, dstAddr                           net.IP
	srcPort, dstPort                           uint16
	protocol                                   uint8
	tcpBits                                    uint16
	ingressInterface, egressInterface          uint32
	nextHop                                    net.IP
	vlanId                                     uint16
	sourceMacAddr, dstMacAddr                  net.HardwareAddr
	flowStartMilliseconds, flowEndMilliseconds uint64
	octetCount, packetCount                    uint64
}

func populateBasicFlowRecord(flow flowTuple, buffer []*entities.InfoElementWithValue) {
	buffer[0] = entities.NewInfoElementWithValue(sourceIPv4AddressElement, flow.srcAddr)
	buffer[1] = entities.NewInfoElementWithValue(destinationIPv4AddressElement, flow.dstAddr)
	buffer[2] = entities.NewInfoElementWithValue(sourceTransportPortElement, flow.srcPort)
	buffer[3] = entities.NewInfoElementWithValue(destinationTransportPortElement, flow.dstPort)
	buffer[4] = entities.NewInfoElementWithValue(protocolIdentifierElement, flow.protocol)

	buffer[5] = entities.NewInfoElementWithValue(flowStartMillisecondsElement, flow.flowStartMilliseconds)
	buffer[6] = entities.NewInfoElementWithValue(flowEndMillisecondsElement, flow.flowEndMilliseconds)
	buffer[7] = entities.NewInfoElementWithValue(octetTotalCountElement, flow.octetCount)
	buffer[8] = entities.NewInfoElementWithValue(packetTotalCountElement, flow.packetCount)
}

func populateExtendedFlowRecord(flow flowTuple, buffer []*entities.InfoElementWithValue) {
	buffer[0] = entities.NewInfoElementWithValue(sourceIPv4AddressElement, flow.srcAddr)
	buffer[1] = entities.NewInfoElementWithValue(destinationIPv4AddressElement, flow.dstAddr)
	buffer[2] = entities.NewInfoElementWithValue(sourceTransportPortElement, flow.srcPort)
	buffer[3] = entities.NewInfoElementWithValue(destinationTransportPortElement, flow.dstPort)
	buffer[4] = entities.NewInfoElementWithValue(protocolIdentifierElement, flow.protocol)

	buffer[5] = entities.NewInfoElementWithValue(tcpBitsElement, flow.tcpBits)
	buffer[6] = entities.NewInfoElementWithValue(ingressInterfaceElement, flow.ingressInterface)
	buffer[7] = entities.NewInfoElementWithValue(egressInterfaceElement, flow.egressInterface)
	buffer[8] = entities.NewInfoElementWithValue(nextHopIPv4AddressElement, flow.nextHop)
	buffer[9] = entities.NewInfoElementWithValue(vlanIdElement, flow.vlanId)
	buffer[10] = entities.NewInfoElementWithValue(sourceMacAddrElement, flow.sourceMacAddr)
	buffer[11] = entities.NewInfoElementWithValue(dstMacAddrElement, flow.dstMacAddr)

	buffer[12] = entities.NewInfoElementWithValue(flowStartMillisecondsElement, flow.flowStartMilliseconds)
	buffer[13] = entities.NewInfoElementWithValue(flowEndMillisecondsElement, flow.flowEndMilliseconds)
	buffer[14] = entities.NewInfoElementWithValue(octetTotalCountElement, flow.octetCount)
	buffer[15] = entities.NewInfoElementWithValue(packetTotalCountElement, flow.packetCount)
}

// generate a basic flow record, instantiating each IE from scratch. Benchmarks instantiating info elements as well as adding a record to a data set
func generateFlowMessageBasic(flow flowTuple, buffer []*entities.InfoElementWithValue) {
	populateBasicFlowRecord(flow, buffer)

	dataSet.AddRecord(buffer, 1)
}

func generateFlowMessageExtended(flow flowTuple, buffer []*entities.InfoElementWithValue) {
	populateExtendedFlowRecord(flow, buffer)

	dataSet.AddRecord(buffer, 1)
}

func generateFlowMessageSameRecord(flowRecord []*entities.InfoElementWithValue) {
	dataSet.AddRecord(flowRecord, 1)
}

// Tests creating a record with new info elements and adding it to an existing data set
func BenchmarkSetAddRecordNewRecordBasic(b *testing.B) {
	buffer := make([]*entities.InfoElementWithValue, 9)
	for n := 0; n < b.N; n++ {
		generateFlowMessageBasic(flow, buffer)
	}
}

// Tests creating a record with new info elements and adding it to an existing data set
func BenchmarkSetAddRecordNewRecordExtended(b *testing.B) {
	buffer := make([]*entities.InfoElementWithValue, 16)
	for n := 0; n < b.N; n++ {
		generateFlowMessageExtended(flow, buffer)
	}
}

// Tests adding a pre-populated record to an existing data set
func BenchmarkSetAddRecordSameRecordBasic(b *testing.B) {
	buffer := make([]*entities.InfoElementWithValue, 9)
	populateBasicFlowRecord(flow, buffer)

	for n := 0; n < b.N; n++ {
		generateFlowMessageSameRecord(buffer)
	}
}

// Tests adding a pre-populated record to an existing data set
func BenchmarkSetAddRecordSameRecordExtended(b *testing.B) {
	buffer := make([]*entities.InfoElementWithValue, 16)
	populateExtendedFlowRecord(flow, buffer)

	for n := 0; n < b.N; n++ {
		generateFlowMessageSameRecord(buffer)
	}
}
