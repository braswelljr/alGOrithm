package filesystems

import "fmt"

// RAIDResult is a RAID run: the layout (Disks[d][s] is disk d's block in
// stripe s), which block is what ("D3", "P", "Q", "copy of D3"), the data
// read back after the failures, and whether it matches.
type RAIDResult struct {
	Level     int        `json:"level"`
	Disks     [][]int    `json:"disks"`
	Roles     [][]string `json:"roles"`
	Failed    []int      `json:"failed"`
	Recovered []int      `json:"recovered"`
	Intact    bool       `json:"intact"`
	Usable    float64    `json:"usable"` // fraction of raw capacity that holds data
	Log       []string   `json:"log"`
}

// RAID stores data (one byte, 0-255, per block) on the given number of disks at RAID
// level 0, 1, 5 or 6, fails the disks in failed, and reads the data back.
//
//	RAID 0 stripes data across every disk: fast and full capacity, but one
//	       failure loses data.
//	RAID 1 mirrors every block on every disk: survives all but one failure,
//	       at 1/disks capacity.
//	RAID 5 gives each stripe one parity block P, the XOR of its data blocks,
//	       on a rotating disk: any one failed block is the XOR of the rest.
//	RAID 6 adds a second syndrome Q = sum of g^i * D_i in GF(2^8)
//	       (Reed-Solomon), so any two failures can be solved for.
//
// Time: O(len(data)) times O(disks). Space: O(len(data)).
func RAID(level, disks int, data []int, failed []int) (RAIDResult, error) {
	down, err := checkRAID(level, disks, data, failed)
	if err != nil {
		return RAIDResult{}, err
	}
	res := RAIDResult{Level: level, Disks: make([][]int, disks), Roles: make([][]string, disks), Failed: failed}
	per := map[int]int{0: disks, 1: 1, 5: disks - 1, 6: disks - 2}[level] // data blocks per stripe
	res.Usable = float64(per) / float64(disks)
	layout := res.stripe(data, per)
	tolerance := map[int]int{0: 0, 1: disks - 1, 5: 1, 6: 2}[level]
	res.Log = append(res.Log, fmt.Sprintf("RAID %d on %d disks: %d stripes, %.0f%% of the raw space holds data, survives %d failed disks", level, disks, len(layout), 100*res.Usable, tolerance))
	if len(failed) > tolerance {
		res.Log = append(res.Log, fmt.Sprintf("disks %v failed: more than RAID %d tolerates, the data is lost", failed, level))
		return res, nil
	}
	res.recover(data, layout, down)
	if len(failed) > 0 {
		res.Log = append(res.Log, fmt.Sprintf("disks %v failed and were rebuilt from the others: data intact %v", failed, res.Intact))
	}
	return res, nil
}

// checkRAID validates the array and returns the failed disks as a set.
func checkRAID(level, disks int, data, failed []int) (map[int]bool, error) {
	minDisks := map[int]int{0: 2, 1: 2, 5: 3, 6: 4}[level]
	if minDisks == 0 {
		return nil, inputErr("level 0, 1, 5 or 6")
	}
	if disks < minDisks || disks > 8 {
		return nil, inputErr("RAID %d needs %d-8 disks", level, minDisks)
	}
	if len(data) == 0 || len(data) > 48 {
		return nil, inputErr("1-48 data bytes")
	}
	down := map[int]bool{}
	for _, d := range failed {
		if d < 0 || d >= disks || down[d] {
			return nil, inputErr("failed disks are distinct, 0-%d", disks-1)
		}
		down[d] = true
	}
	for _, v := range data {
		if v < 0 || v > 255 {
			return nil, inputErr("data bytes are 0-255")
		}
	}
	return down, nil
}

func (res *RAIDResult) put(d, v int, role string) {
	res.Disks[d] = append(res.Disks[d], v)
	res.Roles[d] = append(res.Roles[d], role)
}

// stripe writes data across the disks, per data blocks a stripe, and returns
// the layout: layout[s][d] is the data index disk d holds in stripe s, or -1
// for P and -2 for Q.
func (res *RAIDResult) stripe(data []int, per int) [][]int {
	disks := len(res.Disks)
	layout := make([][]int, (len(data)+per-1)/per)
	for s := range layout {
		block := func(i int) int {
			if k := s*per + i; k < len(data) {
				return data[k]
			}
			return 0 // padding
		}
		switch res.Level {
		case 0:
			layout[s] = make([]int, disks)
			for d := 0; d < disks; d++ {
				layout[s][d] = s*per + d
				res.put(d, block(d), fmt.Sprintf("D%d", s*per+d))
			}
		case 1:
			layout[s] = make([]int, disks)
			for d := 0; d < disks; d++ {
				layout[s][d] = s
				role := fmt.Sprintf("D%d", s)
				if d > 0 {
					role = "copy of " + role
				}
				res.put(d, block(0), role)
			}
		default:
			layout[s] = res.parityStripe(s, per, block)
		}
	}
	return layout
}

