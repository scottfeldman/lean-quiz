package main

import "time"

type status int

const (
	statusWaiting status = iota
	statusOpen
	statusFinished
)

func curriculumStart(loc *time.Location) time.Time {
	return time.Date(curriculumYear, curriculumMonth, curriculumDay, 0, 0, 0, 0, loc)
}

// openDay reports which lesson is available at now.
// Lessons unlock at 08:00 local time, one per calendar day, starting on curriculumStart.
// Before 08:00, the previous day's lesson is still the current one.
func openDay(now, start time.Time, total int) (int, status) {
	loc := now.Location()
	now = now.In(loc)
	y, m, d := now.Date()
	unlock := time.Date(y, m, d, 8, 0, 0, 0, loc)
	effective := time.Date(y, m, d, 12, 0, 0, 0, loc)
	if now.Before(unlock) {
		effective = effective.AddDate(0, 0, -1)
	}
	startNoon := time.Date(start.Year(), start.Month(), start.Day(), 12, 0, 0, 0, loc)
	if effective.Before(startNoon) {
		return 0, statusWaiting
	}
	n := daysBetween(effective, startNoon) + 1
	if n > total {
		return total, statusFinished
	}
	return n, statusOpen
}

func unlockedThrough(now, start time.Time, total int) int {
	day, st := openDay(now, start, total)
	switch st {
	case statusWaiting:
		return 0
	case statusFinished:
		return total
	default:
		return day
	}
}

func dayDate(start time.Time, day int) time.Time {
	y, m, d := start.Date()
	base := time.Date(y, m, d, 12, 0, 0, 0, start.Location())
	return base.AddDate(0, 0, day-1)
}

func daysBetween(later, earlier time.Time) int {
	later = time.Date(later.Year(), later.Month(), later.Day(), 12, 0, 0, 0, later.Location())
	earlier = time.Date(earlier.Year(), earlier.Month(), earlier.Day(), 12, 0, 0, 0, earlier.Location())
	return int(later.Sub(earlier).Round(24*time.Hour) / (24 * time.Hour))
}
