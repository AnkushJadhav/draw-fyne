package main

import (
	"image/color"
	"image/color/palette"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	width  = 1200
	height = 700
)

var colors = palette.Plan9

type Draw struct {
	widget.BaseWidget

	currColor    color.Color
	strokeWidth  int
	history      [][]color.Color
	canvas       *canvas.Raster
	size         fyne.Size
	mousePressed bool
	lastPos      fyne.Position
}

type drawRenderer struct {
	canvas  *canvas.Raster
	objects []fyne.CanvasObject
	draw    *Draw
}

func NewDraw(size fyne.Size) *Draw {
	h := make([][]color.Color, int(size.Height)*2)
	for i := 0; i < int(size.Height)*2; i++ {
		t := make([]color.Color, int(size.Width)*2)
		for j := 0; j < int(size.Width)*2; j++ {
			t[j] = color.RGBA{255, 255, 255, 1}
		}
		h[i] = t
	}
	d := &Draw{size: size, mousePressed: false, history: h, strokeWidth: 0, lastPos: fyne.Position{}, currColor: color.RGBA{0, 0, 0, 1}}
	d.ExtendBaseWidget(d)
	return d
}

func (d *Draw) Clear() {
	for i := 0; i < len(d.history); i++ {
		for j := 0; j < len(d.history[i]); j++ {
			d.history[i][j] = color.RGBA{255, 255, 255, 1}
		}
	}
	d.canvas.Refresh()
}

func (d *Draw) canvasRenderer() func(int, int, int, int) color.Color {
	return func(x int, y int, w int, h int) color.Color {
		return d.history[y][x]
	}
}

func (d *Draw) CreateRenderer() fyne.WidgetRenderer {
	d.ExtendBaseWidget(d)
	c := canvas.NewRasterWithPixels(d.canvasRenderer())
	c.Resize(d.size)

	d.canvas = c
	dr := &drawRenderer{canvas: c, objects: []fyne.CanvasObject{c}, draw: d}
	return dr
}

func (d *Draw) MouseDown(m *desktop.MouseEvent) {
	d.mousePressed = true
	py, px := m.Position.Y*2, m.Position.X*2
	drawPointWithWidth(d.history, int(px), int(py), d.strokeWidth, d.currColor)
	d.canvas.Refresh()
	cur := fyne.NewPos(px, py)
	d.lastPos = cur
}

func (d *Draw) MouseUp(m *desktop.MouseEvent) {
	d.mousePressed = false
	d.lastPos = fyne.Position{}
}

func (d *Draw) MouseIn(m *desktop.MouseEvent) {
	if d.mousePressed {
		py, px := m.Position.Y*2, m.Position.X*2
		cur := fyne.NewPos(px, py)
		d.lastPos = cur
	}
}

func drawPoint(history [][]color.Color, px int, py int, color color.Color) {
	if py > 0 && px > 0 && py < len(history) && px < len(history[py]) {
		history[py][px] = color
	}
}

func drawPointWithWidth(history [][]color.Color, px int, py int, width int, color color.Color) {
	for i := -width; i <= width; i++ {
		for j := -width; j <= width; j++ {
			drawPoint(history, px+i, py+j, color)
		}
	}
}

func (d *Draw) blaDraw(last fyne.Position, cur fyne.Position) {
	dx := cur.X - last.X
	dy := cur.Y - last.Y
	if dx < 0 {
		dx = 0 - dx
	}
	if dy < 0 {
		dy = 0 - dy
	}
	var sx int32
	var sy int32
	if last.X < cur.X {
		sx = 1
	} else {
		sx = -1
	}
	if last.Y < cur.Y {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	cx := last.X
	cy := last.Y
	for {
		drawPointWithWidth(d.history, int(cx), int(cy), d.strokeWidth, d.currColor)
		d.canvas.Refresh()
		if (cx == cur.X) && (cy == cur.Y) {
			break
		}
		e2 := 2 * err
		if e2 > (0 - dy) {
			err = err - dy
			cx = cx + float32(sx)
		}
		if e2 < dx {
			err = err + dx
			cy = cy + float32(sy)
		}
	}
	return
}

func (d *Draw) MouseMoved(m *desktop.MouseEvent) {
	if d.mousePressed {
		py, px := m.Position.Y*2, m.Position.X*2
		cur := fyne.NewPos(px, py)
		last := d.lastPos
		if last.IsZero() {
			last = cur
		}
		d.blaDraw(last, cur)
		d.lastPos = cur
	}
}

func (d *Draw) MouseOut() {}

func (dr *drawRenderer) Layout(s fyne.Size) {
	posX, posY := 0, 0
	dr.canvas.Move(fyne.NewPos(float32(posX), float32(posY)))
	dr.canvas.Refresh()
}

func (dr *drawRenderer) MinSize() fyne.Size {
	return dr.draw.size
}

func (dr *drawRenderer) Refresh() {
	dr.canvas.Refresh()
}

func (dr *drawRenderer) Objects() []fyne.CanvasObject {
	return dr.objects
}

func (dr *drawRenderer) Destroy() {
}

func main() {
	a := app.New()
	w := a.NewWindow("Draw")

	cp := widget.NewList(func() int {
		return len(colors)
	}, func() fyne.CanvasObject {
		c := canvas.NewRectangle(colors[0])
		c.Resize(fyne.NewSize(50, 50))
		return c
	}, func(lii widget.ListItemID, co fyne.CanvasObject) {
		co.(*canvas.Rectangle).FillColor = colors[lii]
		co.Refresh()
	})
	d := NewDraw(fyne.NewSize(float32(width), float32(height)))

	e := widget.NewEntry()
	e.OnChanged = func(s string) {
		dr, err := strconv.Atoi(s)
		if err != nil {
			e.SetText("")
		} else {
			if dr > 50 {
				e.SetText("50")
				dr = 50
			}
			d.strokeWidth = dr
		}
	}

	l := widget.NewLabel("Size")

	cp.OnSelected = func(id widget.ListItemID) {
		d.currColor = colors[id]
	}
	b := widget.NewButtonWithIcon("Clear", theme.CancelIcon(), func() {
		d.Clear()
	})
	s := widget.NewSeparator()
	tb := container.NewHBox(l, e, s, b)
	c := container.NewBorder(tb, nil, cp, nil, d)

	w.CenterOnScreen()
	w.SetContent(c)
	w.ShowAndRun()
}
