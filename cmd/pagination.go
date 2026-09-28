package cmd

import "fmt"

func validatePaginationArgs(count, pages int) error {
	if count <= 0 {
		return fmt.Errorf("--count must be greater than zero")
	}
	if pages <= 0 {
		return fmt.Errorf("--pages must be greater than zero")
	}

	maxInt := int(^uint(0) >> 1)
	if count > maxInt/pages {
		return fmt.Errorf("--count multiplied by --pages is too large")
	}
	return nil
}
