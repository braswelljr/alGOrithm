package randomized

// ChiSquare returns Pearson's chi-square statistic for observed counts
// against a uniform expectation (every bucket expecting the mean count), and
// its degrees of freedom, len(counts) - 1: the sum of (observed -
// expected)^2 / expected. For a fair generator the statistic is near the
// degrees of freedom; far above them (beyond df + 3 sqrt(2 df) or so) the
// counts are too uneven to be chance. 0, 0 for fewer than two buckets or no
// counts.
//
// Time: O(k). Space: O(1).
func ChiSquare(counts []int) (stat float64, df int) {
	total := 0
	for _, c := range counts {
		total += c
	}
	if len(counts) < 2 || total == 0 {
		return 0, 0
	}
	expected := float64(total) / float64(len(counts))
	for _, c := range counts {
		d := float64(c) - expected
		stat += d * d / expected
	}
	return stat, len(counts) - 1
}
