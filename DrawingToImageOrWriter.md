`Plot.Save()` makes it easy to save a plot to a file. However, often one wants to plot directly to an `image.Image` or an `io.Writer`. This is possible. The trick is to create your own `plot.DrawArea`. The following examples illustrate.

## Drawing to an image.Image ##
```
package main

import (
	"image"
	"image/png"
	"os"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/vg/vgimg"
)

const dpi = 96

func main() {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	l, err := plotter.NewLine(plotter.XYs{{0, 0}, {1, 1}, {2, 2}})
	if err != nil {
		panic(err)
	}
	p.Add(l)

	// Draw the plot to an in-memory image.
	img := image.NewRGBA(image.Rect(0, 0, 3*dpi, 3*dpi))
	da := plot.MakeDrawArea(vgimg.NewImage(img))
	p.Draw(da)

	// Same the image.
	f, err := os.Create("test.png")
	if err != nil {
		panic(err)
	}
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
}
```

## Writing a plot to an io.Writer ##
```
package main

import (
	"os"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/vg"
	"code.google.com/p/plotinum/vg/vgsvg"
)

const dpi = 96

func main() {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	l, err := plotter.NewLine(plotter.XYs{{0, 0}, {1, 1}, {2, 2}})
	if err != nil {
		panic(err)
	}
	p.Add(l)

	// Create a Canvas for writing SVG images.
	c := vgsvg.New(vg.Inches(3), vg.Inches(3))

	// Draw to the Canvas.
	da := plot.MakeDrawArea(c)
	p.Draw(da)

	// Write the Canvas to a io.Writer (in this case, os.Stdout).
	if _, err := c.WriteTo(os.Stdout); err != nil {
		panic(err)
	}
}
```