# Plotinum has moved #
See http://github.com/gonum/plot instead.


---


![https://plotinum.googlecode.com/files/logo-small.png](https://plotinum.googlecode.com/files/logo-small.png)

(Gopher by Renee French---thanks for the wonderful mascot!)

Plotinum provides an API for building and drawing plots in Go. **See [the wiki](https://code.google.com/p/plotinum/wiki/Examples) for some example plots.**

There is a discussion list on Google Groups: [plotinum-discuss@googlegroups.com](https://groups.google.com/group/plotinum-discuss).

Plotinum is split into a few packages:

  * The [plot](http://godoc.org/code.google.com/p/plotinum/plot) package provides simple interface for laying out a plot and provides primitives for drawing to it.

  * The [plotter](http://godoc.org/code.google.com/p/plotinum/plotter) package provides a standard set of 'Plotters' which use the primitives provided by the plot package for drawing lines, scatter plots, box plots, error bars, etc. to a plot.  You do not need to use the plotter package to make use of Plotinum, however: see [the wiki](https://code.google.com/p/plotinum/wiki/CreatingCustomPlotters) for a tutorial on making your own custom plotters.

  * The [plotutil](http://godoc.org/code.google.com/p/plotinum/plotutil) package contains a few routines that allow some common plot types to be made very easily.  This package is quite new so it is not as well tested as the others and it is bound to change.

  * The [vg](http://godoc.org/code.google.com/p/plotinum/vg) package provides a generic vector graphics API that sits on top of other vector graphics back-ends such as a custom EPS back-end, [draw2d](http://code.google.com/p/draw2d/), [SVGo](https://github.com/ajstarks/svgo), and [gopdf](https://bitbucket.org/zombiezen/gopdf/).

You can get Plotinum using go get:
```
go get code.google.com/p/plotinum/...
```

If you write a cool plotter that you think others may be interested in using, please email me so that I can link to it in the Plotinum wiki or possibly integrate it into the plotter package.