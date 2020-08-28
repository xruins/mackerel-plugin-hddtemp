package smart

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

type Config struct {
	devices []string
}

func Fetch(devices []string) (map[string]float64, error) {
	result, err := execSmartctl(devices)

	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]float64, len(devices))
	for dev, output := range result {
		name := removeLeadingDev(dev)
		temperature, err := getTemperature(output)
		if err != nil {
			return nil, fmt.Errorf("failed to get temperature of %s. err: %s", dev, err)
		}
		resultMap[name] = temperature
	}

	return resultMap, nil
}

// getTemperatureFromAttrStringRegexp is the regexp used to match leading digits.
var getTemperatureFromAttrStringRegexp = regexp.MustCompile("^[0-9]+")

func getTemperatureFromAttrString(s string) float64 {
	b := []byte(s)
	leadingDigits := getTemperatureFromAttrStringRegexp.Find(b)
	lds := string(leadingDigits)
	ret, _ := strconv.ParseFloat(lds, 64)
	return ret
}

func getTemperature(jsonByte []byte) (float64, error) {
	var scr smartctlResult
	err := json.Unmarshal(jsonByte, &scr)
	if err != nil {
		return 0, err
	}

	for _, tb := range scr.AtaSmartAttributes.Table {
		if strings.Contains(tb.Name, "Temperature") {
			rawString, ok := tb.Raw["string"].(string)
			if !ok {
				return 0, fmt.Errorf("malformed raw column. got: %v", tb.Raw["string"])
			}
			f := getTemperatureFromAttrString(rawString)
			if f == 0 {
				return 0, errors.New("malformed temperature value")
			}
			return f, nil
		}
	}
	return 0, errors.New("missing temperature column")
}

type table struct {
	Name string                 `json:"name"`
	Raw  map[string]interface{} `json:"raw"`
}

type ataSmartAttributes struct {
	Revision int      `json:"revision"`
	Table    []*table `json:"table"`
}

// removeLeadingDev removes leading "/dev/" from path to block device path.
func removeLeadingDev(s string) string {
	return strings.TrimPrefix(s, "/dev/")
}

// smartctlResult defines the struct to unmarshal JSON of smartctl output.
type smartctlResult struct {
	AtaSmartAttributes *ataSmartAttributes `json:"ata_smart_attributes"`
}

// execSmartctl returns the outputs of "smartctl" command for block devices.
func execSmartctl(devices []string) (map[string][]byte, error) {
	ret := make(map[string][]byte, len(devices))

	var eg errgroup.Group
	mutex := &sync.Mutex{}

	for _, dev := range devices {
		eg.Go(func() error {
			out, err := exec.Command("smartctl", "-j", "-a", dev).Output()
			if err != nil {
				return err
			}
			mutex.Lock()
			ret[dev] = out
			mutex.Unlock()

			return nil
		})
		err := eg.Wait()
		if err != nil {
			return nil, err
		}
	}

	eg.Wait()

	return ret, nil
}
