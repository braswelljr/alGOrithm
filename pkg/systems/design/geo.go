package design

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
)

const geohashDigits = "0123456789bcdefghjkmnpqrstuvwxyz"

// Geohash encodes a latitude and longitude as a geohash of precision
// characters: it halves the longitude and latitude ranges in turn, one bit
// each, and writes every five bits as a base-32 digit. Points that share a
// prefix are close, which turns "what is near me" into a prefix lookup.
// Each character narrows the cell: 6 characters are about 1.2 km x 0.6 km.
//
// Time: O(precision). Space: O(precision).
func Geohash(lat, lon float64, precision int) (string, error) {
	if err := check(lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180 && precision >= 1 && precision <= 12, "latitude -90 to 90, longitude -180 to 180, precision 1 to 12"); err != nil {
		return "", err
	}
	latR, lonR := [2]float64{-90, 90}, [2]float64{-180, 180}
	var b strings.Builder
	even, bit, ch := true, 0, 0
	for b.Len() < precision {
		r, v := &latR, lat
		if even {
			r, v = &lonR, lon
		}
		mid := (r[0] + r[1]) / 2
		ch <<= 1
		if v >= mid {
			ch |= 1
			r[0] = mid
		} else {
			r[1] = mid
		}
		even = !even
		if bit++; bit == 5 {
			b.WriteByte(geohashDigits[ch])
			bit, ch = 0, 0
		}
	}
	return b.String(), nil
}

// geohashCell is the size in degrees of a geohash cell of precision characters.
func geohashCell(precision int) (dLat, dLon float64) {
	bits := 5 * precision
	lonBits := (bits + 1) / 2
	return 180 / math.Pow(2, float64(bits-lonBits)), 360 / math.Pow(2, float64(lonBits))
}

// neighbors are the geohash of the point's cell and its eight neighbors.
func neighbors(lat, lon float64, precision int) []string {
	dLat, dLon := geohashCell(precision)
	seen := map[string]bool{}
	var out []string
	for _, y := range []float64{-1, 0, 1} {
		for _, x := range []float64{-1, 0, 1} {
			g, err := Geohash(math.Max(-90, math.Min(90, lat+y*dLat)), math.Max(-180, math.Min(180, lon+x*dLon)), precision)
			if err == nil && !seen[g] {
				seen[g] = true
				out = append(out, g)
			}
		}
	}
	sort.Strings(out)
	return out
}

// distanceKm is the great-circle distance between two points (haversine).
func distanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	rad := math.Pi / 180
	dLat, dLon := (lat2-lat1)*rad, (lon2-lon1)*rad
	h := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * 6371 * math.Asin(math.Sqrt(h))
}

// ProximityConfig is a city of businesses and a search.
type ProximityConfig struct {
	Businesses int     // places, spread over a city about 20 km across
	Precision  int     // geohash length of the index's cells
	Lat, Lon   float64 // where the user is
	RadiusKm   float64 // how far to search
	Seed       int64   // where the businesses are
}

// Proximity models a "places near me" service (Yelp, Google Maps): every
// business is indexed by its geohash, and a search reads the user's cell and
// its eight neighbors (a point near a cell's edge has close places in the
// next cell), then filters those few candidates by real distance. It
// reports the candidates read against the full scan a database without a
// spatial index would do. A quadtree, which splits only crowded areas, is the
// structures/spatial quadtree entry. It leaves out ranking and reviews.
//
// Time: O(Businesses) to index, O(candidates) per search. Space: O(Businesses).
func Proximity(c ProximityConfig) (Report, error) {
	if err := check(c.Businesses >= 1 && c.Businesses <= 200_000 && c.Precision >= 3 && c.Precision <= 8 && c.RadiusKm > 0, "1 to 200,000 businesses, precision 3 to 8, a positive radius"); err != nil {
		return Report{}, err
	}
	if _, err := Geohash(c.Lat, c.Lon, c.Precision); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	type place struct{ lat, lon float64 }
	index := map[string][]int{}
	places := make([]place, c.Businesses)
	for i := range places {
		places[i] = place{c.Lat + (rng.Float64()-0.5)*0.18, c.Lon + (rng.Float64()-0.5)*0.24}
		g, _ := Geohash(places[i].lat, places[i].lon, c.Precision)
		index[g] = append(index[g], i)
	}
	cells := neighbors(c.Lat, c.Lon, c.Precision)
	candidates, found := 0, 0
	for _, g := range cells {
		for _, i := range index[g] {
			candidates++
			if distanceKm(c.Lat, c.Lon, places[i].lat, places[i].lon) <= c.RadiusKm {
				found++
			}
		}
	}
	truth := 0
	for _, p := range places {
		if distanceKm(c.Lat, c.Lon, p.lat, p.lon) <= c.RadiusKm {
			truth++
		}
	}
	dLat, dLon := geohashCell(c.Precision)
	cellKm := dLat * 111
	a := arch([]Component{
		{ID: "client", Kind: Client, Label: "App"},
		{ID: "lb", Kind: Balancer, Label: "Load balancer"},
		{ID: "lbs", Kind: Service, Label: "Location service"},
		{ID: "geo", Kind: Database, Label: "Geohash index", Notes: "geohash -> business ids (read replicas)"},
		{ID: "biz", Kind: Database, Label: "Business details"},
	}, "client>lb", "lb>lbs", "lbs>geo:9 cells", "lbs>biz:details")
	f := a.frame(fmt.Sprintf("search %.1f km around the user: read cells %s", c.RadiusKm, strings.Join(cells, " ")))
	a.send(f, "client", "lb", 1, OK)
	a.send(f, "lb", "lbs", 1, OK)
	for range cells {
		a.send(f, "lbs", "geo", 1, Hit)
	}
	a.send(f, "lbs", "biz", 1, OK)
	f.load("geo", candidates, 0, -1)
	g, _ := Geohash(c.Lat, c.Lon, c.Precision)
	var r Report
	r.Arch = a
	r.say("the user is in cell %s; cells of %d characters are about %.2f km tall and %.2f degrees wide", g, c.Precision, cellKm, dLon)
	r.say("%d candidates read from %d cells instead of %d businesses; %d within %.1f km", candidates, len(cells), c.Businesses, found, c.RadiusKm)
	if found < truth {
		r.say("%d more lie within the radius but outside the 9 cells: the radius is larger than a cell, so use a shorter geohash", truth-found)
	}
	r.put("candidates read", float64(candidates))
	r.put("found", float64(found))
	r.put("full scan reads", float64(c.Businesses))
	r.put("missed (radius larger than the cells)", float64(truth-found))
	return r, nil
}

