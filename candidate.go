package skkdic

import "strings"

type Candidate struct {
	Text       string
	Annotation string
}

func parseCandidates(midashi string, s string) []*Candidate {
	count := 0
	p, q := 0, 0
	for p < len(s) && s[p] >= 0x20 {
		if s[p] == '/' {
			p++
			q = p
			if q >= len(s) || s[q] < 0x20 {
				break
			}
			if s[q] == '[' {
				for q < len(s) && s[q] != ']' {
					q++
				}
				p = q
				continue
			}
			for q < len(s) && s[q] != '/' {
				q++
			}
			if p == q {
				continue
			}
			count++
			p = q
		} else {
			p++
		}
	}

	if count == 0 {
		return nil
	}

	candidates := make([]*Candidate, 0, count)

	p, q = 0, 0
	for p < len(s) && s[p] >= 0x20 {
		if s[p] == '/' {
			p++
			q = p
			if q >= len(s) || s[q] < 0x20 {
				break
			}
			if s[q] == '[' {
				for q < len(s) && s[q] != ']' {
					q++
				}
				p = q
				continue
			}
			for q < len(s) && s[q] != '/' {
				q++
			}
			if p == q {
				continue
			}

			text, annotation, _ := strings.Cut(s[p:q], ";")
			candidates = append(candidates, &Candidate{
				Text:       text,
				Annotation: annotation,
			})
			p = q
		} else {
			p++
		}
	}

	return candidates
}

func (c *Candidate) String() string {
	if c.Annotation == "" {
		return c.Text
	}

	return c.Text + ";" + c.Annotation
}

func (c *Candidate) joinAnnotation(annotation, delimiter string) {
	if annotation == "" {
		return
	}

	if c.Annotation == "" {
		c.Annotation = annotation
	} else if c.Annotation != annotation {
		c.Annotation += delimiter + annotation
	}
}