// parityStripe writes stripe s of RAID 5 or 6: the data blocks, P (their XOR)
// on a disk that rotates with the stripe, and for RAID 6 Q (a Reed-Solomon
// sum in GF(256)) on the next disk.
func (res *RAIDResult) parityStripe(s, per int, block func(int) int) []int {
	disks := len(res.Disks)
	row := make([]int, disks)
	p := (disks - 1 - s%disks + disks) % disks // the parity disk rotates
	q := (p + 1) % disks
	i, P, Q := 0, 0, 0
	for d := 0; d < disks; d++ {
		switch {
		case d == p:
			row[d] = -1
		case res.Level == 6 && d == q:
			row[d] = -2
		default:
			row[d] = s*per + i
			P ^= block(i)
			Q ^= gfMul(gfPow(i), block(i))
			i++
		}
	}
	for d := 0; d < disks; d++ {
		switch row[d] {
		case -1:
			res.put(d, P, "P")
		case -2:
			res.put(d, Q, "Q")
		default:
			res.put(d, block(row[d]-s*per), fmt.Sprintf("D%d", row[d]))
		}
	}
	return row
}

// recover rebuilds each stripe from the disks that are up and checks the
// data came back whole.
func (res *RAIDResult) recover(data []int, layout [][]int, down map[int]bool) {
	disks := len(res.Disks)
	res.Recovered = make([]int, len(data))
	for s := range layout {
		vals := make([]int, disks)
		for d := 0; d < disks; d++ {
			vals[d] = res.Disks[d][s]
		}
		rebuildStripe(res.Level, layout[s], vals, down)
		for d := 0; d < disks; d++ {
			if k := layout[s][d]; k >= 0 && k < len(data) && (res.Level != 1 || d == 0) {
				res.Recovered[k] = vals[d]
			}
		}
	}
	res.Intact = true
	for i := range data {
		if res.Recovered[i] != data[i] {
			res.Intact = false
		}
	}
}

// rebuildStripe fills in vals for the down disks of one stripe.
func rebuildStripe(level int, layout, vals []int, down map[int]bool) {
	disks := len(vals)
	switch level {
	case 1:
		for d := range vals {
			if !down[d] {
				for x := range vals {
					vals[x] = vals[d]
				}
				return
			}
		}
	case 5, 6:
		idx := map[int]int{} // data disk -> its index i within the stripe
		var pDisk, qDisk = -1, -1
		i := 0
		for d := 0; d < disks; d++ {
			switch layout[d] {
			case -1:
				pDisk = d
			case -2:
				qDisk = d
			default:
				idx[d] = i
				i++
			}
		}
		var lostData []int
		for d := range down {
			if layout[d] >= 0 {
				lostData = append(lostData, d)
			}
		}
		syndromes := func() (P, Q int) {
			for d, k := range idx {
				if !down[d] {
					P ^= vals[d]
					Q ^= gfMul(gfPow(k), vals[d])
				}
			}
			return P, Q
		}
		P, Q := syndromes()
		switch {
		case len(lostData) == 1 && !down[pDisk]: // XOR with P
			vals[lostData[0]] = vals[pDisk] ^ P
		case len(lostData) == 1: // P is down too: use Q
			x := lostData[0]
			vals[x] = gfDiv(vals[qDisk]^Q, gfPow(idx[x]))
		case len(lostData) == 2:
			x, y := lostData[0], lostData[1]
			pp, qq := vals[pDisk]^P, vals[qDisk]^Q // Dx^Dy and g^x Dx ^ g^y Dy
			gx, gy := gfPow(idx[x]), gfPow(idx[y])
			vals[x] = gfDiv(qq^gfMul(gy, pp), gx^gy)
			vals[y] = pp ^ vals[x]
		}
		P, Q = 0, 0
		for d, k := range idx {
			P ^= vals[d]
			Q ^= gfMul(gfPow(k), vals[d])
		}
		if pDisk >= 0 {
			vals[pDisk] = P
		}
		if qDisk >= 0 {
			vals[qDisk] = Q
		}
	}
}

// GF(2^8) with the polynomial x^8+x^4+x^3+x^2+1 (0x11d) and generator 2,
// as RAID 6 uses.
var gfExp, gfLog = func() ([512]int, [256]int) {
	var e [512]int
	var l [256]int
	x := 1
	for i := 0; i < 255; i++ {
		e[i] = x
		l[x] = i
		x <<= 1
		if x&0x100 != 0 {
			x ^= 0x11d
		}
	}
	for i := 255; i < 512; i++ {
		e[i] = e[i-255]
	}
	return e, l
}()

func gfMul(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return gfExp[gfLog[a]+gfLog[b]]
}

func gfDiv(a, b int) int {
	if a == 0 {
		return 0
	}
	return gfExp[(gfLog[a]-gfLog[b]+255)%255]
}

func gfPow(i int) int { return gfExp[i%255] }
