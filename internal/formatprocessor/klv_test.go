package formatprocessor

import (
	"testing"
	"time"

	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/pion/rtp"
	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediamtx/internal/unit"
)

func TestKLVProcessor(t *testing.T) {
	t.Run("rtp packet", func(t *testing.T) {
		// Create a KLV format
		fmt := &format.Generic{
			PayloadTyp: 96, // Common dynamic payload type for KLV
		}

		// Create processor
		proc, err := New(
			1500, // udpMaxPayloadSize
			fmt,
			true, // generateRTPPackets
			nil,  // logger
		)
		require.NoError(t, err)

		// Create a sample KLV RTP packet
		pkt := &rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				PayloadType:    96,
				SequenceNumber: 1234,
				Timestamp:      987654321,
				SSRC:           12345678,
			},
			Payload: []byte{
				0x06, 0x0E, 0x2B, 0x34, // KLV header (16-byte Universal Key)
				0x02, 0x0B, 0x01, 0x01,
				0x0E, 0x01, 0x03, 0x01,
				0x01, 0x00, 0x00, 0x00,
				0x0B, // Length (BER Short Form)
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, // Value
			},
		}

		// Process the RTP packet
		ntp := time.Now()
		u, err := proc.ProcessRTPPacket(pkt, ntp, 0, false)
		require.NoError(t, err)
		require.NotNil(t, u)

		// Verify the output unit
		require.IsType(t, &unit.Generic{}, u)
		gu := u.(*unit.Generic)
		require.Len(t, gu.RTPPackets, 1)
		require.Equal(t, pkt, gu.RTPPackets[0])
		require.Equal(t, ntp, gu.NTP)
	})

	t.Run("unit not supported", func(t *testing.T) {
		fmt := &format.Generic{
			PayloadTyp: 96,
		}

		proc, err := New(1500, fmt, true, nil)
		require.NoError(t, err)

		err = proc.ProcessUnit(nil)
		require.Error(t, err)
		require.Equal(t, "KLV metadata must be processed as RTP packets", err.Error())
	})
}
