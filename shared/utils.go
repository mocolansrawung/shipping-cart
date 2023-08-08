package shared

import (
	"fmt"
	"strconv"
)

func ConvertQueryParamsToInt(idStr string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("error converting ID parameter to integer: %w", err)
	}

	return id, nil
}
