// MIT License

// Copyright (c) 2022 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package charts

import (
	"bytes"
	"errors"
	"math"

	"github.com/golang/freetype/truetype"
	"github.com/wcharczuk/go-chart/v2"
)

// ValueFormatter turns a numeric value into its string representation, for
// example when drawing axis ticks or data labels.
type ValueFormatter func(float64) string

// Painter wraps a go-chart renderer and draws chart primitives within a
// bounded Box. Painters can be nested through Child so that each chart
// component draws inside its own coordinate system.
type Painter struct {
	render chart.Renderer
	box    Box
	font   *truetype.Font
	parent *Painter
	style  Style
	theme  ColorPalette
	// 类型
	outputType     string
	valueFormatter ValueFormatter
}

// PainterOptions describes the required arguments to build a new Painter
// through NewPainter.
type PainterOptions struct {
	// Draw type, "svg" or "png", default type is "png"
	Type string
	// The width of draw painter
	Width int
	// The height of draw painter
	Height int
	// The font for painter
	Font *truetype.Font
}

// PainterOption mutates a Painter. It is used by NewPainter and Painter.Child
// to apply optional configuration such as padding, style or theme.
type PainterOption func(*Painter)

// TicksOption configures how Painter.Ticks draws tick marks.
type TicksOption struct {
	// the first tick
	First  int
	Length int
	Orient string
	Count  int
	Unit   int
}

// MultiTextOption configures how Painter.MultiText lays out a list of labels.
type MultiTextOption struct {
	TextList []string
	Orient   string
	Unit     int
	Position string
	Align    string
	// The text rotation of label
	TextRotation float64
	Offset       Box
	// The first text index
	First int
}

// GridOption configures how Painter.Grid draws grid lines.
type GridOption struct {
	Column      int
	Row         int
	ColumnSpans []int
	// 忽略不展示的column
	IgnoreColumnLines []int
	// 忽略不展示的row
	IgnoreRowLines []int
}

// PainterPaddingOption sets the padding of draw painter
func PainterPaddingOption(padding Box) PainterOption {
	return func(p *Painter) {
		p.box.Left += padding.Left
		p.box.Top += padding.Top
		p.box.Right -= padding.Right
		p.box.Bottom -= padding.Bottom
	}
}

// PainterBoxOption sets the box of draw painter
func PainterBoxOption(box Box) PainterOption {
	return func(p *Painter) {
		if box.IsZero() {
			return
		}
		p.box = box
	}
}

// PainterFontOption sets the font of draw painter
func PainterFontOption(font *truetype.Font) PainterOption {
	return func(p *Painter) {
		if font == nil {
			return
		}
		p.font = font
	}
}

// PainterStyleOption sets the style of draw painter
func PainterStyleOption(style Style) PainterOption {
	return func(p *Painter) {
		p.SetStyle(style)
	}
}

// PainterThemeOption sets the theme of draw painter
func PainterThemeOption(theme ColorPalette) PainterOption {
	return func(p *Painter) {
		if theme == nil {
			return
		}
		p.theme = theme
	}
}

// PainterWidthHeightOption set width or height of draw painter
func PainterWidthHeightOption(width, height int) PainterOption {
	return func(p *Painter) {
		if width > 0 {
			p.box.Right = p.box.Left + width
		}
		if height > 0 {
			p.box.Bottom = p.box.Top + height
		}
	}
}

// NewPainter creates a painter
func NewPainter(opts PainterOptions, opt ...PainterOption) (*Painter, error) {
	if opts.Width <= 0 || opts.Height <= 0 {
		return nil, errors.New("width/height can not be nil")
	}
	font := opts.Font
	if font == nil {
		f, err := GetDefaultFont()
		if err != nil {
			return nil, err
		}
		font = f
	}
	fn := chart.PNG
	if opts.Type == ChartOutputSVG {
		fn = chart.SVG
	}
	width := opts.Width
	height := opts.Height
	r, err := fn(width, height)
	if err != nil {
		return nil, err
	}
	r.SetFont(font)

	p := &Painter{
		render: r,
		box: Box{
			Right:  opts.Width,
			Bottom: opts.Height,
		},
		font: font,
		// 类型
		outputType: opts.Type,
	}
	p.setOptions(opt...)
	if p.theme == nil {
		p.theme = NewTheme(ThemeLight)
	}
	return p, nil
}
func (p *Painter) setOptions(opts ...PainterOption) {
	for _, fn := range opts {
		fn(p)
	}
}

