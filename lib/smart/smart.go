package smart

import (
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
		resultMap[name] = output
	}

	return resultMap, nil
}

// removeLeadingDev removes leading "/dev/" from path to block device path.
func removeLeadingDev(s string) string {
	return strings.TrimPrefix(s, "/dev/")
}

var hddTempRegexp = regexp.MustCompile(`(\d+)°C$`)

// execSmartctl returns the outputs of "smartctl" command for block devices.
func execSmartctl(devices []string) (map[string]float64, error) {
	ret := make(map[string]float64, len(devices))

	var eg errgroup.Group
	mutex := &sync.Mutex{}

	for _, dev := range devices {
		eg.Go(func() error {
			out, err := exec.Command("hddtemp", "-n", dev).Output()
			if err != nil {
				return err
			}

			s := strings.TrimSpace(string(out))
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return fmt.Errorf("malformed temperature. got: %v err: %s", s, err)
			}

			mutex.Lock()
			ret[dev] = float64(i)
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
