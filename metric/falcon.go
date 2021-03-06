package metric

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mia0x75/go-metrics"
	"github.com/mia0x75/gopfc/g"
	bhttp "github.com/niean/gotools/http/httpclient/beego"
)

const (
	GAUGE = "GAUGE"
)

func PushToFalcon() {
	cfg := g.Config()
	step := cfg.Step
	api := cfg.Push.Api
	debug := cfg.Debug

	// align push start ts
	alignPushStartTs(step)

	for _ = range time.Tick(time.Duration(step) * time.Second) {
		selfMeter("pfc.push.cnt", 1) // statistics

		fms := FalconMetrics()
		start := time.Now()
		err := push(fms, api, debug)
		selfGauge("pfc.push.ms", int64(time.Since(start)/time.Millisecond)) // statistics

		if err != nil {
			if debug {
				log.Printf("[perfcounter] send to %s error: %v", api, err)
			}
			selfGauge("pfc.push.size", int64(0)) // statistics
		} else {
			selfGauge("pfc.push.size", int64(len(fms))) // statistics
		}
	}
}

func FalconMetric(types []string) []*MetricValue {
	fd := []*MetricValue{}
	for _, ty := range types {
		if r, ok := values[ty]; ok && r != nil {
			data := _falconMetric(r)
			fd = append(fd, data...)
		}
	}
	return fd
}

func FalconMetrics() []*MetricValue {
	data := make([]*MetricValue, 0)
	for _, r := range values {
		nd := _falconMetric(r)
		data = append(data, nd...)
	}
	return data
}

// internal
func _falconMetric(r metrics.Registry) []*MetricValue {
	cfg := g.Config()
	endpoint := cfg.Hostname
	step := cfg.Step
	prefix := cfg.Prefix
	tags := cfg.Tags
	ts := time.Now().Unix()

	data := make([]*MetricValue, 0)
	r.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Gauge:
			data = append(data, &MetricValue{
				Endpoint:  endpoint,
				Metric:    fmt.Sprintf("%s.%s", prefix, name),
				Value:     metric.Value(),
				Step:      step,
				Type:      GAUGE,
				Tags:      getTags(tags, "value"),
				Timestamp: ts,
			})
		case metrics.GaugeFloat64:
			data = append(data, &MetricValue{
				Endpoint:  endpoint,
				Metric:    fmt.Sprintf("%s.%s", prefix, name),
				Value:     metric.Value(),
				Step:      step,
				Type:      GAUGE,
				Tags:      getTags(tags, "value"),
				Timestamp: ts,
			})
		case metrics.Counter:
			data = append(data, &MetricValue{
				Endpoint:  endpoint,
				Metric:    fmt.Sprintf("%s.%s", prefix, name),
				Value:     metric.Count(),
				Step:      step,
				Type:      GAUGE,
				Tags:      getTags(tags, "count"),
				Timestamp: ts,
			})
		case metrics.Meter:
			m := metric.Snapshot()
			data = append(data, &MetricValue{
				Endpoint:  endpoint,
				Metric:    fmt.Sprintf("%s.%s", prefix, name),
				Value:     m.RateStep(),
				Step:      step,
				Type:      GAUGE,
				Tags:      getTags(tags, "rate"),
				Timestamp: ts,
			})
			data = append(data, &MetricValue{
				Endpoint:  endpoint,
				Metric:    fmt.Sprintf("%s.%s", prefix, name),
				Value:     m.Count(),
				Step:      step,
				Type:      GAUGE,
				Tags:      getTags(tags, "sum"),
				Timestamp: ts,
			})
		case metrics.Histogram:
			m := metric.Snapshot()
			values := make(map[string]interface{})
			ps := metric.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
			values["min"] = m.Min()
			values["max"] = m.Max()
			values["mean"] = m.Mean()
			values["50th"] = ps[0]
			values["75th"] = ps[1]
			values["95th"] = ps[2]
			values["99th"] = ps[3]
			values["999th"] = ps[4]
			for key, val := range values {
				data = append(data, &MetricValue{
					Endpoint:  endpoint,
					Metric:    fmt.Sprintf("%s.%s", prefix, name),
					Value:     val,
					Step:      step,
					Type:      GAUGE,
					Tags:      getTags(tags, key),
					Timestamp: ts,
				})
			}
		}
	})

	return data
}

func getTags(tags, catalog string) string {
	if tags == "" {
		return fmt.Sprintf("catalog=%s", catalog)
	}
	return fmt.Sprintf("%s,catalog=%s", tags, catalog)
}

//
func push(data []*MetricValue, url string, debug bool) error {
	dlen := len(data)
	pkg := 200 //send pkg items once
	sent := 0
	for {
		if sent >= dlen {
			break
		}

		end := sent + pkg
		if end > dlen {
			end = dlen
		}

		pkgData := data[sent:end]
		jr, err := json.Marshal(pkgData)
		if err != nil {
			return err
		}

		response, err := bhttp.Post(url).Body(jr).String()
		if err != nil {
			return err
		}
		sent = end

		if debug {
			log.Printf("[perfcounter] push result: %v, data: %v\n", response, pkgData)
		}
	}
	return nil
}

//
func alignPushStartTs(stepSec int64) {
	nw := time.Duration(time.Now().UnixNano())
	step := time.Duration(stepSec) * time.Second
	sleepNano := step - nw%step
	if sleepNano > 0 {
		time.Sleep(sleepNano)
	}
}

//
type MetricValue struct {
	Endpoint  string      `json:"endpoint"`
	Metric    string      `json:"metric"`
	Value     interface{} `json:"value"`
	Step      int64       `json:"step"`
	Type      string      `json:"counterType"`
	Tags      string      `json:"tags"`
	Timestamp int64       `json:"timestamp"`
}

func (this *MetricValue) String() string {
	return fmt.Sprintf(
		"<Endpoint:%s, Metric:%s, Tags:%s, Type:%s, Step:%d, Timestamp:%d, Value:%v>",
		this.Endpoint,
		this.Metric,
		this.Tags,
		this.Type,
		this.Step,
		this.Timestamp,
		this.Value,
	)
}
