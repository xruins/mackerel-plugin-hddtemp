package main

import (
	"flag"
	"fmt"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/xruins/mackerel-plugin-hddtemp/lib/smart"
)

type HDDTempPlugin struct {
	prefix  string
	devices []string
}

func (htp *HDDTempPlugin) FetchMetrics() (map[string]float64, error) {
	result, err := smart.Fetch(htp.devices)
	if err != nil {
		return nil, err
	}
	metrics := make(map[string]float64, len(result))

	for k, v := range result {
		key := fmt.Sprintf("%s.temperature", k)
		metrics[key] = v
	}

	return metrics, nil
}
func (htp *HDDTempPlugin) GraphDefinition() map[string]mp.Graphs {
	return graphdef
}

func (htp *HDDTempPlugin) MetricKeyPrefix() string {
	if htp.prefix == "" {
		return "hddtemp"
	}
	return htp.prefix
}

var graphdef = map[string]mp.Graphs{
	"#": {
		Label: "HDD Temperature",
		Unit:  "float",
		Metrics: []mp.Metrics{
			{Name: "temperature", Label: "Temperature", Diff: false},
		},
	},
}

func main() {
	optPrefix := flag.String("metric-key-prefix", "", "Metric key prefix")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	hddtemp := HDDTempPlugin{
		prefix:  *optPrefix,
		devices: flag.Args(),
	}

	plugin := mp.NewMackerelPlugin(&hddtemp)
	plugin.Tempfile = *optTempfile
	plugin.Run()
}
