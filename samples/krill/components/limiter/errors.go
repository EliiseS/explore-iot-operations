package limiter

import (
	"fmt"

	"github.com/explore-iot-ops/samples/krill/lib/errors"
)

type InvalidLimitError struct {
	errors.BadRequest
	value int
}

func (err *InvalidLimitError) Error() string {
	return fmt.Sprintf("limiter cannot have a limit of less than 1 (provided value %d)", err.value)
}

type InvalidPeriodSecondsError struct {
	errors.BadRequest
	value int
}

func (err *InvalidPeriodSecondsError) Error() string {
	return fmt.Sprintf("limiter cannot have a period seconds value of less than 1 (provided value %d)", err.value)
}
