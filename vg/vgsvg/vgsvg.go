// Copyright 2012 The Plotinum Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license
// that can be found in the LICENSE file.

// The vgsvg package uses svgo (github.com/ajstarks/svgo)
// as a backend for vg.
package vgsvg

import (
	"bufio"
	"bytes"
	"code.google.com/p/plotinum/vg"
	"fmt"
	svgo "github.com/ajstarks/svgo"
	"image/color"
	"io"
	"math"
)

const (
	// inkscape, Chrome, FireFox, and gpicview all seem
	// to use 90 dots-per-inch.  I can't find anywhere that
	// this is actually specified, however.
	dpi = 90

	// pr is the amount of precision to use when outputting float64s.
	pr = 5
)

type Canvas struct {
	svg *svgo.SVG
	buf *bytes.Buffer
	ht  float64
	stk []context
}

type context struct {
	color      color.Color
	dashArray  []vg.Length
	dashOffset vg.Length
	lineWidth  vg.Length
	gEnds      int
}

func New(w, h vg.Length) *Canvas {
	buf := new(bytes.Buffer)
	c := &Canvas{
		svg: svgo.New(buf),
		buf: buf,
		ht:  w.Points(),
		stk: []context{context{}},
	}

	// This is like svg.Start, except it uses floats
	// and specifies the units.
	fmt.Fprintf(buf, `<?xml version="1.0"?>
<!-- Generated by SVGo and Plotinum VG -->
<svg width="%.*gin" height="%.*gin"
	xmlns="http://www.w3.org/2000/svg" 
	xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		pr, w.Inches(), pr, h.Inches())

	// Swap the origin to the bottom left.
	// This must be matched with a </g> when saving,
	// before the closing </svg>.
	c.svg.Gtransform(fmt.Sprintf("scale(1, -1) translate(0, -%.*g)", pr, h.Dots(c)))

	vg.Initialize(c)
	return c
}

func (c *Canvas) cur() *context {
	return &c.stk[len(c.stk)-1]
}

func (c *Canvas) SetLineWidth(w vg.Length) {
	c.cur().lineWidth = w
}

func (c *Canvas) SetLineDash(dashes []vg.Length, offs vg.Length) {
	c.cur().dashArray = dashes
	c.cur().dashOffset = offs
}

func (c *Canvas) SetColor(clr color.Color) {
	c.cur().color = clr
}

func (c *Canvas) Rotate(rot float64) {
	rot = rot * 180 / math.Pi
	c.svg.Rotate(rot)
	c.cur().gEnds++
}

func (c *Canvas) Translate(x, y vg.Length) {
	c.svg.Gtransform(fmt.Sprintf("translate(%.*g, %.*g)", pr, x.Dots(c), pr, y.Dots(c)))
	c.cur().gEnds++
}

func (c *Canvas) Scale(x, y float64) {
	c.svg.ScaleXY(x, y)
	c.cur().gEnds++
}

func (c *Canvas) Push() {
	top := *c.cur()
	top.gEnds = 0
	c.stk = append(c.stk, top)
}

func (c *Canvas) Pop() {
	for i := 0; i < c.cur().gEnds; i++ {
		c.svg.Gend()
	}
	c.stk = c.stk[:len(c.stk)-1]
}

func (c *Canvas) Stroke(path vg.Path) {
	c.svg.Path(c.pathData(path),
		style(elm("fill", "#000000", "none"),
			elm("stroke", "none", colorString(c.cur().color)),
			elm("stroke-opacity", "1", opacityString(c.cur().color)),
			elm("stroke-width", "1", "%.*g", pr, c.cur().lineWidth.Dots(c)),
			elm("stroke-dasharray", "none", dashArrayString(c)),
			elm("stroke-dashoffset", "0", "%.*g", pr, c.cur().dashOffset.Dots(c))))
}

func (c *Canvas) Fill(path vg.Path) {
	c.svg.Path(c.pathData(path),
		style(elm("fill", "#000000", colorString(c.cur().color))))
}

func (c *Canvas) pathData(path vg.Path) string {
	buf := new(bytes.Buffer)
	var x, y float64
	for _, comp := range path {
		switch comp.Type {
		case vg.MoveComp:
			fmt.Fprintf(buf, "M%.*g,%.*g", pr, comp.X.Dots(c), pr, comp.Y.Dots(c))
			x = comp.X.Dots(c)
			y = comp.Y.Dots(c)
		case vg.LineComp:
			fmt.Fprintf(buf, "L%.*g,%.*g", pr, comp.X.Dots(c), pr, comp.Y.Dots(c))
			x = comp.X.Dots(c)
			y = comp.Y.Dots(c)
		case vg.ArcComp:
			r := comp.Radius.Dots(c)
			x0 := comp.X.Dots(c) + r*math.Cos(comp.Start)
			y0 := comp.Y.Dots(c) + r*math.Sin(comp.Start)
			if x0 != x || y0 != y {
				fmt.Fprintf(buf, "L%.*g,%.*g", pr, x0, pr, y0)
			}
			if math.Abs(comp.Angle) >= 2*math.Pi {
				x, y = circle(buf, c, &comp)
			} else {
				x, y = arc(buf, c, &comp)
			}
		case vg.CloseComp:
			buf.WriteString("Z")
		default:
			panic(fmt.Sprintf("Unknown path component type: %d\n", comp.Type))
		}
	}
	return buf.String()
}

// circle adds circle path data to the given writer.
// Circles must be drawn using two arcs because
// SVG disallows the start and end point of an arc
// from being at the same location.
func circle(w io.Writer, c *Canvas, comp *vg.PathComp) (x, y float64) {
	angle := 2 * math.Pi
	if comp.Angle < 0 {
		angle = -2 * math.Pi
	}
	angle += remainder(comp.Angle, 2*math.Pi)
	if angle >= 4*math.Pi {
		panic("Impossible angle")
	}

	r := comp.Radius.Dots(c)
	x0 := comp.X.Dots(c) + r*math.Cos(comp.Start+angle/2)
	y0 := comp.Y.Dots(c) + r*math.Sin(comp.Start+angle/2)
	x = comp.X.Dots(c) + r*math.Cos(comp.Start+angle)
	y = comp.Y.Dots(c) + r*math.Sin(comp.Start+angle)

	fmt.Fprintf(w, "A%.*g,%.*g 0 %d %d %.*g,%.*g", pr, r, pr, r,
		large(angle/2), sweep(angle/2), pr, x0, pr, y0) //
	fmt.Fprintf(w, "A%.*g,%.*g 0 %d %d %.*g,%.*g", pr, r, pr, r,
		large(angle/2), sweep(angle/2), pr, x, pr, y)
	return
}

// remainder returns the remainder of x/y.
// We don't use math.Remainder because it
// seems to return incorrect values due to how
// IEEE defines the remainder operation…
func remainder(x, y float64) float64 {
	return (x/y - math.Trunc(x/y)) * y
}

// arc adds arc path data to the given writer.
// Arc can only be used if the arc's angle is
// less than a full circle, if it is greater then
// circle should be used instead.
func arc(w io.Writer, c *Canvas, comp *vg.PathComp) (x, y float64) {
	r := comp.Radius.Dots(c)
	x = comp.X.Dots(c) + r*math.Cos(comp.Start+comp.Angle)
	y = comp.Y.Dots(c) + r*math.Sin(comp.Start+comp.Angle)
	fmt.Fprintf(w, "A%.*g,%.*g 0 %d %d %.*g,%.*g", pr, r, pr, r,
		large(comp.Angle), sweep(comp.Angle), pr, x, pr, y)
	return
}

// sweep returns the arc sweep flag value for
// the given angle.
func sweep(a float64) int {
	if a < 0 {
		return 0
	}
	return 1
}

// large returns the arc's large flag value for
// the given angle.
func large(a float64) int {
	if math.Abs(a) >= math.Pi {
		return 1
	}
	return 0
}

func (c *Canvas) FillString(font vg.Font, x, y vg.Length, str string) {
	fontStr, ok := fontMap[font.Name()]
	if !ok {
		panic(fmt.Sprintf("Unknown font: %s", font.Name()))
	}
	sty := style(fontStr,
		elm("font-size", "medium", "%.*gpt", pr, font.Size.Points()),
		elm("fill", "#000000", colorString(c.cur().color)))
	if sty != "" {
		sty = "\n\t" + sty
	}
	fmt.Fprintf(c.buf, `<text x="%.*g" y="%.*g" transform="scale(1, -1)"%s>%s</text>`+"\n",
		pr, x.Dots(c), pr, -y.Dots(c), sty, str)
}

var (
	// fontMap maps Postscript-style font names to their
	// corresponding SVG style string.
	fontMap = map[string]string{
		"Courier":               "font-family:Courier;font-weight:normal;font-style:normal",
		"Courier-Bold":          "font-family:Courier;font-weight:bold;font-style:normal",
		"Courier-Oblique":       "font-family:Courier;font-weight:normal;font-style:oblique",
		"Courier-BoldOblique":   "font-family:Courier;font-weight:bold;font-style:oblique",
		"Helvetica":             "font-family:Helvetica;font-weight:normal;font-style:normal",
		"Helvetica-Bold":        "font-family:Helvetica;font-weight:bold;font-style:normal",
		"Helvetica-Oblique":     "font-family:Helvetica;font-weight:normal;font-style:oblique",
		"Helvetica-BoldOblique": "font-family:Helvetica;font-weight:bold;font-style:oblique",
		"Times-Roman":           "font-family:Times;font-weight:normal;font-style:normal",
		"Times-Bold":            "font-family:Times;font-weight:bold;font-style:normal",
		"Times-Italic":          "font-family:Times;font-weight:normal;font-style:italic",
		"Times-BoldItalic":      "font-family:Times;font-weight:bold;font-style:italic",
	}
)

func (c *Canvas) DPI() float64 {
	return dpi
}

// WriteTo writes the canvas to an io.Writer.
func (c *Canvas) WriteTo(w io.Writer) (int64, error) {
	b := bufio.NewWriter(w)
	n, err := c.buf.WriteTo(b)
	if err != nil {
		return n, err
	}

	// Close the groups and svg in the output buffer
	// so that the Canvas is not closed and can be
	// used again if needed.
	for i := 0; i < c.nEnds(); i++ {
		m, err := fmt.Fprintln(b, "</g>")
		n += int64(m)
		if err != nil {
			return n, err
		}
	}

	m, err := fmt.Fprintln(b, "</svg>\n")
	n += int64(m)
	if err != nil {
		return n, err
	}

	return n, b.Flush()
}

// nEnds returns the number of group ends
// needed before the SVG is saved.
func (c *Canvas) nEnds() int {
	n := 1 // close the transform that moves the origin
	for _, ctx := range c.stk {
		n += ctx.gEnds
	}
	return n
}

// style returns a style string composed of
// all of the given elements.  If the elements
// are all empty then the empty string is
// returned.
func style(elms ...string) string {
	str := ""
	for _, e := range elms {
		if e == "" {
			continue
		}
		if str != "" {
			str += ";"
		}
		str += e
	}
	if str == "" {
		return ""
	}
	return "style=\"" + str + "\""
}

// elm returns a style element string with the
// given key and value.  If the value matches
// default then the empty string is returned.
func elm(key, def, f string, vls ...interface{}) string {
	value := fmt.Sprintf(f, vls...)
	if value == def {
		return ""
	}
	return key + ":" + value
}

// dashArrayString returns a string representing the
// dash array specification.
func dashArrayString(c *Canvas) string {
	str := ""
	for i, d := range c.cur().dashArray {
		str += fmt.Sprintf("%.*g", pr, d.Dots(c))
		if i < len(c.cur().dashArray)-1 {
			str += ","
		}
	}
	if str == "" {
		str = "none"
	}
	return str
}

// colorString returns the hexadecimal string
// representation of the coloro.
func colorString(clr color.Color) string {
	if clr == nil {
		clr = color.Black
	}
	r, g, b, _ := clr.RGBA()
	return fmt.Sprintf("#%02X%02X%02X",
		int(float64(r)/math.MaxUint16*255),
		int(float64(g)/math.MaxUint16*255),
		int(float64(b)/math.MaxUint16*255))
}

// opacityString returns the opacity value of
// the given color.
func opacityString(clr color.Color) string {
	if clr == nil {
		clr = color.Black
	}
	_, _, _, a := clr.RGBA()
	return fmt.Sprintf("%.*g", pr, float64(a)/math.MaxUint16)
}
