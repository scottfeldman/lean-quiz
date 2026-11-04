package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	g "maragu.dev/gomponents"
)

//go:embed static/*
var staticFS embed.FS

// errWritten means the handler already sent the HTTP response.
var errWritten = errors.New("response written")

type quizApp struct {
	fb    *fiber.App
	store *store
	now   func() time.Time
	start time.Time
	days  []day
}

func newQuizApp(st *store, now func() time.Time) *quizApp {
	if now == nil {
		now = time.Now
	}
	a := &quizApp{
		store: st,
		now:   now,
		days:  curriculum(),
	}
	a.start = curriculumStart(time.Local)
	fb := fiber.New()
	fb.Get("/", a.home)
	fb.Get("/day/:n", a.show)
	fb.Post("/day/:n/grade", a.grade)
	fb.Post("/day/:n/reset", a.reset)
	fb.Get("/static/:file", a.static)
	a.fb = fb
	return a
}

func (a *quizApp) home(c fiber.Ctx) error {
	now := a.now()
	n, st := openDay(now, a.start, len(a.days))
	switch st {
	case statusWaiting:
		return writeHTML(c, http.StatusOK, waitingPage(a.start, len(a.days)))
	case statusFinished:
		return writeHTML(c, http.StatusOK, finishedPage(len(a.days)))
	default:
		return a.renderDay(c, n, "")
	}
}

func (a *quizApp) show(c fiber.Ctx) error {
	n, err := a.unlocked(c)
	if err != nil {
		return written(err)
	}
	return a.renderDay(c, n, "")
}

func (a *quizApp) grade(c fiber.Ctx) error {
	n, err := a.unlocked(c)
	if err != nil {
		return written(err)
	}
	item := a.days[n-1]
	answers, complete := readAnswers(c, item)
	if !complete {
		st := dayState{Answers: answers}
		return a.renderExplicit(c, n, st, "Answer every question.")
	}
	st := dayState{Answers: answers, Submitted: true}
	if err := a.store.put(n, st); err != nil {
		return err
	}
	return a.renderExplicit(c, n, st, "")
}

func (a *quizApp) reset(c fiber.Ctx) error {
	n, err := a.unlocked(c)
	if err != nil {
		return written(err)
	}
	st := dayState{Answers: freshAnswers(a.days[n-1])}
	if err := a.store.put(n, st); err != nil {
		return err
	}
	return a.renderExplicit(c, n, st, "")
}

func (a *quizApp) renderDay(c fiber.Ctx, n int, notice string) error {
	item := a.days[n-1]
	st := a.store.get(n, item)
	return a.renderExplicit(c, n, st, notice)
}

func (a *quizApp) renderExplicit(c fiber.Ctx, n int, st dayState, notice string) error {
	now := a.now()
	today, status := openDay(now, a.start, len(a.days))
	v := quizView{
		num:      n,
		item:     a.days[n-1],
		state:    st,
		notice:   notice,
		unlocked: unlockedThrough(now, a.start, len(a.days)),
		total:    len(a.days),
		today:    today,
		st:       status,
		when:     dayDate(a.start, n).Format("Monday, January 2"),
	}
	node := dayPage(v)
	if c.Get("HX-Request") == "true" {
		node = quizSection(v)
	}
	code := http.StatusOK
	if notice != "" {
		code = http.StatusBadRequest
	}
	return writeHTML(c, code, node)
}

func (a *quizApp) unlocked(c fiber.Ctx) (int, error) {
	n, err := strconv.Atoi(c.Params("n"))
	unlocked := unlockedThrough(a.now(), a.start, len(a.days))
	if err != nil || n < 1 || n > len(a.days) {
		if werr := writeHTML(c, http.StatusNotFound, lockedPage(1, a.start, unlocked, len(a.days))); werr != nil {
			return 0, werr
		}
		return 0, errWritten
	}
	if n > unlocked {
		if werr := writeHTML(c, http.StatusNotFound, lockedPage(n, a.start, unlocked, len(a.days))); werr != nil {
			return 0, werr
		}
		return 0, errWritten
	}
	return n, nil
}

func written(err error) error {
	if errors.Is(err, errWritten) {
		return nil
	}
	return err
}

func (a *quizApp) static(c fiber.Ctx) error {
	name := path.Base(c.Params("file"))
	if name == "." || name == "/" || strings.Contains(name, "..") {
		return c.SendStatus(http.StatusNotFound)
	}
	b, err := staticFS.ReadFile("static/" + name)
	if err != nil {
		return c.SendStatus(http.StatusNotFound)
	}
	switch {
	case strings.HasSuffix(name, ".css"):
		c.Type("css")
	case strings.HasSuffix(name, ".js"):
		c.Type("js")
	}
	return c.Send(b)
}

func readAnswers(c fiber.Ctx, item day) ([][]string, bool) {
	out := make([][]string, len(item.questions))
	complete := true
	for i, q := range item.questions {
		if q.fill() {
			vals := make([]string, q.slots())
			for j := 0; j < q.slots(); j++ {
				vals[j] = c.FormValue(fmt.Sprintf("q%db%d", i, j))
				if strings.TrimSpace(vals[j]) == "" {
					complete = false
				}
			}
			out[i] = vals
			continue
		}
		raw := strings.TrimSpace(c.FormValue(fmt.Sprintf("q%d", i)))
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 || v >= len(q.choices) {
			complete = false
			out[i] = []string{""}
			continue
		}
		out[i] = []string{raw}
	}
	return out, complete
}

func writeHTML(c fiber.Ctx, status int, node g.Node) error {
	var buf bytes.Buffer
	if err := node.Render(&buf); err != nil {
		return err
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Cache-Control", "no-store")
	return c.Status(status).SendString(buf.String())
}
