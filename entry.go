package skkdic

type entry struct {
	Midashi    string
	Candidates []*Candidate

	andSet map[string]struct{}
}

func (e *entry) IsOkuriAri() bool {
	return isOkuriAri(e.Midashi)
}

func (e *entry) addCandidate(c *Candidate, delimiter string) {
	for _, candidate := range e.Candidates {
		if candidate.Text == c.Text {
			candidate.joinAnnotation(c.Annotation, delimiter)
			return
		}
	}
	e.Candidates = append(e.Candidates, c)
}

func (e *entry) removeCandidateAt(i int) {
	if i < 0 || i >= len(e.Candidates) {
		return
	}
	e.Candidates = append(e.Candidates[:i], e.Candidates[i+1:]...)
}

func (e *entry) removeCandidate(candidate *Candidate) {
	for i, c := range e.Candidates {
		if c.Text == candidate.Text {
			e.removeCandidateAt(i)
			break
		}
	}
}

func (e *entry) andCandidate(c *Candidate) {
	for _, candidate := range e.Candidates {
		if candidate.Text == c.Text {
			if e.andSet == nil {
				e.andSet = map[string]struct{}{}
			}
			e.andSet[candidate.Text] = struct{}{}
			return
		}
	}
}

func (e *entry) clean() {
	for i := 0; i < len(e.Candidates); {
		c := e.Candidates[i]
		_, ok := e.andSet[c.Text]
		if ok {
			i++
		} else {
			e.removeCandidateAt(i)
		}
	}
	e.andSet = nil
}

func lessEntry(a, b *entry) bool {
	return a.Midashi < b.Midashi
}

func lessEntryReverse(a, b *entry) bool {
	return b.Midashi < a.Midashi
}
