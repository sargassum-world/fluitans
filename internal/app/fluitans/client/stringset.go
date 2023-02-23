package client

// StringSet

type StringSet map[string]struct{}

func NewStringSet(strings []string) StringSet {
	set := make(map[string]struct{})
	for _, s := range strings {
		set[s] = struct{}{}
	}
	return set
}

func (ss StringSet) Contains(set StringSet) bool {
	if ss == nil || set == nil {
		return false
	}
	if len(set) > len(ss) {
		return false
	}

	for s := range set {
		if _, ok := ss[s]; !ok {
			return false
		}
	}
	return true
}

func (ss StringSet) Equals(set StringSet) bool {
	if ss == nil || set == nil {
		return false
	}
	if len(set) != len(ss) {
		return false
	}

	// This might not be the most efficient algorithm, but it's fine for now
	return ss.Contains(set) && set.Contains(ss)
}

func (ss StringSet) Difference(set StringSet) StringSet {
	difference := make(map[string]struct{})
	for s := range ss {
		if _, ok := set[s]; !ok {
			difference[s] = struct{}{}
		}
	}
	return difference
}
