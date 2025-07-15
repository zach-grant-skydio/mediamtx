package formatprocessor

import (
	"fmt"
	"time"

	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/format/rtpmpegts"
	"github.com/pion/rtp"

	"github.com/bluenviron/mediamtx/internal/logger"
	"github.com/bluenviron/mediamtx/internal/unit"
)

type mpegts struct {
	RTPMaxPayloadSize  int
	Format             *format.MPEGTS
	GenerateRTPPackets bool
	Parent             logger.Writer

	encoder *rtpmpegts.Encoder
	decoder *rtpmpegts.Decoder
}

func (t *mpegts) initialize() error {
	return nil
}

func (t *mpegts) ProcessUnit(uu unit.Unit) error {
	tunit := uu.(*unit.MPEGTS)

	if tunit.Data == nil {
		return nil
	}

	// generate RTP packets
	if t.GenerateRTPPackets {
		if t.encoder == nil {
			var err error
			t.encoder, err = t.Format.CreateEncoder()
			if err != nil {
				return err
			}
		}

		pkts, err := t.encoder.Encode(tunit.Data)
		if err != nil {
			return err
		}

		tunit.RTPPackets = pkts
	}

	return nil
}

func (t *mpegts) ProcessRTPPacket(
	pkt *rtp.Packet,
	ntp time.Time,
	pts int64,
	hasNonRTSPReaders bool,
) (unit.Unit, error) {
	u := &unit.MPEGTS{
		Base: unit.Base{
			RTPPackets: []*rtp.Packet{pkt},
			NTP:        ntp,
			PTS:        pts,
		},
	}

	// remove padding
	pkt.Padding = false
	pkt.PaddingSize = 0

	if len(pkt.Payload) > t.RTPMaxPayloadSize {
		return nil, fmt.Errorf("RTP payload size (%d) is greater than maximum allowed (%d)",
			len(pkt.Payload), t.RTPMaxPayloadSize)
	}

	// decode from RTP
	if hasNonRTSPReaders || t.decoder != nil {
		if t.decoder == nil {
			var err error
			t.decoder, err = t.Format.CreateDecoder()
			if err != nil {
				return nil, err
			}
		}

		data, err := t.decoder.Decode(pkt)
		if err != nil {
			if err == rtpmpegts.ErrMorePacketsNeeded {
				return u, nil
			}
			return nil, err
		}

		u.Data = data
	}

	// route packet as is
	return u, nil
}
