package conda

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/pathlib"
)

func LeaseInterceptor(hash string) (string, bool, error) {
	if !common.IsLeaseRequest() {
		err := DoesLeasingAllowUsage(hash)
		if err != nil {
			return "", false, err
		}
	} else {
		leased := IsLeasedEnvironment(hash)
		if leased && CouldExtendLease(hash) {
			common.Debug("Lease of %q was extended for %q!", hash, WhoLeased(hash))
			return LiveFrom(hash), true, nil
		}
		if leased && !IsLeasePristine(hash) {
			return "", false, fmt.Errorf("Cannot get environment %q because it is dirty and leased by %q!", hash, WhoLeased(hash))
		}
		if !IsLeasedEnvironment(hash) {
			err := TakeLease(hash, common.LeaseContract)
			if err != nil {
				return "", false, err
			}
			common.Debug("Lease of %q taken by %q!", hash, WhoLeased(hash))
		}
	}
	return "", false, nil
}

func DoesLeasingAllowUsage(hash string) error {
	if !IsLeasedEnvironment(hash) || IsLeasePristine(hash) {
		return nil
	}
	reason, err := readLeaseFile(hash)
	if err != nil {
		return err
	}
	return fmt.Errorf("Environment leased to %q is dirty! Cannot use it for now!", reason)
}

func CouldExtendLease(hash string) bool {
	reason, err := readLeaseFile(hash)
	if err != nil {
		return false
	}
	if reason != common.LeaseContract {
		return false
	}
	pathlib.TouchWhen(LeaseFileFrom(hash), time.Now())
	return true
}

func TakeLease(hash, reason string) error {
	return writeLeaseFile(hash, reason)
}

func DropLease(hash, reason string) error {
	if !IsLeasedEnvironment(hash) {
		return fmt.Errorf("Not a leased environment: %q!", hash)
	}
	if reason != WhoLeased(hash) {
		return fmt.Errorf("Environment %q is not leased by %q!", hash, reason)
	}
	return os.Remove(LeaseFileFrom(hash))
}

func WhoLeased(hash string) string {
	if !IsLeasedEnvironment(hash) {
		return "~"
	}
	reason, err := readLeaseFile(hash)
	if err != nil {
		return err.Error()
	}
	return reason
}

func LeaseExpires(hash string) time.Duration {
	leasefile := LeaseFileFrom(hash)
	stamp, err := pathlib.Modtime(leasefile)
	if err != nil {
		return 0 * time.Second
	}
	deadline := stamp.Add(1 * time.Hour)
	delta := deadline.Sub(time.Now()).Round(1 * time.Second)
	if delta < 0*time.Second {
		return 0 * time.Second
	}
	return delta
}

func IsLeasedEnvironment(hash string) bool {
	leasefile := LeaseFileFrom(hash)
	exists := pathlib.IsFile(leasefile)
	if !exists {
		return false
	}
	return LeaseExpires(hash) > 0*time.Second
}

func IsLeasePristine(hash string) bool {
	return IsPristine(LiveFrom(hash))
}

func LeaseFileFrom(hash string) string {
	return fmt.Sprintf("%s.lease", LiveFrom(hash))
}

func IsSameLease(hash, reason string) bool {
	if !IsLeasedEnvironment(hash) {
		return false
	}
	content, err := readLeaseFile(hash)
	return err == nil && content == reason
}

func readLeaseFile(hash string) (string, error) {
	leasefile := LeaseFileFrom(hash)
	content, err := ioutil.ReadFile(leasefile)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writeLeaseFile(hash, reason string) error {
	leasefile := LeaseFileFrom(hash)
	flatReason := strings.TrimSpace(reason)
	return ioutil.WriteFile(leasefile, []byte(flatReason), 0o640)
}
