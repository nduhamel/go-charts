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
	"math"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/wcharczuk/go-chart/v2"
)

// SeriesData represents a single data point of a Series, including its
// numeric value and an optional per-point style override.
type SeriesData struct {
	// The value of series data
	Value float64
	// The style of series data
	Style Style
}

// NewSeriesListDataFromValues returns a series list
func NewSeriesListDataFromValues(values [][]float64, chartType ...string) SeriesList {
	seriesList := make(SeriesList, len(values))
	for index, value := range values {
		seriesList[index] = NewSeriesFromValues(value, chartType...)
	}
	return seriesList
}

// NewSeriesFromValues returns a series
func NewSeriesFromValues(values []float64, chartType ...string) Series {
	s := Series{
		Data: NewSeriesDataFromValues(values),
	}
	if len(chartType) != 0 {
		s.Type = chartType[0]
	}
	return s
}

// NewSeriesDataFromValues return a series data
func NewSeriesDataFromValues(values []float64) []SeriesData {
	data := make([]SeriesData, len(values))
	for index, value := range values {
		data[index] = SeriesData{
			Value: value,
		}
	}
	return data
}

// SeriesLabel configures how data labels are rendered next to a series.
type SeriesLabel struct {
	// Data label formatter, which supports string template.
	// {b}: the name of a data item.
	// {c}: the value of a data item.
	// {d}: the percent of a data item(pie chart).
	Formatter string
	// The color for label
	Color Color
	// Show flag for label
	Show bool
	// Distance to the host graphic element.
	Distance int
	// The position of label
	Position string
	// The offset of label's position
	Offset Box
	// The font size of label
	FontSize float64
}

// Mark data type identifiers understood by SeriesMarkData.
const (
	// SeriesMarkDataTypeMax marks the maximum value of the series.
	SeriesMarkDataTypeMax = "max"
	// SeriesMarkDataTypeMin marks the minimum value of the series.
	SeriesMarkDataTypeMin = "min"
	// SeriesMarkDataTypeAverage marks the arithmetic mean of the series. It
	// is only supported by mark lines.
	SeriesMarkDataTypeAverage = "average"
)

// SeriesMarkData describes a single mark point or mark line entry.
type SeriesMarkData struct {
	// The mark data type, it can be "max", "min", "average".
	// The "average" is only for mark line
	Type string
}

// SeriesMarkPoint groups the configuration used to render mark points on a
// series.
type SeriesMarkPoint struct {
	// The width of symbol, default value is 30
	SymbolSize int
	// The mark data of series mark point
	Data []SeriesMarkData
}

// SeriesMarkLine groups the configuration used to render mark lines on a
// series.
type SeriesMarkLine struct {
	// The mark data of series mark line
	Data []SeriesMarkData
}

// Series describes a single data series to be drawn by the chart renderer.
type Series struct {
	index int
	// The type of series, it can be "line", "bar" or "pie".
	// Default value is "line"
	Type string
	// The data list of series
	Data []SeriesData
	// The Y axis index, it should be 0 or 1.
	// Default value is 0
	AxisIndex int
	// The style for series
	Style chart.Style
	// The label for series
	Label SeriesLabel
	// The name of series
	Name string
	// Radius for Pie chart, e.g.: 40%, default is "40%"
	Radius string
	// Round for bar chart
	RoundRadius int
	// Mark point for series
	MarkPoint SeriesMarkPoint
	// Make line for series
	MarkLine SeriesMarkLine
	// Max value of series
	Min *float64
	// Min value of series
	Max *float64
}
// SeriesList is an ordered collection of Series handed to ChartOption.
type SeriesList []Series

func (sl SeriesList) init() {
	if len(sl) == 0 {
		return
	}
	if sl[len(sl)-1].index != 0 {
		return
	}
	for i := 0; i < len(sl); i++ {
		if sl[i].Type == "" {
			sl[i].Type = ChartTypeLine
		}
		sl[i].index = i
	}
}

// Filter returns the subset of the list whose Type matches chartType.
func (sl SeriesList) Filter(chartType string) SeriesList {
	arr := make(SeriesList, 0)
	for index, item := range sl {
		if item.Type == chartType {
			arr = append(arr, sl[index])
		}
	}
	return arr
}