// Child returns a new Painter that shares the underlying renderer with p and
// applies the provided options on top of p's configuration. It is used to
// create sub-regions (for example to add padding or select a bounding box)
// without affecting the parent.
func (p *Painter) Child(opt ...PainterOption) *Painter {
	child := &Painter{
		// 格式化
		valueFormatter: p.valueFormatter,
		// render
		render: p.render,
		box:    p.box.Clone(),
		font:   p.font,
		parent: p,
		style:  p.style,
		theme:  p.theme,
	}
	child.setOptions(opt...)
	return child
}

// SetStyle replaces the current style applied to the underlying renderer. If
// the style has no font, the Painter's own font is used.
func (p *Painter) SetStyle(style Style) {
	if style.Font == nil {
		style.Font = p.font
	}
	p.style = style
	style.WriteToRenderer(p.render)
}

func overrideStyle(defaultStyle Style, style Style) Style {
	if style.StrokeWidth == 0 {
		style.StrokeWidth = defaultStyle.StrokeWidth
	}
	if style.StrokeColor.IsZero() {
		style.StrokeColor = defaultStyle.StrokeColor
	}
	if style.StrokeDashArray == nil {
		style.StrokeDashArray = defaultStyle.StrokeDashArray
	}
	if style.DotColor.IsZero() {
		style.DotColor = defaultStyle.DotColor
	}
	if style.DotWidth == 0 {
		style.DotWidth = defaultStyle.DotWidth
	}
	if style.FillColor.IsZero() {
		style.FillColor = defaultStyle.FillColor
	}
	if style.FontSize == 0 {
		style.FontSize = defaultStyle.FontSize
	}
	if style.FontColor.IsZero() {
		style.FontColor = defaultStyle.FontColor
	}
	if style.Font == nil {
		style.Font = defaultStyle.Font
	}
	return style
}

// OverrideDrawingStyle merges the given style on top of the current one and
// applies the result to the renderer's drawing options.
func (p *Painter) OverrideDrawingStyle(style Style) *Painter {
	s := overrideStyle(p.style, style)
	p.SetDrawingStyle(s)
	return p
}

// SetDrawingStyle applies style as the renderer's drawing options (stroke,
// fill, dot, ...).
func (p *Painter) SetDrawingStyle(style Style) *Painter {
	style.WriteDrawingOptionsToRenderer(p.render)
	return p
}

// SetTextStyle applies style as the renderer's text options. If the style has
// no font, the Painter's own font is used.
func (p *Painter) SetTextStyle(style Style) *Painter {
	if style.Font == nil {
		style.Font = p.font
	}
	style.WriteTextOptionsToRenderer(p.render)
	return p
}

// OverrideTextStyle merges the given style on top of the current one and
// applies the result to the renderer's text options.
func (p *Painter) OverrideTextStyle(style Style) *Painter {
	s := overrideStyle(p.style, style)
	p.SetTextStyle(s)
	return p
}

// ResetStyle writes the Painter's base style back to the renderer, undoing
// any override applied since the last SetStyle call.
func (p *Painter) ResetStyle() *Painter {
	p.style.WriteToRenderer(p.render)
	return p
}

// Bytes returns the encoded bytes of the underlying canvas in the chart's
// output format (SVG or PNG).
func (p *Painter) Bytes() ([]byte, error) {
	buffer := bytes.Buffer{}
	err := p.render.Save(&buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), err
}

// MoveTo moves the drawing cursor to (x, y) expressed in the Painter's local
// coordinate system.
func (p *Painter) MoveTo(x, y int) *Painter {
	p.render.MoveTo(x+p.box.Left, y+p.box.Top)
	return p
}

// ArcTo draws an elliptical arc centered at (cx, cy) with radii rx and ry,
// starting at startAngle and sweeping by delta radians.
func (p *Painter) ArcTo(cx, cy int, rx, ry, startAngle, delta float64) *Painter {
	p.render.ArcTo(cx+p.box.Left, cy+p.box.Top, rx, ry, startAngle, delta)
	return p
}

// LineTo draws a straight line from the current cursor position to (x, y).
func (p *Painter) LineTo(x, y int) *Painter {
	p.render.LineTo(x+p.box.Left, y+p.box.Top)
	return p
}

