package server

import "testing"

func TestIPRange_Basic(t *testing.T) {
	r, err := newIPRange("192.168.0.0/24,127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cases := []struct {
		ip   string
		want bool
	}{
		{"192.168.0.1", true},
		{"192.168.0.255", true},
		{"192.168.1.1", false},
		{"127.0.0.1", true},
		{"10.0.0.1", false},
	}
	for _, c := range cases {
		if got := r.includeStr(c.ip); got != c.want {
			t.Errorf("includeStr(%q) = %v, want %v", c.ip, got, c.want)
		}
	}
}

func TestIPRange_IPv6(t *testing.T) {
	r, err := newIPRange("::1,2001:db8::/32")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cases := []struct {
		ip   string
		want bool
	}{
		{"::1", true},
		{"2001:db8::1", true},
		{"2001:db9::1", false},
		{"192.168.1.1", false},
	}
	for _, c := range cases {
		if got := r.includeStr(c.ip); got != c.want {
			t.Errorf("includeStr(%q) = %v, want %v", c.ip, got, c.want)
		}
	}
}

func TestIPRange_DefaultAllowAll(t *testing.T) {
	r, err := newIPRange("0.0.0.0/0,::/0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.includeStr("1.2.3.4") {
		t.Error("0.0.0.0/0 should include 1.2.3.4")
	}
	if !r.includeStr("fe80::1") {
		t.Error("::/0 should include fe80::1")
	}
}

func TestIPRange_Invalid(t *testing.T) {
	_, err := newIPRange("not-an-ip")
	if err == nil {
		t.Error("expected error for invalid input")
	}
	_, err = newIPRange("")
	if err == nil {
		t.Error("expected error for empty input")
	}
	_, err = newIPRange("   ,   ")
	if err == nil {
		t.Error("expected error for only-whitespace input")
	}
}

func TestIPRange_SpacesAroundEntries(t *testing.T) {
	r, err := newIPRange("  192.168.1.0/24 ,  10.0.0.1  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.includeStr("192.168.1.5") {
		t.Error("expected 192.168.1.5 to be included")
	}
	if !r.includeStr("10.0.0.1") {
		t.Error("expected 10.0.0.1 to be included")
	}
}
