// Copyright 2020 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/vmware/go-ipfix/pkg/collector"
	"github.com/vmware/go-ipfix/pkg/intermediate"
	"github.com/vmware/go-ipfix/pkg/registry"
)

var templatePacket = []byte{0, 10, 0, 104, 95, 208, 69, 75, 0, 0, 0, 0, 0, 0, 0, 1, 0, 2, 0, 88, 1, 0, 0, 14, 0, 8, 0, 4, 0, 12, 0, 4, 0, 7, 0, 2, 0, 11, 0, 2, 0, 4, 0, 1, 0, 151, 0, 4, 0, 86, 0, 8, 0, 2, 0, 8, 128, 101, 255, 255, 0, 0, 220, 186, 128, 103, 255, 255, 0, 0, 220, 186, 128, 106, 0, 4, 0, 0, 220, 186, 128, 108, 0, 2, 0, 0, 220, 186, 128, 86, 0, 8, 0, 0, 114, 121, 128, 2, 0, 8, 0, 0, 114, 121}
var dataPacket1 = []byte{0, 10, 0, 81, 95, 208, 77, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0, 65, 10, 0, 0, 1, 10, 0, 0, 2, 4, 210, 22, 46, 6, 74, 249, 240, 112, 0, 0, 0, 0, 0, 0, 3, 232, 0, 0, 0, 0, 0, 0, 1, 244, 0, 4, 112, 111, 100, 50, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 144, 0, 0, 0, 0, 0, 0, 0, 200}
var dataPacket2 = []byte{0, 10, 0, 81, 95, 208, 77, 154, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0, 65, 10, 0, 0, 1, 10, 0, 0, 2, 4, 210, 22, 46, 6, 74, 249, 244, 88, 0, 0, 0, 0, 0, 0, 1, 144, 0, 0, 0, 0, 0, 0, 0, 200, 4, 112, 111, 100, 49, 0, 10, 0, 0, 3, 18, 131, 0, 0, 0, 0, 0, 0, 3, 232, 0, 0, 0, 0, 0, 0, 1, 244}

/*
dataPacket1:
	"sourceIPv4Address": 10.0.0.1
	"destinationIPv4Address": 10.0.0.2
	"sourceTransportPort": 1234
	"destinationTransportPort": 5678
	"protocolIdentifier": 6
	"flowEndSeconds": 1257894000
	"packetTotalCount": 1000
	"packetDeltaCount": 500
	"sourcePodName": ""
	"destinationPodName": "pod2"
	"destinationClusterIPv4":
	"destinationServicePort":
	"reversePacketTotalCount": 400
	"reversePacketDeltaCount": 200
dataPacket 2:
	"sourceIPv4Address": 10.0.0.1
	"destinationIPv4Address": 10.0.0.2
	"sourceTransportPort": 1234
	"destinationTransportPort": 5678
	"protocolIdentifier": 6
	"flowEndSeconds": 1257895000
	"packetTotalCount": 400
	"packetDeltaCount": 200
	"sourcePodName": "pod1"
	"destinationPodName": ""
	"destinationClusterIPv4": 10.0.0.3
	"destinationServicePort": 4739
	"reversePacketTotalCount": 1000
	"reversePacketDeltaCount": 500
*/

var (
	flowKeyRecordMap = make(map[intermediate.FlowKey]intermediate.AggregationFlowRecord)
	flowKey          = intermediate.FlowKey{SourceAddress: "10.0.0.1", DestinationAddress: "10.0.0.2", Protocol: 6, SourcePort: 1234, DestinationPort: 5678}
	correlatefields  = []string{
		"sourcePodName",
		"sourcePodNamespace",
		"sourceNodeName",
		"destinationPodName",
		"destinationPodNamespace",
		"destinationNodeName",
		"destinationClusterIPv4",
		"destinationServicePort",
	}
	nonStatsElementList = []string{
		"flowEndSeconds",
	}
	statsElementList = []string{
		"packetTotalCount",
		"packetDeltaCount",
		"reversePacketTotalCount",
		"reversePacketDeltaCount",
	}
	antreaSourceStatsElementList = []string{
		"packetTotalCountFromSourceNode",
		"packetDeltaCountFromSourceNode",
		"reversePacketTotalCountFromSourceNode",
		"reversePacketDeltaCountFromSourceNode",
	}
	antreaDestinationStatsElementList = []string{
		"packetTotalCountFromDestinationNode",
		"packetDeltaCountFromDestinationNode",
		"reversePacketTotalCountFromDestinationNode",
		"reversePacketDeltaCountFromDestinationNode",
	}
)