// GetMaxMin returns the maximum and minimum values found across every series
// bound to the given Y axis. Null values (see SetNullValue) are skipped.
func (sl SeriesList) GetMaxMin(axisIndex int) (float64, float64) {
	min := math.MaxFloat64
	max := -math.MaxFloat64
	for _, series := range sl {
		if series.AxisIndex != axisIndex {
			continue
		}
		for _, item := range series.Data {
			// 如果为空值，忽略
			if item.Value == nullValue {
				continue
			}
			if item.Value > max {
				max = item.Value
			}
			if item.Value < min {
				min = item.Value
			}
		}
	}
	return max, min
}

// PieSeriesOption configures NewPieSeriesList when building a pie chart
// SeriesList from a flat slice of values.
type PieSeriesOption struct {
	Radius string
	Label  SeriesLabel
	Names  []string
}

// NewPieSeriesList builds a SeriesList for a pie chart from the given values.
// Each value becomes a single-point series. When provided, opts[0] supplies
// the pie radius, label configuration and series names.
func NewPieSeriesList(values []float64, opts ...PieSeriesOption) SeriesList {
	result := make([]Series, len(values))
	var opt PieSeriesOption
	if len(opts) != 0 {
		opt = opts[0]
	}
	for index, v := range values {
		name := ""
		if index < len(opt.Names) {
			name = opt.Names[index]
		}
		s := Series{
			Type: ChartTypePie,
			Data: []SeriesData{
				{
					Value: v,
				},
			},
			Radius: opt.Radius,
			Label:  opt.Label,
			Name:   name,
		}
		result[index] = s
	}
	return result
}

type seriesSummary struct {
	// The index of max value
	MaxIndex int
	// The max value
	MaxValue float64
	// The index of min value
	MinIndex int
	// The min value
	MinValue float64
	// THe average value
	AverageValue float64
}

// Summary returns a statistical summary (min, max and average) of the data
// contained in the series.
func (s *Series) Summary() seriesSummary {
	minIndex := -1
	maxIndex := -1
	minValue := math.MaxFloat64
	maxValue := -math.MaxFloat64
	sum := float64(0)
	for j, item := range s.Data {
		if item.Value < minValue {
			minIndex = j
			minValue = item.Value
		}
		if item.Value > maxValue {
			maxIndex = j
			maxValue = item.Value
		}
		sum += item.Value
	}
	return seriesSummary{
		MaxIndex:     maxIndex,
		MaxValue:     maxValue,
		MinIndex:     minIndex,
		MinValue:     minValue,
		AverageValue: sum / float64(len(s.Data)),
	}
}

// Names returns the Name of each series in the list, preserving order.
func (sl SeriesList) Names() []string {
	names := make([]string, len(sl))
	for index, s := range sl {
		names[index] = s.Name
	}
	return names
}

// LabelFormatter label formatter
type LabelFormatter func(index int, value float64, percent float64) string

// NewPieLabelFormatter returns a pie label formatter
func NewPieLabelFormatter(seriesNames []string, layout string) LabelFormatter {
	if len(layout) == 0 {
		layout = "{b}: {d}"
	}
	return NewLabelFormatter(seriesNames, layout)
}

// NewFunnelLabelFormatter returns a funner label formatter
func NewFunnelLabelFormatter(seriesNames []string, layout string) LabelFormatter {
	if len(layout) == 0 {
		layout = "{b}({d})"
	}
	return NewLabelFormatter(seriesNames, layout)
}

// NewValueLabelFormatter returns a value formatter
func NewValueLabelFormatter(seriesNames []string, layout string) LabelFormatter {
	if len(layout) == 0 {
		layout = "{c}"
	}
	return NewLabelFormatter(seriesNames, layout)
}

// NewLabelFormatter returns a label formaatter
func NewLabelFormatter(seriesNames []string, layout string) LabelFormatter {
	return func(index int, value, percent float64) string {
		// 如果无percent的则设置为<0
		percentText := ""
		if percent >= 0 {
			percentText = humanize.FtoaWithDigits(percent*100, 2) + "%"
		}
		valueText := humanize.FtoaWithDigits(value, 2)
		name := ""
		if len(seriesNames) > index {
			name = seriesNames[index]
		}
		text := strings.ReplaceAll(layout, "{c}", valueText)
		text = strings.ReplaceAll(text, "{d}", percentText)
		text = strings.ReplaceAll(text, "{b}", name)
		return text
	}
}