// QuadCurveTo draws a quadratic Bezier curve ending at (x, y) using (cx, cy)
// as control point.
func (p *Painter) QuadCurveTo(cx, cy, x, y int) *Painter {
	p.render.QuadCurveTo(cx+p.box.Left, cy+p.box.Top, x+p.box.Left, y+p.box.Top)
	return p
}

// Pin draws a map-pin shape centered at (x, y) with the given width, filled
// with the current style.
func (p *Painter) Pin(x, y, width int) *Painter {
	r := float64(width) / 2
	y -= width / 4
	angle := chart.DegreesToRadians(15)
	box := p.box

	startAngle := math.Pi/2 + angle
	delta := 2*math.Pi - 2*angle
	p.ArcTo(x, y, r, r, startAngle, delta)
	p.LineTo(x, y)
	p.Close()
	p.FillStroke()

	startX := x - int(r)
	startY := y
	endX := x + int(r)
	endY := y
	p.MoveTo(startX, startY)

	left := box.Left
	top := box.Top
	cx := x
	cy := y + int(r*2.5)
	p.render.QuadCurveTo(cx+left, cy+top, endX+left, endY+top)
	p.Close()
	p.Fill()
	return p
}

func (p *Painter) arrow(x, y, width, height int, direction string) *Painter {
	halfWidth := width >> 1
	halfHeight := height >> 1
	if direction == PositionTop || direction == PositionBottom {
		x0 := x - halfWidth
		x1 := x0 + width
		dy := -height / 3
		y0 := y
		y1 := y0 - height
		if direction == PositionBottom {
			y0 = y - height
			y1 = y
			dy = 2 * dy
		}
		p.MoveTo(x0, y0)
		p.LineTo(x0+halfWidth, y1)
		p.LineTo(x1, y0)
		p.LineTo(x0+halfWidth, y+dy)
		p.LineTo(x0, y0)
	} else {
		x0 := x + width
		x1 := x0 - width
		y0 := y - halfHeight
		dx := -width / 3
		if direction == PositionRight {
			x0 = x - width
			dx = -dx
			x1 = x0 + width
		}
		p.MoveTo(x0, y0)
		p.LineTo(x1, y0+halfHeight)
		p.LineTo(x0, y0+height)
		p.LineTo(x0+dx, y0+halfHeight)
		p.LineTo(x0, y0)
	}
	p.FillStroke()
	return p
}

// ArrowLeft draws a left-pointing arrow head of the given size at (x, y).
func (p *Painter) ArrowLeft(x, y, width, height int) *Painter {
	p.arrow(x, y, width, height, PositionLeft)
	return p
}

// ArrowRight draws a right-pointing arrow head of the given size at (x, y).
func (p *Painter) ArrowRight(x, y, width, height int) *Painter {
	p.arrow(x, y, width, height, PositionRight)
	return p
}

// ArrowTop draws an upward-pointing arrow head of the given size at (x, y).
func (p *Painter) ArrowTop(x, y, width, height int) *Painter {
	p.arrow(x, y, width, height, PositionTop)
	return p
}

// ArrowBottom draws a downward-pointing arrow head of the given size at
// (x, y).
func (p *Painter) ArrowBottom(x, y, width, height int) *Painter {
	p.arrow(x, y, width, height, PositionBottom)
	return p
}

// Circle draws a circle of the given radius centered at (x, y).
func (p *Painter) Circle(radius float64, x, y int) *Painter {
	p.render.Circle(radius, x+p.box.Left, y+p.box.Top)
	return p
}

// Stroke strokes the current path with the active drawing style.
func (p *Painter) Stroke() *Painter {
	p.render.Stroke()
	return p
}

// Close closes the current sub-path by drawing a line back to its starting
// point.
func (p *Painter) Close() *Painter {
	p.render.Close()
	return p
}

// FillStroke fills and strokes the current path in a single operation.
func (p *Painter) FillStroke() *Painter {
	p.render.FillStroke()
	return p
}

// Fill fills the current path with the active drawing style.
func (p *Painter) Fill() *Painter {
	p.render.Fill()
	return p
}

// Width returns the width in pixels of the Painter's drawing area.
func (p *Painter) Width() int {
	return p.box.Width()
}

// Height returns the height in pixels of the Painter's drawing area.
func (p *Painter) Height() int {
	return p.box.Height()
}

// MeasureText returns the bounding box that text would occupy if rendered
// with the current text style.
func (p *Painter) MeasureText(text string) Box {
	return p.render.MeasureText(text)
}

