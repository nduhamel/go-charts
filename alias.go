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
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// Box is a rectangular region used to describe positions and sizes inside a
// chart. It is re-exported from the underlying go-chart package.
type Box = chart.Box

// Style groups the drawing attributes (stroke, fill, font, ...) applied when
// rendering a chart element. It is re-exported from the underlying go-chart
// package.
type Style = chart.Style

// Color represents an RGBA color. It is re-exported from the go-chart drawing
// package.
type Color = drawing.Color

// BoxZero is the zero value for a Box, useful to detect unset bounds.
var BoxZero = chart.BoxZero

// Point is a two-dimensional pixel coordinate used by the chart renderers.
type Point struct {
	X int
	Y int
}

// Chart type identifiers accepted by ChartOption and Series.
const (
	// ChartTypeLine identifies a line chart.
	ChartTypeLine = "line"
	// ChartTypeBar identifies a vertical bar chart.
	ChartTypeBar = "bar"
	// ChartTypePie identifies a pie chart.
	ChartTypePie = "pie"
	// ChartTypeRadar identifies a radar chart.
	ChartTypeRadar = "radar"
	// ChartTypeFunnel identifies a funnel chart.
	ChartTypeFunnel = "funnel"
	// ChartTypeHorizontalBar identifies a horizontal bar chart.
	ChartTypeHorizontalBar = "horizontalBar"
)

// Output formats supported by Render.
const (
	// ChartOutputSVG selects the SVG output format.
	ChartOutputSVG = "svg"
	// ChartOutputPNG selects the PNG output format.
	ChartOutputPNG = "png"
)

// Position identifiers used to place titles, legends and labels.
const (
	// PositionLeft anchors the element to the left side.
	PositionLeft = "left"
	// PositionRight anchors the element to the right side.
	PositionRight = "right"
	// PositionCenter centers the element horizontally or vertically.
	PositionCenter = "center"
	// PositionTop anchors the element to the top.
	PositionTop = "top"
	// PositionBottom anchors the element to the bottom.
	PositionBottom = "bottom"
)

// Horizontal text alignment identifiers.
const (
	// AlignLeft aligns text to the left edge.
	AlignLeft = "left"
	// AlignRight aligns text to the right edge.
	AlignRight = "right"
	// AlignCenter centers text horizontally.
	AlignCenter = "center"
)

// Orientation identifiers used by legends and similar components.
const (
	// OrientHorizontal lays components out horizontally.
	OrientHorizontal = "horizontal"
	// OrientVertical lays components out vertically.
	OrientVertical = "vertical"
)
