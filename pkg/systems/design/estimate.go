package design

import "fmt"

// EstimateConfig is the traffic of a service to estimate, per day.
type EstimateConfig struct {
	DailyUsers    int // daily active users
	ReadsPerUser  int // reads each user makes a day
	WritesPerUser int // writes each user makes a day
	ObjectBytes   int // bytes stored per write
	PeakFactor    int // peak traffic as a multiple of the average
	Years         int // how long data is kept
	Replicas      int // copies of every stored byte
}

const secondsPerDay = 86_400

// Estimate is a back-of-the-envelope estimate: the average and peak reads and
// writes a second, the storage a day and over the retention period with its
// replicas, and the bandwidth in and out. It is the first step of a design,
// sizing the problem before drawing boxes. It assumes traffic spread evenly
// over the day apart from PeakFactor, and every read returning one object;
// it leaves out metadata, indexes and compression.
//
// Time: O(1). Space: O(1).
func Estimate(c EstimateConfig) (Report, error) {
	if err := check(c.DailyUsers > 0 && c.ReadsPerUser >= 0 && c.WritesPerUser >= 0 && c.ObjectBytes > 0 &&
		c.PeakFactor >= 1 && c.Years >= 1 && c.Replicas >= 1, "every count must be positive (reads and writes may be zero)"); err != nil {
		return Report{}, err
	}
	var r Report
	reads := float64(c.DailyUsers) * float64(c.ReadsPerUser) / secondsPerDay
	writes := float64(c.DailyUsers) * float64(c.WritesPerUser) / secondsPerDay
	perDay := float64(c.DailyUsers) * float64(c.WritesPerUser) * float64(c.ObjectBytes)
	total := perDay * 365 * float64(c.Years) * float64(c.Replicas)
	r.say("reads: %d users x %d a day / 86,400 s = %.1f a second on average, %.1f at peak", c.DailyUsers, c.ReadsPerUser, reads, reads*float64(c.PeakFactor))
	r.say("writes: %d users x %d a day / 86,400 s = %.1f a second on average, %.1f at peak", c.DailyUsers, c.WritesPerUser, writes, writes*float64(c.PeakFactor))
	r.say("read:write ratio %s, so %s", ratio(reads, writes), readHeavy(reads, writes))
	r.say("storage: %s a day, %s over %d years with %d replicas", bytes(perDay), bytes(total), c.Years, c.Replicas)
	r.say("bandwidth: %s/s in, %s/s out on average", bytes(writes*float64(c.ObjectBytes)), bytes(reads*float64(c.ObjectBytes)))
	r.put("reads per second", round1(reads))
	r.put("peak reads per second", round1(reads*float64(c.PeakFactor)))
	r.put("writes per second", round1(writes))
	r.put("peak writes per second", round1(writes*float64(c.PeakFactor)))
	r.put("storage per day (GB)", round1(perDay/1e9))
	r.put("total storage (TB)", round1(total/1e12))
	return r, nil
}

func ratio(reads, writes float64) string {
	if writes == 0 {
		return "all reads"
	}
	return fmt.Sprintf("%.0f:1", reads/writes)
}

func readHeavy(reads, writes float64) string {
	switch {
	case writes == 0 || reads >= 10*writes:
		return "read heavy: cache and replicate reads"
	case reads*10 <= writes:
		return "write heavy: partition writes and batch them"
	default:
		return "balanced: size reads and writes alike"
	}
}

// bytes writes a byte count with a decimal unit.
func bytes(n float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	i := 0
	for n >= 1000 && i < len(units)-1 {
		n /= 1000
		i++
	}
	return fmt.Sprintf("%.1f %s", n, units[i])
}

func round1(x float64) float64 { return float64(int64(x*10+0.5)) / 10 }

// LatencyNumbers is the table of latencies every programmer should know
// (after Jeff Dean and Peter Norvig), in nanoseconds, with what each is
// relative to a main memory reference. The numbers are orders of magnitude
// from about 2012 hardware; what matters is the ratios: memory is 100 times a
// cache hit, an SSD read is a thousand memory reads, a cross-continent round
// trip is a million.
//
// Time: O(1). Space: O(1).
func LatencyNumbers() Report {
	rows := []struct {
		what string
		ns   int
	}{
		{"L1 cache reference (about 0.5 ns)", 1},
		{"branch mispredict", 5},
		{"L2 cache reference", 7},
		{"mutex lock or unlock", 25},
		{"main memory reference", 100},
		{"compress 1 KB with a fast compressor", 3_000},
		{"send 1 KB over a 1 Gbps network", 10_000},
		{"read 4 KB at random from an SSD", 150_000},
		{"read 1 MB sequentially from memory", 250_000},
		{"round trip within a datacenter", 500_000},
		{"read 1 MB sequentially from an SSD", 1_000_000},
		{"disk seek", 10_000_000},
		{"read 1 MB sequentially from disk", 20_000_000},
		{"send a packet California to the Netherlands and back", 150_000_000},
	}
	var r Report
	t := r.table("time (ns)", "x memory reference")
	for _, row := range rows {
		t.add(row.what, row.ns, max(row.ns/100, 0))
	}
	r.say("memory is about 100x an L1 hit; an SSD read is about 1,500 memory reads; a disk seek is 100,000; a round trip across the Atlantic is 1.5 million")
	r.say("so: keep hot data in memory, read sequentially, and avoid round trips across regions on the request path")
	return r
}

// PowersOfTwo is the table of powers of two used to size data, from 2^10 (a
// kilobyte, about a thousand) to 2^60 (an exabyte, about a quintillion),
// with the exact value and its approximate decimal name.
//
// Time: O(1). Space: O(1).
func PowersOfTwo() Report {
	var r Report
	t := r.table("power", "approximate (10^n)")
	names := []struct{ prefix, count string }{
		{"kilo", "thousand"}, {"mega", "million"}, {"giga", "billion"}, {"tera", "trillion"}, {"peta", "quadrillion"}, {"exa", "quintillion"},
	}
	for i, n := range names {
		p := 10 * (i + 1)
		t.add(fmt.Sprintf("2^%d = %d bytes: 1 %sbyte", p, int64(1)<<p, n.prefix), p, 3*(i+1))
		r.say("2^%d is about 10^%d: a %s", p, 3*(i+1), n.count)
	}
	r.say("a 64-bit integer holds 2^63, about 9.2 x 10^18: room for a Snowflake ID's millisecond timestamp for 69 years")
	return r
}
