package main

import (
	"strings"
	"testing"
	"time"
)

func pst() *time.Location { return time.FixedZone("PST", -8*3600) }

func at(y int, m time.Month, d, h, min int) time.Time {
	return time.Date(y, m, d, h, min, 0, 0, pst())
}

func TestOpenDay(t *testing.T) {
	start := time.Date(2026, 11, 3, 0, 0, 0, 0, pst())
	total := 21
	cases := []struct {
		now    time.Time
		day    int
		status status
	}{
		{at(2026, 11, 3, 7, 59), 0, statusWaiting},
		{at(2026, 11, 3, 8, 0), 1, statusOpen},
		{at(2026, 11, 3, 20, 16), 1, statusOpen},
		{at(2026, 11, 4, 7, 59), 1, statusOpen},
		{at(2026, 11, 4, 8, 0), 2, statusOpen},
		{at(2026, 11, 23, 8, 0), 21, statusOpen},
		{at(2026, 11, 24, 7, 59), 21, statusOpen},
		{at(2026, 11, 24, 8, 0), 21, statusFinished},
	}
	for _, tc := range cases {
		day, st := openDay(tc.now, start, total)
		if day != tc.day || st != tc.status {
			t.Errorf("%s: got day %d status %d, want day %d status %d", tc.now, day, st, tc.day, tc.status)
		}
	}
}

func TestCurriculum(t *testing.T) {
	days := curriculum()
	if len(days) != 21 {
		t.Fatalf("got %d days", len(days))
	}
	for i, d := range days {
		if d.title == "" || len(d.lesson) == 0 {
			t.Errorf("day %d missing lesson", i+1)
		}
		if len(d.questions) < 2 {
			t.Errorf("day %d needs more than one question", i+1)
		}
		for j, question := range d.questions {
			if question.prompt == "" || question.why == "" {
				t.Errorf("day %d question %d incomplete", i+1, j+1)
			}
			if question.fill() {
				n := strings.Count(question.code, "___")
				if n == 0 || n != len(question.accept) {
					t.Errorf("day %d question %d has %d blanks and %d answer lists", i+1, j+1, n, len(question.accept))
				}
				for _, alts := range question.accept {
					if len(alts) == 0 || strings.TrimSpace(alts[0]) == "" {
						t.Errorf("day %d question %d has an empty accepted answer", i+1, j+1)
					}
				}
				continue
			}
			if question.answer < 0 || question.answer >= len(question.choices) || len(question.choices) < 2 {
				t.Errorf("day %d question %d answer out of range", i+1, j+1)
			}
		}
	}
}
