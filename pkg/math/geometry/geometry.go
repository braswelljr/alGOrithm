// Package geometry holds computational geometry on integer points, one file
// per algorithm: three convex hulls (Graham scan, Jarvis march, Andrew's
// monotone chain), the closest pair of points, segment intersection, point in
// polygon and polygon area. Every orientation test uses an exact int64 cross
// product, so there is no floating-point error.
//
// TraceX reports what the visualizer draws: points compared (Compare), points
// looked at (Visit), the hull as it grows (Mark "hull" with the hull's point
// indexes in order), a closest or intersecting pair (Mark "pair"), and points
// found inside a polygon (Mark "inside"):
//
//	pts, _ := geometry.ParsePoints("0 0\n4 0\n2 1\n4 4\n0 4")
//	hull := geometry.ConvexHullMonotone(pts) // indexes of the hull, counterclockwise
package geometry

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Point is a point with integer coordinates.
type Point struct {
	X, Y int
}

// MaxPoints is the most points ParsePoints accepts.
const MaxPoints = 10_000

// ErrInput is returned for text that is not "x y" lines, or too many points.
var ErrInput = errors.New("geometry: invalid points")

// ParsePoints reads one point a line, "x y" (spaces, tabs or a comma between
// them). Empty lines and lines starting with # are skipped.
func ParsePoints(text string) ([]Point, error) {
	var pts []Point
	for i, line := range strings.Split(text, "\n") {
		f := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' || r == ',' || r == '\r' })
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		if len(f) != 2 {
			return nil, fmt.Errorf("%w: line %d: want \"x y\"", ErrInput, i+1)
		}
		x, err1 := strconv.Atoi(f[0])
		y, err2 := strconv.Atoi(f[1])
		if err1 != nil || err2 != nil || abs(x) > 1<<30 || abs(y) > 1<<30 {
			return nil, fmt.Errorf("%w: line %d: coordinates must be whole numbers within 2^30", ErrInput, i+1)
		}
		pts = append(pts, Point{x, y})
		if len(pts) > MaxPoints {
			return nil, fmt.Errorf("%w: more than %d points", ErrInput, MaxPoints)
		}
	}
	if len(pts) == 0 {
		return nil, fmt.Errorf("%w: no points", ErrInput)
	}
	return pts, nil
}

// Cross is the cross product of (b - a) and (c - a): positive when a, b, c
// turn counterclockwise, negative when clockwise, zero when collinear.
func Cross(a, b, c Point) int64 {
	return int64(b.X-a.X)*int64(c.Y-a.Y) - int64(b.Y-a.Y)*int64(c.X-a.X)
}

// dist2 is the squared distance between a and b.
func dist2(a, b Point) int64 {
	dx, dy := int64(a.X-b.X), int64(a.Y-b.Y)
	return dx*dx + dy*dy
}

type rec struct{ r trace.Recorder }

func (x rec) compare(i ...int) { x.r.Record(trace.Event{Kind: trace.Compare, Indices: i}) }
func (x rec) visit(i int)      { x.r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}}) }
func (x rec) mark(note string, i ...int) {
	x.r.Record(trace.Event{Kind: trace.Mark, Indices: append([]int(nil), i...), Note: note})
}

var nop = rec{trace.Nop{}}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
