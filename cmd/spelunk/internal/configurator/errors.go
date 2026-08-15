package configurator

import (
	"errors"
)

// ErrCredentialsNotDetected indicates that credentials/configuration were not detected for a source plugin.
var ErrCredentialsNotDetected = errors.New("credentials not detected")
