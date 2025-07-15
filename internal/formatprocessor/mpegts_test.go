package formatprocessor

import (
	"testing"
	"time"

	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/pion/rtp"
	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediamtx/internal/logger"
	"github.com/bluenviron/mediamtx/internal/unit"
)

func TestMPEGTSProcessUnit(t *testing.T) {
	forma := &format.MPEGTS{}

	p, err := New(1472, forma, true, Logger(func(logger.Level, string, ...interface{}) {}))
	require.NoError(t, err)

	// Test processing MPEG-TS unit
	testData := make([]byte, 7*188) // 7 MPEG-TS packets (STANAG 4609 typical)
	for i := 0; i < 7; i++ {
		testData[i*188] = 0x47 // MPEG-TS sync byte
	}

	unit := &unit.MPEGTS{
		Base: unit.Base{
			NTP: time.Now(),
			PTS: 90000,
		},
		Data: testData,
	}

	err = p.ProcessUnit(unit)
	require.NoError(t, err)
	require.NotEmpty(t, unit.RTPPackets)
	require.Equal(t, testData, unit.RTPPackets[0].Payload)
}

func TestMPEGTSProcessRTPPacket(t *testing.T) {
	forma := &format.MPEGTS{}

	p, err := New(1472, forma, true, Logger(func(logger.Level, string, ...interface{}) {}))
	require.NoError(t, err)

	// Test processing RTP packet with MPEG-TS data
	testData := make([]byte, 7*188) // 7 MPEG-TS packets
	for i := 0; i < 7; i++ {
		testData[i*188] = 0x47 // MPEG-TS sync byte
	}

	pkt := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    33,
			SequenceNumber: 1234,
			Timestamp:      45343,
			SSRC:           563423,
		},
		Payload: testData,
	}

	u, err := p.ProcessRTPPacket(pkt, time.Now(), 90000, true)
	require.NoError(t, err)
	require.NotNil(t, u)

	tunit := u.(*unit.MPEGTS)
	require.Equal(t, testData, tunit.Data)
	require.Equal(t, int64(90000), tunit.PTS)
}

func TestMPEGTSProcessRTPPacketPartial(t *testing.T) {
	forma := &format.MPEGTS{}

	p, err := New(1472, forma, true, Logger(func(logger.Level, string, ...interface{}) {}))
	require.NoError(t, err)

	// Test processing partial MPEG-TS packet (less than 188 bytes)
	partialData := make([]byte, 100)

	pkt := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    33,
			SequenceNumber: 1234,
			Timestamp:      45343,
			SSRC:           563423,
		},
		Payload: partialData,
	}

	u, err := p.ProcessRTPPacket(pkt, time.Now(), 90000, true)
	require.NoError(t, err)
	require.NotNil(t, u)

	// Should return unit even if no complete MPEG-TS packets yet
	tunit := u.(*unit.MPEGTS)
	require.Nil(t, tunit.Data) // No complete packets yet
}

func TestMPEGTSProcessEmptyPacket(t *testing.T) {
	forma := &format.MPEGTS{}

	p, err := New(1472, forma, true, Logger(func(logger.Level, string, ...interface{}) {}))
	require.NoError(t, err)

	pkt := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    33,
			SequenceNumber: 1234,
			Timestamp:      45343,
			SSRC:           563423,
		},
		Payload: []byte{},
	}

	u, err := p.ProcessRTPPacket(pkt, time.Now(), 90000, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty RTP payload")
	require.Nil(t, u)
}

func TestMPEGTSOversizedPacket(t *testing.T) {
	forma := &format.MPEGTS{}

	p, err := New(1000, forma, true, Logger(func(logger.Level, string, ...interface{}) {})) // Small max payload size
	require.NoError(t, err)

	// Create oversized payload
	oversizedData := make([]byte, 1500)

	pkt := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			PayloadType:    33,
			SequenceNumber: 1234,
			Timestamp:      45343,
			SSRC:           563423,
		},
		Payload: oversizedData,
	}

	u, err := p.ProcessRTPPacket(pkt, time.Now(), 90000, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "RTP payload size")
	require.Nil(t, u)
}
