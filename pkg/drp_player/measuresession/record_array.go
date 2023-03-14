package measuresession

import (
	"encoding/binary"
	"fmt"

	"github.com/telekom/aml-jens/internal/persistence/datatypes"
)

const (
	RECORD_SIZE int = 64

	TC_JENS_RELAY_ECN_VALID    byte = 1 << 2
	TC_JENS_RELAY_SOJOURN_SLOW byte = 1 << 5
	TC_JENS_RELAY_SOJOURN_MARK byte = 1 << 6
	TC_JENS_RELAY_SOJOURN_DROP byte = 1 << 7
)
const (
	RECORD_TYPE_Q RecoordArrayType = 6
	RECORD_TYPE_P RecoordArrayType = 7
)

type RecoordArrayType byte

// Represents a Record from qdisc
type RecordArray []byte

// Extracts encoded type_id
func (r RecordArray) type_id() RecoordArrayType {
	return RecoordArrayType(r[8])
}

// Extracts encoded timestampMs
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
	return &DB_measure_queue{
		Time:              record.timestamp() + diffTimeMs,
		Memoryusagebytes:  uint32(binary.LittleEndian.Uint32(record[12:16])),
		PacketsInQueue:    uint16(binary.LittleEndian.Uint16(record[10:12])),
		Fk_session_tag_id: session_id,
	}, nil
}

type PacketMeasure struct {
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
	const IP_TEMPLATE = "%d.%d.%d.%d"
	if record.type_id() != RECORD_TYPE_P {
		return nil, fmt.Errorf("Cant Parse recordarray %v as PacketMeasure: invalid type", record)
	}
	srcIp := fmt.Sprintf(IP_TEMPLATE, uint8(record[28]), uint8(record[29]), uint8(record[30]), uint8(record[31]))
	dstIp := fmt.Sprintf(IP_TEMPLATE, uint8(record[44]), uint8(record[45]), uint8(record[46]), uint8(record[47]))
	nextHdr := record[53]
	var srcPort uint16 = 0
	var dstPort uint16 = 0
	if nextHdr == 6 || nextHdr == 17 {
		srcPort = uint16(binary.LittleEndian.Uint16(record[54:56]))
		dstPort = uint16(binary.LittleEndian.Uint16(record[56:58]))
	}
	if srcIp == "0.0.0.0" && dstIp == "0.0.0.0" {
		//Non-ip packet - ignore!
		return nil, nil
	}
	var prio uint8 = record[51] & 0b11000000

	flow := datatypes.DB_network_flow{
		Source_ip:        srcIp,
		Source_port:      srcPort,
		Destination_ip:   dstIp,
		Destination_port: dstPort,
		Session_id:       session_id,
		Prio:             prio,
	}
	packetMeasure := PacketMeasure{
		timestampMs:    record.timestamp(),
		sojournTimeMs:  uint32(binary.LittleEndian.Uint32(record[12:16])) / 1e3,
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
