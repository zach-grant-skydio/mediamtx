package unit

// MPEGTS is a MPEG-TS data unit.
type MPEGTS struct {
	Base
	Data []byte
}
