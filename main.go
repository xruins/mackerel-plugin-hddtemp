package main

import (
	"errors"
	"flag"
	"fmt"
	"os/exec"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/xruins/mackerel-plugin-hddtemp/lib/hddtemp"
	"github.com/xruins/mackerel-plugin-hddtemp/lib/smart"
)

type HDDTempPlugin struct {
	prefix  string
	devices []string
	method  Method
}

type Method string

const (
	MethodAuto     = "auto"
	MethodSmartctl = "smartctl"
	MethodHDDTemp  = "hddtemp"
)

type fetcher interface {
	Fetch([]string) (map[string]float64, error)
}

func isExecutable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return (err == nil)
}

func (htp *HDDTempPlugin) FetchMetrics() (map[string]float64, error) {
	var f fetcher

	switch htp.method {
	case MethodAuto:
		if isExecutable(MethodSmartctl) {
			f = &smart.SmartctlFetcher{}
			break
		} else if isExecutable(MethodHDDTemp) {
			f = &hddtemp.HDDTempFetcher{}
			break
		}
		return nil, errors.New("neither smartctl nor hddtemp executable")
	case MethodSmartctl:
		if isExecutable(MethodSmartctl) {
			f = &smart.SmartctlFetcher{}
			break
		}
		return nil, errors.New("could not find smartctl executable")
	case MethodHDDTemp:
		if isExecutable(MethodHDDTemp) {
			f = &hddtemp.HDDTempFetcher{}
			break
		}
		return nil, errors.New("could not find hddtemp executable")
	default:
		return nil, errors.New(`malformed method. choose one: "auto","smartctl","hddtemp".`)
	}

	result, err := f.Fetch(htp.devices)
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
	optMethod := flag.String("method", "auto", `method to fetch HDD temperature. choose one: "auto","smartctl","hddtemp"`)
	flag.Parse()

	hddtemp := HDDTempPlugin{
		prefix:  *optPrefix,
		devices: flag.Args(),
		method:  Method(*optMethod),
	}

	plugin := mp.NewMackerelPlugin(&hddtemp)
	plugin.Tempfile = *optTempfile
	plugin.Run()
}