// MeasureTextMaxWidthHeight returns the maximum width and height required to
// render any of the strings in textList with the current text style.
func (p *Painter) MeasureTextMaxWidthHeight(textList []string) (int, int) {
	maxWidth := 0
	maxHeight := 0
	for _, text := range textList {
		box := p.MeasureText(text)
		if maxWidth < box.Width() {
			maxWidth = box.Width()
		}
		if maxHeight < box.Height() {
			maxHeight = box.Height()
		}
	}
	return maxWidth, maxHeight
}

// LineStroke draws a poly-line connecting the given points. Points whose Y
// coordinate equals math.MaxInt32 are treated as breaks that split the line
// into independent segments.
func (p *Painter) LineStroke(points []Point) *Painter {
	shouldMoveTo := false
	for index, point := range points {
		x := point.X
		y := point.Y
		if y == int(math.MaxInt32) {
			p.Stroke()
			shouldMoveTo = true
			continue
		}
		if shouldMoveTo || index == 0 {
			p.MoveTo(x, y)
			shouldMoveTo = false
		} else {
			p.LineTo(x, y)
		}
	}
	p.Stroke()
	return p
}

// SmoothLineStroke draws a smoothed curve passing through the given points
// using quadratic Bezier segments.
func (p *Painter) SmoothLineStroke(points []Point) *Painter {
	prevX := 0
	prevY := 0
	// TODO 如何生成平滑的折线
	for index, point := range points {
		x := point.X
		y := point.Y
		if index == 0 {
			p.MoveTo(x, y)
		} else {
			cx := prevX + (x-prevX)/5
			cy := y + (y-prevY)/2
			p.QuadCurveTo(cx, cy, x, y)
		}
		prevX = x
		prevY = y
	}
	p.Stroke()
	return p
}

// SetBackground fills the canvas with the given color. When inside[0] is
// true, the fill is clipped to the Painter's box instead of covering the
// whole canvas.
func (p *Painter) SetBackground(width, height int, color Color, inside ...bool) *Painter {
	r := p.render
	s := chart.Style{
		FillColor: color,
	}
	// 背景色
	p.SetDrawingStyle(s)
	defer p.ResetStyle()
	if len(inside) != 0 && inside[0] {
		p.MoveTo(0, 0)
		p.LineTo(width, 0)
		p.LineTo(width, height)
		p.LineTo(0, height)
		p.LineTo(0, 0)
	} else {
		// 设置背景色不使用box，因此不直接使用Painter
		r.MoveTo(0, 0)
		r.LineTo(width, 0)
		r.LineTo(width, height)
		r.LineTo(0, height)
		r.LineTo(0, 0)
	}
	p.FillStroke()
	return p
}
// MarkLine draws a horizontal mark line starting at (x, y) and spanning
// width pixels, with a dot on the left and an arrow on the right.
func (p *Painter) MarkLine(x, y, width int) *Painter {
	arrowWidth := 16
	arrowHeight := 10
	endX := x + width
	radius := 3
	p.Circle(3, x+radius, y)
	p.render.Fill()
	p.MoveTo(x+radius*3, y)
	p.LineTo(endX-arrowWidth, y)
	p.Stroke()
	p.ArrowRight(endX, y, arrowWidth, arrowHeight)
	return p
}

// Polygon draws a regular polygon with the given number of sides, inscribed
// in the circle of the given radius and centered at center.
func (p *Painter) Polygon(center Point, radius float64, sides int) *Painter {
	points := getPolygonPoints(center, radius, sides)
	for i, item := range points {
		if i == 0 {
			p.MoveTo(item.X, item.Y)
		} else {
			p.LineTo(item.X, item.Y)
		}
	}
	p.LineTo(points[0].X, points[0].Y)
	p.Stroke()
	return p
}

