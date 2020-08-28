package main

import (
	"flag"

	"github.com/k0kubun/pp"
	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/xruins/mackerel-plugin-hddtemp/lib/smart"
)

type HDDTempPlugin struct {
	Prefix  string
	Devices []string
}

func (htp *HDDTempPlugin) FetchMetrics() (map[string]float64, error) {
	result, err := smart.Fetch(htp.Devices)
	if err != nil {
		return nil, err
	}
	metrics := make(map[string]float64, len(result))

	for k, v := range result {
		metrics["hddtemp.temperature."+k] = v
	}

	return metrics, nil
}
func (htp *HDDTempPlugin) GraphDefinition() map[string]mp.Graphs {
	return graphdef
}

func (htp *HDDTempPlugin) MetricKeyPrefix() string {
	if htp.Prefix != "" {
		return htp.Prefix
	}
	return ""
}

var graphdef = map[string]mp.Graphs{
	"hddtemp.temperature": {
		Label: "HDD Temperature",
		Unit:  "float",
		Metrics: []mp.Metrics{
			{Name: "*", Label: "Temperature", Diff: false},
		},
	},
}

func main() {
	optPrefix := flag.String("metric-key-prefix", "", "Metric key prefix")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	pp.Println(flag.Args())

	hddtemp := HDDTempPlugin{
		Prefix:  *optPrefix,
		Devices: flag.Args(),
	}

	plugin := mp.NewMackerelPlugin(&hddtemp)
	pp.Println(plugin.FetchMetrics())
	plugin.Tempfile = *optTempfile
	plugin.Run()
}