// RideConfig is a city of drivers and the riders asking for them.
type RideConfig struct {
	Drivers int   // drivers on a grid of cells
	Riders  int   // ride requests, in order
	Size    int   // the grid is Size x Size cells
	MaxRing int   // the furthest ring of cells searched before giving up
	Seed    int64 // where drivers and riders are
}

// RideMatching models how a ride-hailing service (Uber, Lyft) matches a
// rider to a driver: drivers report their location every few seconds into a
// grid index (a geohash or S2 cell in production); a request searches the
// rider's cell, then rings of cells around it, until it finds free drivers,
// and takes the nearest, who is then busy. It reports the pickup distances
// and the cells searched, and the riders left unmatched. It leaves out
// pricing, ETAs on real roads and drivers declining.
//
// Time: O(Riders * MaxRing^2). Space: O(Drivers).
func RideMatching(c RideConfig) (Report, error) {
	if err := check(c.Drivers >= 1 && c.Drivers <= 100_000 && c.Riders >= 1 && c.Riders <= 100_000 && c.Size >= 2 && c.Size <= 1000 && c.MaxRing >= 0,
		"1 to 100,000 drivers and riders, a grid of 2 to 1,000"); err != nil {
		return Report{}, err
	}
	rng := rand.New(rand.NewSource(c.Seed))
	type pos struct{ x, y int }
	cells := map[pos][]int{}
	where := make([]pos, c.Drivers)
	for d := range where {
		where[d] = pos{rng.Intn(c.Size), rng.Intn(c.Size)}
		cells[where[d]] = append(cells[where[d]], d)
	}
	busy := make([]bool, c.Drivers)
	a := arch([]Component{
		{ID: "driver", Kind: Client, Label: "Driver app"},
		{ID: "loc", Kind: Service, Label: "Location service", Notes: "updates every few seconds"},
		{ID: "index", Kind: Cache, Label: "Driver location index", Notes: "cell -> drivers"},
		{ID: "rider", Kind: Client, Label: "Rider app"},
		{ID: "match", Kind: Service, Label: "Matching service"},
		{ID: "trips", Kind: Database, Label: "Trips"},
	}, "driver>loc", "loc>index", "rider>match:request", "match>index:search rings", "match>trips:create trip", "match>driver:offer")
	matched, searched, unmatched := 0, 0, 0
	total := 0.0
	for q := 0; q < c.Riders; q++ {
		rp := pos{rng.Intn(c.Size), rng.Intn(c.Size)}
		best, bestD, ring := -1, math.MaxFloat64, 0
		for ; ring <= c.MaxRing && best < 0; ring++ {
			for y := rp.y - ring; y <= rp.y+ring; y++ {
				for x := rp.x - ring; x <= rp.x+ring; x++ {
					if max(abs(x-rp.x), abs(y-rp.y)) != ring {
						continue
					}
					searched++
					for _, d := range cells[pos{x, y}] {
						if dd := math.Hypot(float64(where[d].x-rp.x), float64(where[d].y-rp.y)); !busy[d] && dd < bestD {
							best, bestD = d, dd
						}
					}
				}
			}
		}
		var f *Frame
		if a.recording() {
			f = a.frame("")
			a.send(f, "driver", "loc", q, OK)
			a.send(f, "loc", "index", q, OK)
			a.send(f, "rider", "match", q, OK)
			a.send(f, "match", "index", q, map[bool]Status{true: Hit, false: Miss}[best >= 0])
		}
		if best < 0 {
			unmatched++
			if f != nil {
				f.Note = fmt.Sprintf("rider %d at (%d,%d): no free driver within %d rings", q, rp.x, rp.y, c.MaxRing)
			}
			continue
		}
		busy[best] = true
		matched++
		total += bestD
		if f != nil {
			f.Note = fmt.Sprintf("rider %d at (%d,%d): driver %d, %.1f cells away, found in ring %d", q, rp.x, rp.y, best, bestD, ring-1)
			a.send(f, "match", "trips", q, OK)
			a.send(f, "match", "driver", q, OK)
		}
	}
	var r Report
	r.Arch = a
	avg := 0.0
	if matched > 0 {
		avg = total / float64(matched)
	}
	r.say("%d of %d riders matched, average pickup %.2f cells; %d unmatched (no free driver nearby)", matched, c.Riders, avg, unmatched)
	r.say("%d cells searched in all, %.1f per request: the index turns a city-wide scan into a few nearby cells", searched, float64(searched)/float64(c.Riders))
	r.put("matched", float64(matched))
	r.put("unmatched", float64(unmatched))
	r.put("average pickup (cells)", math.Round(avg*100)/100)
	r.put("cells searched per request", math.Round(float64(searched)/float64(c.Riders)*10)/10)
	return r, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
