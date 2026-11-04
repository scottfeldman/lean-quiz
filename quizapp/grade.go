package main

import (
	"strconv"
	"strings"
	"unicode"
)

func (q question) fill() bool { return q.code != "" }

func (q question) slots() int {
	if q.fill() {
		return strings.Count(q.code, "___")
	}
	return 1
}

func (q question) matches(ans []string) bool {
	if len(ans) != q.slots() {
		return false
	}
	if q.fill() {
		for i, alts := range q.accept {
			if !oneOf(ans[i], alts) {
				return false
			}
		}
		return true
	}
	n, err := strconv.Atoi(strings.TrimSpace(ans[0]))
	return err == nil && n == q.answer
}

func oneOf(got string, alts []string) bool {
	for _, alt := range alts {
		if sameLean(got, alt) {
			return true
		}
	}
	return false
}

func sameLean(got, want string) bool {
	g := normLean(got)
	w := normLean(want)
	if g == w {
		return true
	}
	if strings.Contains(want, "⟨") {
		g = strings.NewReplacer("<", "⟨", ">", "⟩").Replace(g)
		if g == w {
			return true
		}
	}
	return stripSpace(g) != "" && stripSpace(g) == stripSpace(w)
}

func normLean(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func stripSpace(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
