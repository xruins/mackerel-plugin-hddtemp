package smart

import (
	"encoding/json"
	"os/exec"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

type SmartctlFetcher struct{}

func (s *SmartctlFetcher) Fetch(devices []string) (map[string]float64, error) {
	result, err := execSmartctl(devices)

	if err != nil {
		return nil, xerrors.Errorf("failed to fetch temperature : %w", err)
	}

	resultMap := make(map[string]float64, len(devices))
	for dev, output := range result {
		name := strings.TrimPrefix(dev, "/dev/")
		resultMap[name] = output
	}

	return resultMap, nil
}

type temperature struct {
	Current float64 `json:"current"`
}

// smartctlResult is the struct to unmarshal JSON of smartctl output.
type smartctlResult struct {
	Temperature temperature `json:"temperature"`
}

// execSmartctl returns the outputs of "smartctl" command for block devices.
func execSmartctl(devices []string) (map[string]float64, error) {
	ret := make(map[string]float64, len(devices))

	var eg errgroup.Group
	mutex := &sync.Mutex{}

	for _, dev := range devices {
		eg.Go(func() error {
			out, err := exec.Command("smartctl", "-a", "-j", dev).Output()
			if err != nil {
				return xerrors.Errorf("failed to execute smartctl command : %w", err)
			}

			var result smartctlResult
			err = json.Unmarshal(out, &result)
			if err != nil {
				return xerrors.Errorf("got malformed json : %w", err)
			}

			mutex.Lock()
			ret[dev] = result.Temperature.Current
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
