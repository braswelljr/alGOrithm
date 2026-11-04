package networking

import "testing"

// TestPrefixContains checks membership in a CIDR prefix at both edges and just
// outside them.
func TestPrefixContains(t *testing.T) {
	p, err := ParsePrefix("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	for addr, want := range map[string]bool{
		"192.168.1.0":   true,
		"192.168.1.255": true,
		"192.168.2.0":   false,
		"192.168.0.255": false,
	} {
		a, err := ParseAddr(addr)
		if err != nil {
			t.Fatal(err)
		}
		if p.Contains(a) != want {
			t.Errorf("%v contains %s: got %v, want %v", p, addr, !want, want)
		}
	}
}
