package compare

func Difference(master, slave []string) (diff []string) {
	m := make(map[string]bool, len(slave))

	for _, item := range slave {
		m[item] = true
	}

	for _, item := range master {
		if !m[item] {
			m[item] = true
			diff = append(diff, item)
		}
	}
	return
}

func Intersection(master, slave []string) (diff []string) {
	m := make(map[string]bool, len(slave))

	for _, item := range slave {
		m[item] = true
	}

	for _, item := range master {
		if m[item] {
			m[item] = false
			diff = append(diff, item)
		}
	}
	return
}
