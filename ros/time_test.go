package ros

import "testing"

func TestFromSec(t *testing.T) {
	time := FromSec(3 + 7e-9)
	if time.Sec != 3 {
		t.Errorf("FromSec() said %v seconds, should be 3.", time.Sec)
	} else {
		t.Log("FromSec() seconds correct.")
	}
	if time.NSec != 7 {
		t.Errorf("FromSec() said %v nanoseconds, should be 7.", time.NSec)
	} else {
		t.Log("FromSec() nanoseconds correct.")
	}
}

func TestToSec(t *testing.T) {
	var time Time
	time.Sec = 2
	time.NSec = 5
	seconds := time.ToSec()
	if seconds != (2 + 5e-9) {
		t.Errorf("ToSec() gave %v seconds, should be 2.000000005.", seconds)
	} else {
		t.Log("ToSec() seconds correct.")
	}
}