package solutions

func Intersection(x, y []int) []int {

	if len(x) == 0 || len(y) == 0 {
		return []int{}
	}

	var (
		counter map[int]int = make(map[int]int, len(x))
		result  []int       = make([]int, 0, min(len(x), len(y)))
		temp    int
		ok      bool
	)

	for _, item := range y {
		counter[item]++
	}

	for _, item := range x {
		temp, ok = counter[item]
		if !ok || temp == 0 {
			continue
		}
		result = append(result, item)
		counter[item]--
	}

	return result
}
