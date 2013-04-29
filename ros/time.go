package ros

import "math"

type Time struct {
	Sec uint32
	NSec uint32
}

type Duration struct {
	Sec int32
	NSec int32
}

func FromSec(seconds float64) Time {
	var t Time
	floor_secs := math.Floor(seconds)
	t.Sec = uint32(floor_secs)
	t.NSec = uint32(1e9*(seconds - floor_secs))
	return t
}

func (t Time) ToSec() float64 {
	seconds := float64(t.Sec)
	seconds += float64(t.NSec) / 1e9
	return seconds
}
