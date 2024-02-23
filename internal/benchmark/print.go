package benchmark

import (
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"

	"github.com/wcharczuk/go-chart/v2"
)

func replaceAllRegex(str string, regex string, replace string) string {
	return regexp.MustCompile(regex).ReplaceAllString(str, replace)
}

func render(metrics []metric, config tokenBucketRunConfig) {
	accepted := chart.ContinuousSeries{
		Name: "accepted",
		Style: chart.Style{
			StrokeColor: chart.ColorAlternateGreen,
			FillColor:   chart.ColorAlternateGreen.WithAlpha(64),
		},
	}

	rejected := chart.ContinuousSeries{
		Name: "rejected",
		Style: chart.Style{
			StrokeColor: chart.ColorOrange,
			FillColor:   chart.ColorOrange.WithAlpha(64),
		},
	}

	store := chart.ContinuousSeries{
		Name:  "store ops",
		Style: chart.Style{StrokeColor: chart.ColorBlue},
	}

	target := chart.ContinuousSeries{
		Name:  "target",
		Style: chart.Style{StrokeColor: chart.ColorRed},
	}

	xTicks := []chart.Tick{}

	for _, m := range metrics {
		xTicks = append(xTicks, chart.Tick{
			Value: float64(m.time),
			Label: fmt.Sprintf("%d", m.time),
		})

		accepted.XValues = append(accepted.XValues, float64(m.time))
		accepted.YValues = append(accepted.YValues, float64(m.accepted))

		rejected.XValues = append(rejected.XValues, float64(m.time))
		rejected.YValues = append(rejected.YValues, float64(m.accepted+m.rejected))

		store.XValues = append(store.XValues, float64(m.time))
		store.YValues = append(store.YValues, float64(m.stored+m.loaded))

		target.XValues = append(target.XValues, float64(m.time))
		target.YValues = append(target.YValues, config.receivePerSecond*float64(config.keys))
	}

	series := []chart.ContinuousSeries{
		accepted, rejected, store, target,
	}

	var maxY float64
	for _, serie := range series {
		maxY = math.Max(maxY, slices.Max(serie.YValues))
	}

	roundFactor := math.Pow10(len(strconv.Itoa(int(maxY))) - 1)
	maxY = math.Ceil(maxY/float64(roundFactor)) * float64(roundFactor)

	name := fmt.Sprintf(
		"token bucket (burst=%.1f,rate=%.1f), keys=%d",
		config.receiveBurst, config.receivePerSecond, config.keys,
	)

	yTicks := chart.Ticks{}
	for i := 0.0; i <= 1; i += 0.1 {
		value := float64(i) * maxY
		tick := chart.Tick{
			Value: value,
			Label: fmt.Sprintf("%.0f", value),
		}
		yTicks = append(yTicks, tick)
	}

	graph := chart.Chart{
		Title: name,
		Width: 750,
		XAxis: chart.XAxis{
			Ticks: xTicks,
		},
		YAxis: chart.YAxis{
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: maxY,
			},
			Ticks: yTicks,
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top:  20,
				Left: 100,
			},
		},
		Series: []chart.Series{
			accepted, rejected, store, target,
		},
	}

	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph),
	}

	filename := name
	filename = replaceAllRegex(filename, `(=|\(|\)|,|\.|\s)+`, "_")
	filename = fmt.Sprintf("../../images/%s.png", filename)

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	err = graph.Render(chart.PNG, file)
	if err != nil {
		log.Fatal(err)
	}
}
