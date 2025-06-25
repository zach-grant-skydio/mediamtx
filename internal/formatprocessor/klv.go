package formatprocessor

import (
	"fmt"
	"time"

	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/pion/rtp"

	"github.com/bluenviron/mediamtx/internal/logger"
	"github.com/bluenviron/mediamtx/internal/unit"
)

type klv struct {
	UDPMaxPayloadSize  int
	Format            *format.Generic
	GenerateRTPPackets bool
	Parent            logger.Writer
}

func (t *klv) initialize() error {
	// KLV metadata typically doesn't need RTP packet generation
	// as it's usually carried in the RTP payload directly
	return nil
}

func (t *klv) ProcessUnit(u unit.Unit) error {
	// For KLV, we expect the unit to be already in RTP format
	// since it's typically carried in RTP packets
	return fmt.Errorf("KLV metadata must be processed as RTP packets")
}

func (t *klv) ProcessRTPPacket(
	pkt *rtp.Packet,
	ntp time.Time,
	pts int64,
	_ bool,
) (unit.Unit, error) {
	// Create a generic unit with the RTP packet
	u := &unit.Generic{
		Base: unit.Base{
			RTPPackets: []*rtp.Packet{pkt},
			NTP:        ntp,
			PTS:        pts,
		},
	}

	// Remove padding
	pkt.Padding = false
	pkt.PaddingSize = 0

	if pkt.MarshalSize() > t.UDPMaxPayloadSize {
		return nil, fmt.Errorf("payload size (%d) is greater than maximum allowed (%d)",
			pkt.MarshalSize(), t.UDPMaxPayloadSize)
	}

	return u, nil
}
