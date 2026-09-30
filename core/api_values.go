package core

import (
	"encoding/json"
	"math"
	"strconv"
)

// parseJSONID preserves exact string and JSON-number IDs. A float64 beyond
// JavaScript's safe-integer range may already have been rounded by decoding,
// so returning it as an ID could make a later API request target the wrong ID.
func parseJSONID(value interface{}) string {
	switch id := value.(type) {
	case string:
		return id
	case json.Number:
		if _, err := strconv.ParseInt(id.String(), 10, 64); err == nil {
			return id.String()
		}
	case float64:
		const maxSafeInteger = 1<<53 - 1
		if !math.IsNaN(id) && !math.IsInf(id, 0) && math.Trunc(id) == id && math.Abs(id) <= maxSafeInteger {
			return strconv.FormatInt(int64(id), 10)
		}
	case int:
		return strconv.Itoa(id)
	case int64:
		return strconv.FormatInt(id, 10)
	}
	return ""
}
