package kit

func MergeMap[K comparable, V any](maps ...map[K]V) map[K]V {
	if len(maps) == 0 {
		return nil
	}
	if len(maps) == 1 {
		return maps[0]
	}
	m := make(map[K]V)
	for _, mp := range maps {
		for k, v := range mp {
			m[k] = v
		}
	}
	return m
}
