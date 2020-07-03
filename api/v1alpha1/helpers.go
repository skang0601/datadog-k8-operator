package v1alpha1

import "encoding/json"

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func toJsonNumber(n string) *json.Number {
	var number *json.Number

	if n == "" {
		number = nil
	} else {
		temp := json.Number(n)
		number = &temp
	}
	return number
}
