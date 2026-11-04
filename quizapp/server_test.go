package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testApp(t *testing.T, now time.Time) *quizApp {
	t.Helper()
	st, err := loadStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	app := newQuizApp(st, func() time.Time { return now })
	app.start = time.Date(2026, 11, 3, 0, 0, 0, 0, pst())
	return app
}

func body(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestHomeShowsLessonAndQuiz(t *testing.T) {
	app := testApp(t, at(2026, 11, 3, 20, 0))
	resp, err := app.fb.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	html := body(t, resp)
	for _, want := range []string{"Lesson", "Quiz", "Defining a function", `id="quiz"`, "Check answers"} {
		if !strings.Contains(html, want) {
			t.Errorf("home page missing %q", want)
		}
	}
}

func TestGradeAndReset(t *testing.T) {
	app := testApp(t, at(2026, 11, 3, 20, 0))
	day := curriculum()[0]
	req := httptest.NewRequest(http.MethodPost, "/day/1/grade", strings.NewReader(answerForm(day).Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	resp, err := app.fb.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	html := body(t, resp)
	want := fmt.Sprintf("%d of %d", len(day.questions), len(day.questions))
	if !strings.Contains(html, want) {
		t.Fatalf("score missing:\n%s", html)
	}
	if strings.Contains(html, "<html") {
		t.Fatal("htmx response should be the quiz fragment")
	}

	req = httptest.NewRequest(http.MethodPost, "/day/1/reset", nil)
	resp, err = app.fb.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	html = body(t, resp)
	if !strings.Contains(html, "Check answers") {
		t.Fatalf("reset did not restore the form:\n%s", html)
	}
}

func TestSpacedFillCounts(t *testing.T) {
	q := curriculum()[0].questions[1]
	if !q.matches([]string{"n+n", "by   rfl"}) {
		t.Fatal("spacing should not change a Lean term")
	}
	if q.matches([]string{"n * n", "rfl"}) {
		t.Fatal("a square is not double")
	}
}

func answerForm(item day) url.Values {
	form := url.Values{}
	for i, q := range item.questions {
		if q.fill() {
			for j := range q.accept {
				form.Set(fmt.Sprintf("q%db%d", i, j), q.accept[j][0])
			}
			continue
		}
		form.Set(fmt.Sprintf("q%d", i), fmt.Sprintf("%d", q.answer))
	}
	return form
}

func TestFutureDayLocked(t *testing.T) {
	app := testApp(t, at(2026, 11, 3, 20, 0))
	resp, err := app.fb.Test(httptest.NewRequest(http.MethodGet, "/day/2", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if !strings.Contains(body(t, resp), "Day 2 opens") {
		t.Fatal("expected locked message")
	}
}

func TestBeforeEightWaits(t *testing.T) {
	app := testApp(t, at(2026, 11, 3, 7, 30))
	resp, err := app.fb.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	html := body(t, resp)
	if !strings.Contains(html, "8:00am") {
		t.Fatalf("expected waiting page:\n%s", html)
	}
}
