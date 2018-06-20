package utils

func ArrayContains(arr []string, search string) bool {
	return ArrayIndex(arr, search) != -1
}
func ArrayIndex(arr []string, search string) int {
	for i, v := range arr {
		if v == search {
			return i
		}
	}
	return -1
}