// FillArea fills the polygon defined by the given points with the active
// drawing style.
func (p *Painter) FillArea(points []Point) *Painter {
	var x, y int
	for index, point := range points {
		x = point.X
		y = point.Y
		if index == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	p.Fill()
	return p
}

// Text renders body at (x, y) using the current text style.
func (p *Painter) Text(body string, x, y int) *Painter {
	p.render.Text(body, x+p.box.Left, y+p.box.Top)
	return p
}

// TextRotation renders body at (x, y) rotated by the given angle in radians.
// The rotation is cleared after the text is drawn.
func (p *Painter) TextRotation(body string, x, y int, radians float64) {
	p.render.SetTextRotation(radians)
	p.render.Text(body, x+p.box.Left, y+p.box.Top)
	p.render.ClearTextRotation()
}

// SetTextRotation sets the rotation angle, in radians, that will be applied
// to subsequent text operations.
func (p *Painter) SetTextRotation(radians float64) {
	p.render.SetTextRotation(radians)
}

// ClearTextRotation removes any text rotation previously set with
// SetTextRotation.
func (p *Painter) ClearTextRotation() {
	p.render.ClearTextRotation()
}

// TextFit renders body inside a box of the given width, wrapping on word
// boundaries. The optional textAligns argument selects the horizontal
// alignment (AlignLeft, AlignRight or AlignCenter). It returns the bounding
// box of the rendered text.
func (p *Painter) TextFit(body string, x, y, width int, textAligns ...string) chart.Box {
	style := p.style
	textWarp := style.TextWrap
	style.TextWrap = chart.TextWrapWord
	r := p.render
	lines := chart.Text.WrapFit(r, body, width, style)
	p.SetTextStyle(style)
	var output chart.Box

	textAlign := ""
	if len(textAligns) != 0 {
		textAlign = textAligns[0]
	}
	for index, line := range lines {
		if line == "" {
			continue
		}
		x0 := x
		y0 := y + output.Height()
		lineBox := r.MeasureText(line)
		switch textAlign {
		case AlignRight:
			x0 += width - lineBox.Width()
		case AlignCenter:
			x0 += (width - lineBox.Width()) >> 1
		}
		p.Text(line, x0, y0)
		output.Right = chart.MaxInt(lineBox.Right, output.Right)
		output.Bottom += lineBox.Height()
		if index < len(lines)-1 {
			output.Bottom += +style.GetTextLineSpacing()
		}
	}
	p.style.TextWrap = textWarp
	return output
}

// Ticks draws a series of tick marks as described by opt.
func (p *Painter) Ticks(opt TicksOption) *Painter {
	if opt.Count <= 0 || opt.Length <= 0 {
		return p
	}
	count := opt.Count
	first := opt.First
	width := p.Width()
	height := p.Height()
	unit := 1
	if opt.Unit > 1 {
		unit = opt.Unit
	}
	var values []int
	isVertical := opt.Orient == OrientVertical
	if isVertical {
		values = autoDivide(height, count)
	} else {
		values = autoDivide(width, count)
	}
	for index, value := range values {
		if index < first {
			continue
		}
		if (index-first)%unit != 0 {
			continue
		}
		if isVertical {
			p.LineStroke([]Point{
				{
					X: 0,
					Y: value,
				},
				{
					X: opt.Length,
					Y: value,
				},
			})
		} else {
			p.LineStroke([]Point{
				{
					X: value,
					Y: opt.Length,
				},
				{
					X: value,
					Y: 0,
				},
			})
		}
	}
	return p
}

// MultiText draws the list of labels described by opt, evenly distributed
// along the requested orientation.
func (p *Painter) MultiText(opt MultiTextOption) *Painter {
	if len(opt.TextList) == 0 {
		return p
	}
	count := len(opt.TextList)
	positionCenter := true
	showIndex := opt.Unit / 2
	if containsString([]string{
		PositionLeft,
		PositionTop,
	}, opt.Position) {
		positionCenter = false
		count--
		// 非居中
		showIndex = 0
	}
	width := p.Width()
	height := p.Height()
	var values []int
	isVertical := opt.Orient == OrientVertical
	if isVertical {
		values = autoDivide(height, count)
	} else {
		values = autoDivide(width, count)
	}
	isTextRotation := opt.TextRotation != 0
	offset := opt.Offset
	for index, text := range opt.TextList {
		if index < opt.First {
			continue
		}
		if opt.Unit != 0 && (index-opt.First)%opt.Unit != showIndex {
			continue
		}
		if isTextRotation {
			p.ClearTextRotation()
			p.SetTextRotation(opt.TextRotation)
		}
		box := p.MeasureText(text)
		start := values[index]
		if positionCenter {
			start = (values[index] + values[index+1]) >> 1
		}
		x := 0
		y := 0
		if isVertical {
			y = start + box.Height()>>1
			switch opt.Align {
			case AlignRight:
				x = width - box.Width()
			case AlignCenter:
				x = width - box.Width()>>1
			default:
				x = 0
			}
		} else {
			x = start - box.Width()>>1
		}
		x += offset.Left
		y += offset.Top
		p.Text(text, x, y)
	}
	if isTextRotation {
		p.ClearTextRotation()
	}
	return p
}

// Grid draws horizontal and vertical grid lines as described by opt.
func (p *Painter) Grid(opt GridOption) *Painter {
	width := p.Width()
	height := p.Height()
	drawLines := func(values []int, ignoreIndexList []int, isVertical bool) {
		for index, v := range values {
			if containsInt(ignoreIndexList, index) {
				continue
			}
			x0 := 0
			y0 := 0
			x1 := 0
			y1 := 0
			if isVertical {

				x0 = v
				x1 = v
				y1 = height
			} else {
				x1 = width
				y0 = v
				y1 = v
			}
			p.LineStroke([]Point{
				{
					X: x0,
					Y: y0,
				},
				{
					X: x1,
					Y: y1,
				},
			})
		}
	}
	columnCount := sumInt(opt.ColumnSpans)
	if columnCount == 0 {
		columnCount = opt.Column
	}
	if columnCount > 0 {
		values := autoDivideSpans(width, columnCount, opt.ColumnSpans)
		drawLines(values, opt.IgnoreColumnLines, true)
	}
	if opt.Row > 0 {
		values := autoDivide(height, opt.Row)
		drawLines(values, opt.IgnoreRowLines, false)
	}
	return p
}

// Dots draws a filled dot at each of the given points.
func (p *Painter) Dots(points []Point) *Painter {
	for _, item := range points {
		p.Circle(2, item.X, item.Y)
	}
	p.FillStroke()
	return p
}

// Rect draws a rectangle with the given bounding box and fills it using the
// active drawing style.
func (p *Painter) Rect(box Box) *Painter {
	p.MoveTo(box.Left, box.Top)
	p.LineTo(box.Right, box.Top)
	p.LineTo(box.Right, box.Bottom)
	p.LineTo(box.Left, box.Bottom)
	p.LineTo(box.Left, box.Top)
	p.FillStroke()
	return p
}

// RoundedRect draws a rectangle with rounded corners of the given radius
// and fills it using the active drawing style.
func (p *Painter) RoundedRect(box Box, radius int) *Painter {
	r := (box.Right - box.Left) / 2
	if radius > r {
		radius = r
	}
	rx := float64(radius)
	ry := float64(radius)
	p.MoveTo(box.Left+radius, box.Top)
	p.LineTo(box.Right-radius, box.Top)

	cx := box.Right - radius
	cy := box.Top + radius
	// right top
	p.ArcTo(cx, cy, rx, ry, -math.Pi/2, math.Pi/2)

	p.LineTo(box.Right, box.Bottom-radius)

	// right bottom
	cx = box.Right - radius
	cy = box.Bottom - radius
	p.ArcTo(cx, cy, rx, ry, 0.0, math.Pi/2)

	p.LineTo(box.Left+radius, box.Bottom)

	// left bottom
	cx = box.Left + radius
	cy = box.Bottom - radius
	p.ArcTo(cx, cy, rx, ry, math.Pi/2, math.Pi/2)

	p.LineTo(box.Left, box.Top+radius)

	// left top
	cx = box.Left + radius
	cy = box.Top + radius
	p.ArcTo(cx, cy, rx, ry, math.Pi, math.Pi/2)

	p.Close()
	p.FillStroke()
	p.Fill()
	return p
}

// LegendLineDot draws the small line-with-dot indicator used in legends
// inside the given box.
func (p *Painter) LegendLineDot(box Box) *Painter {
	width := box.Width()
	height := box.Height()
	strokeWidth := 3
	dotHeight := 5

	p.render.SetStrokeWidth(float64(strokeWidth))
	center := (height-strokeWidth)>>1 - 1
	p.MoveTo(box.Left, box.Top-center)
	p.LineTo(box.Right, box.Top-center)
	p.Stroke()
	p.Circle(float64(dotHeight), box.Left+width>>1, box.Top-center)
	p.FillStroke()
	return p
}

// GetRenderer returns the underlying go-chart renderer. It is mainly useful
// for interoperating with code that already depends on go-chart.
func (p *Painter) GetRenderer() chart.Renderer {
	return p.render
}
