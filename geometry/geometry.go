/*
 * MIT License
 * 
 * Copyright 2020 Lester Kortenhoeven
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 */

package geometry

import (
	"fmt"
	"math"
)

const epsilon = 0.00001

type Value interface {
	shift(dx float64, dy float64) Value
	intersect(other Value) Value
	fmt.GoStringer
}

type nowhere struct {
}
type everywhere struct {
}
type point struct {
	x float64
	y float64
}
type line struct {
	angle float64
	d     float64
}
type lineSegment struct {
	x1 float64
	y1 float64
	x2 float64
	y2 float64
}

/* nowhere */
var Nowhere = nowhere{}

func (nw nowhere) shift(dx float64, dy float64) Value {
	return Nowhere
}
func (nw nowhere) intersect(other Value) Value {
	return Nowhere
}
func (nw nowhere) GoString() string {
	return "\"Nowhere\""
}

/* nowhere */
var Everywhere = everywhere{}

func (ew everywhere) shift(dx float64, dy float64) Value {
	return Everywhere
}
func (ew everywhere) intersect(other Value) Value {
	return other
}
func (ew everywhere) GoString() string {
	return "\"Everywhere\""
}

/* point */
func NewPoint(x float64, y float64) point {
	return point{x, y}
}
func (p point) shift(dx float64, dy float64) Value {
	return point{x: p.x + dx, y: p.y + dy}
}
func (p point) intersect(other Value) Value {
	switch ot := other.(type) {
	case nowhere:
		return Nowhere
	case everywhere:
		return p
	case point:
		if realClose(p.x, ot.x) && realClose(p.y, ot.y) {
			return p
		} else {
			return Nowhere
		}
	case line, lineSegment:
		return ot.intersect(p)
	}
	panic("Should never been reached")
}
func (p point) GoString() string {
	return fmt.Sprintf("{\"Point\":[%v,%v]}", p.x, p.y)
}

/* line: sin(angle)*x + cos(angle)*y = d */
func NewLine(angle float64, d float64) line {
	// make d positiv and angle between 0 and 2pi
	if d < 0 {
		angle = angle + math.Pi
		d = -d
	}
	angle = math.Mod(angle, 2*math.Pi)
	if angle < 0 {
		angle = angle + 2*math.Pi
	}
	return line{angle, d}
}
func (ln line) shift(dx float64, dy float64) Value {
	return line{ln.angle, ln.d + math.Sin(ln.angle)*dx + math.Cos(ln.angle)*dy}
}
func (ln line) intersect(other Value) Value {
	switch ot := other.(type) {
	case nowhere:
		return Nowhere
	case everywhere:
		return ln
	case point:
		if realClose(math.Sin(ln.angle)*ot.x+math.Cos(ln.angle)*ot.y, ln.d) {
			return ot
		} else {
			return Nowhere
		}
	case line:
		if realCloseAngle(ln.angle, ot.angle) {
			if realClose(ln.d, ot.d) {
				return ln
			} else {
				return Nowhere
			}
		} else if realCloseAngle(ln.angle, ot.angle+math.Pi) {
			if realClose(ln.d, 0) && realClose(ot.d, 0) {
				return ln
			} else {
				return Nowhere
			}
		} else {
			x := (ln.d*math.Cos(ot.angle) - ot.d*math.Cos(ln.angle)) / math.Sin(ln.angle-ot.angle)
			y := (ot.d*math.Sin(ln.angle) - ln.d*math.Sin(ot.angle)) / math.Sin(ln.angle-ot.angle)
			return point{x, y}
		}
	case lineSegment:
		return ot.intersect(ln)
	}
	panic("Should never been reached")
}
func (ln line) GoString() string {
	return fmt.Sprintf("{\"Line\":[%v,%v]}", ln.angle, ln.d)
}

