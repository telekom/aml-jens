package measuresession

import (
	"encoding/binary"
	"fmt"

	"github.com/telekom/aml-jens/internal/persistence/datatypes"
)

const (
	RECORD_SIZE   = 64
	RECORD_TYPE_Q = 6
	RECORD_TYPE_P = 7

	TC_JENS_RELAY_ECN_VALID    = 1 << 2
	TC_JENS_RELAY_SOJOURN_SLOW = 1 << 5
	TC_JENS_RELAY_SOJOURN_MARK = 1 << 6
	TC_JENS_RELAY_SOJOURN_DROP = 1 << 7
)

type RecordArray []byte

func (r RecordArray) type_id() byte {
	return r[8]
}
func (r RecordArray) timestamp() uint64 {
	return uint64(binary.LittleEndian.Uint64(r[0:8])) / 1e6

}

// Creates a new DB_measure_queue from the supplied Record.
//
// Possible Returns:
//   - nil, err               -> in event of actual error
//   - &DB_measure_queue, nil -> everything OK
func (record RecordArray) AsDB_measure_queue(diffTimeMs uint64, session_id int) (*DB_measure_queue, error) {
	if record.type_id() != RECORD_TYPE_Q {
		return nil, fmt.Errorf("Cant Parse recordarray %v as DB_measure_queue: invalid type", record)
	}
	numberOfPacketsInQueue := uint16(binary.LittleEndian.Uint16(record[10:12]))
	memUsageBytes := uint32(binary.LittleEndian.Uint32(record[12:16]))
	return &DB_measure_queue{
		Time:              record.timestamp() + diffTimeMs,
		Memoryusagebytes:  memUsageBytes,
		PacketsInQueue:    numberOfPacketsInQueue,
		Fk_session_tag_id: session_id,
	}, nil
}

type PacketMeasure struct {
	//internal use, as unique ID
	timestampMs    uint64
	sojournTimeMs  uint32
	ecnIn          uint8
	ecnOut         uint8
	ecnValid       bool
	slow           bool
	mark           bool
	drop           bool
	ipVersion      uint8
	packetSizeByte uint32
	net_flow       *datatypes.DB_network_flow
}

// Creates a new PacketMeasure from the supplied Record.
//
// Possible Returns:
//   - nil, err            -> in event of actual error
//   - &PacketMeasure, nil -> everything OK
//   - nil, nil            -> Skipped (due to ip = 0.0.0.0)
func (record RecordArray) AsPacketMeasure(session_id int) (*PacketMeasure, error) {
	if record.type_id() != RECORD_TYPE_P {
		return nil, fmt.Errorf("Cant Parse recordarray %v as PacketMeasure: invalid type", record)
	}
	timestampMs := uint64(binary.LittleEndian.Uint64(record[0:8])) / 1e6
	sojournTimeMs := uint32(binary.LittleEndian.Uint32(record[12:16])) / 1e3

	srcIp := fmt.Sprintf("%d.%d.%d.%d", uint8(record[28]), uint8(record[29]), uint8(record[30]), uint8(record[31]))
	dstIp := fmt.Sprintf("%d.%d.%d.%d", uint8(record[44]), uint8(record[45]), uint8(record[46]), uint8(record[47]))
	nextHdr := record[53]
	var srcPort uint16 = 0
	var dstPort uint16 = 0
	if nextHdr == 6 || nextHdr == 17 {
		srcPort = uint16(binary.LittleEndian.Uint16(record[54:56]))
		dstPort = uint16(binary.LittleEndian.Uint16(record[56:58]))
	}
	if srcIp == "0.0.0.0" && dstIp == "0.0.0.0" {
		return nil, nil
	}
	flow := datatypes.DB_network_flow{
		Source_ip:        srcIp,
		Source_port:      srcPort,
		Destination_ip:   dstIp,
		Destination_port: dstPort,
		Session_id:       session_id,
	}
	packetMeasure := PacketMeasure{
		timestampMs:    timestampMs,
		sojournTimeMs:  sojournTimeMs,
		ecnIn:          record[9] & 3,
		ecnOut:         (record[9] & 24) >> 3,
		ecnValid:       (record[9] & TC_JENS_RELAY_ECN_VALID) != 0,
		slow:           (record[9] & TC_JENS_RELAY_SOJOURN_SLOW) != 0,
		mark:           (record[9] & TC_JENS_RELAY_SOJOURN_MARK) != 0,
		drop:           (record[9] & TC_JENS_RELAY_SOJOURN_DROP) != 0,
		ipVersion:      record[52],
		packetSizeByte: uint32(binary.LittleEndian.Uint32(record[48:52])),
		net_flow:       &flow,
	}
	return &packetMeasure, nil
}