func TestCollectorToIntermediate(t *testing.T) {
	registry.LoadRegistry()

	aggregatedFields := &intermediate.AggregationElements{
		NonStatsElements:                   nonStatsElementList,
		StatsElements:                      statsElementList,
		AggregatedSourceStatsElements:      antreaSourceStatsElementList,
		AggregatedDestinationStatsElements: antreaDestinationStatsElementList,
	}
	address, err := net.ResolveTCPAddr("tcp", "0.0.0.0:4739")
	if err != nil {
		t.Error(err)
	}
	// Initialize aggregation process and collecting process
	cpInput := collector.CollectorInput{
		Address:       address,
		MaxBufferSize: 1024,
		TemplateTTL:   0,
		IsEncrypted:   false,
		ServerCert:    nil,
		ServerKey:     nil,
	}
	cp, _ := collector.InitCollectingProcess(cpInput)

	apInput := intermediate.AggregationInput{
		MessageChan:       cp.GetMsgChan(),
		WorkerNum:         2,
		CorrelateFields:   correlatefields,
		AggregateElements: aggregatedFields,
	}
	ap, _ := intermediate.InitAggregationProcess(apInput)
	go cp.Start()
	waitForCollectorReady(t, cp)
	go func() {
		conn, err := net.DialTCP("tcp", nil, address)
		if err != nil {
			t.Errorf("TCP Collecting Process does not start correctly.")
		}
		defer conn.Close()
		conn.Write(templatePacket)
		conn.Write(dataPacket1)
		conn.Write(dataPacket2)
	}()
	go ap.Start()
	waitForAggregationToFinish(t, ap)
	cp.Stop()
	ap.Stop()

	assert.Equal(t, 1, len(flowKeyRecordMap), "Aggregation process should store the data record to map with corresponding flow key.")

	assert.NotNil(t, flowKeyRecordMap[flowKey])
	record := flowKeyRecordMap[flowKey].Record
	assert.Equal(t, 24, len(record.GetOrderedElementList()))
	for _, element := range record.GetOrderedElementList() {
		switch element.Element.Name {
		case "sourcePodName":
			assert.Equal(t, "pod1", element.Value)
		case "destinationPodName":
			assert.Equal(t, "pod2", element.Value)
		case "flowEndSeconds":
			assert.Equal(t, uint32(1257895000), element.Value)
		case "packetTotalCount":
			assert.Equal(t, uint64(400), element.Value)
		case "packetDeltaCount":
			assert.Equal(t, uint64(700), element.Value)
		case "destinationClusterIPv4":
			assert.Equal(t, net.IP{10, 0, 0, 3}, element.Value)
		case "destinationServicePort":
			assert.Equal(t, uint16(4739), element.Value)
		case "reversePacketDeltaCount":
			assert.Equal(t, uint64(700), element.Value)
		case "reversePacketTotalCount":
			assert.Equal(t, uint64(1000), element.Value)
		case "packetTotalCountFromSourceNode":
			assert.Equal(t, uint64(400), element.Value)
		case "packetDeltaCountFromSourceNode":
			assert.Equal(t, uint64(200), element.Value)
		case "packetTotalCountFromDestinationNode":
			assert.Equal(t, uint64(1000), element.Value)
		case "packetDeltaCountFromDestinationNode":
			assert.Equal(t, uint64(500), element.Value)
		}
	}

}

func copyFlowKeyRecordMap(key intermediate.FlowKey, aggregationFlowRecord intermediate.AggregationFlowRecord) error {
	flowKeyRecordMap[key] = aggregationFlowRecord
	return nil
}

func waitForCollectorReady(t *testing.T, cp *collector.CollectingProcess) {
	checkConn := func() (bool, error) {
		if strings.Split(cp.GetAddress().String(), ":")[1] == "0" {
			return false, fmt.Errorf("random port is not resolved")
		}
		if _, err := net.Dial(cp.GetAddress().Network(), cp.GetAddress().String()); err != nil {
			return false, err
		}
		return true, nil
	}
	if err := wait.Poll(100*time.Millisecond, 500*time.Millisecond, checkConn); err != nil {
		t.Errorf("Cannot establish connection to %s", cp.GetAddress().String())
	}
}

func waitForAggregationToFinish(t *testing.T, ap *intermediate.AggregationProcess) {
	checkConn := func() (bool, error) {
		ap.ForAllRecordsDo(copyFlowKeyRecordMap)
		if len(flowKeyRecordMap) > 0 {
			ie1, _ := flowKeyRecordMap[flowKey].Record.GetInfoElementWithValue("sourcePodName")
			ie2, _ := flowKeyRecordMap[flowKey].Record.GetInfoElementWithValue("destinationPodName")
			if ie1.Value == "pod1" && ie2.Value == "pod2" {
				return true, nil
			} else {
				return false, nil
			}
		} else {
			return false, fmt.Errorf("aggregation process does not process and store data correctly")
		}
	}
	if err := wait.Poll(100*time.Millisecond, 500*time.Millisecond, checkConn); err != nil {
		t.Error(err)
	}
}