/* lineSegment */
func NewLineSegment(x1 float64, y1 float64, x2 float64, y2 float64) Value {
	if realClose(x1, x2) {
		if realClose(y1, y2) {
			return point{x1, y1}
		} else if y1 < y2 {
			return lineSegment{x1, y1, x2, y2}
		} else {
			return lineSegment{x2, y2, x1, y1}
		}
	} else {
		if x1 < x2 {
			return lineSegment{x1, y1, x2, y2}
		} else {
			return lineSegment{x2, y2, x1, y1}
		}
	}
}
func (ls lineSegment) shift(dx float64, dy float64) Value {
	return lineSegment{ls.x1 + dx, ls.y1 + dy, ls.x2 + dx, ls.y2 + dy}
}
func (ls lineSegment) intersect(other Value) Value {
	switch ot := other.(type) {
	case nowhere:
		return Nowhere
	case everywhere:
		return ls
	case point:
		p := ls.toLine().intersect(ot)
		switch pt := p.(type) {
		case nowhere:
			return Nowhere
		case point:
			if between(ls.x1, pt.x, ls.x2) && between(ls.y1, pt.y, ls.y2) {
				return pt
			} else {
				return Nowhere
			}
		}
	case line:
		p := ls.toLine().intersect(ot)
		switch pt := p.(type) {
		case nowhere:
			return Nowhere
		case point:
			if between(ls.x1, pt.x, ls.x2) && between(ls.y1, pt.y, ls.y2) {
				return pt
			} else {
				return Nowhere
			}
		case line:
			return ls
		}
	case lineSegment:
		p := ls.toLine().intersect(ot)
		switch pt := p.(type) {
		case nowhere:
			return Nowhere
		case point:
			if between(ls.x1, pt.x, ls.x2) && between(ls.y1, pt.y, ls.y2) {
				return pt
			} else {
				return Nowhere
			}
		case lineSegment:
			// ls and ot ar on the same line
			if realClose(ls.x1, ot.x2) && realClose(ls.y1, ot.y2) {
				return point{ls.x1, ls.y1} // touch in one point
			} else if realClose(ls.x2, ot.x1) && realClose(ls.y2, ot.y1) {
				return point{ls.x2, ls.y2} // touch in one point
			} else if between(ls.x1, ot.x1, ls.x2) && between(ls.y1, ot.y1, ls.y2) {
				x1 := ot.x1
				y1 := ot.y1
				var x2 float64
				var y2 float64
				if between(ls.x1, ot.x2, ls.x2) && between(ls.y1, ot.y2, ls.y2) {
					x2 = ot.x2
					y2 = ot.y2
				} else {
					x2 = ls.x2
					y2 = ls.y2
				}
				return lineSegment{x1, y1, x2, y2}
			} else if between(ot.x1, ls.x1, ot.x2) && between(ot.y1, ls.y1, ot.y2) {
				x1 := ls.x1
				y1 := ls.y1
				var x2 float64
				var y2 float64
				if between(ot.x1, ls.x2, ot.x2) && between(ot.y1, ls.y2, ot.y2) {
					x2 = ls.x2
					y2 = ls.y2
				} else {
					x2 = ot.x2
					y2 = ot.y2
				}
				return lineSegment{x1, y1, x2, y2}
			} else {
				return Nowhere
			}
		}
	}
	panic("Should never been reached")
}
func (ls lineSegment) GoString() string {
	return fmt.Sprintf("{\"LineSegment\":[%v,%v,%v,%v]}", ls.x1, ls.y1, ls.x2, ls.y2)
}
func (ls lineSegment) toLine() line {
	var angle float64
	dx := ls.x1 - ls.x2
	if dx == 0 {
		angle = math.Pi / 2
	} else {
		dy := ls.y2 - ls.y1
		angle = math.Atan(dy / dx)
	}
	return line{angle, ls.x1*math.Sin(angle) + ls.y1*math.Cos(angle)}
}

func realClose(f1 float64, f2 float64) bool {
	return math.Abs(f1-f2) < epsilon
}
func realCloseAngle(f1 float64, f2 float64) bool {
	d := math.Abs(math.Mod(f1, 2*math.Pi) - math.Mod(f2, 2*math.Pi))
	return d < epsilon || (d > 2*math.Pi-epsilon && d < 2*math.Pi+epsilon) || d > 4*math.Pi-epsilon
}
func between(f1 float64, f2 float64, f3 float64) bool {
	return math.Min(f1, f3)-epsilon < f2 && f2 < math.Max(f1, f3)+epsilon
}

func Shift(dx float64, dy float64, gv Value) Value {
	return gv.shift(dx, dy)
}
func Intersect(gv1 Value, gv2 Value) Value {
	return gv1.intersect(gv2)
}
