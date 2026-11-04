package main

import (
	"fmt"
	"strings"
	"time"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func page(title string, body ...g.Node) g.Node {
	return Doctype(HTML(Lang("en"),
		Head(
			Meta(Charset("utf-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1")),
			TitleEl(g.Text(title)),
			Link(Rel("stylesheet"), Href("/static/app.css")),
			Script(Src("/static/htmx.min.js")),
		),
		Body(body...),
	))
}

func shell(title, when, note string, unlocked, current, total int, main ...g.Node) g.Node {
	kids := []g.Node{
		Header(
			P(Class("brand"), g.Text("Lean")),
			P(Class("when"), g.Text(when)),
		),
	}
	if note != "" {
		kids = append(kids, P(Class("note"), g.Text(note)))
	}
	kids = append(kids, dayNav(unlocked, current, total))
	kids = append(kids, main...)
	return page(title, Div(Class("wrap"), g.Group(kids)))
}

func dayNav(unlocked, current, total int) g.Node {
	items := make([]g.Node, 0, total)
	for n := 1; n <= total; n++ {
		label := fmt.Sprintf("%d", n)
		if n > unlocked {
			items = append(items, Li(Span(Class("locked"), g.Text(label))))
			continue
		}
		link := A(Href(fmt.Sprintf("/day/%d", n)), g.Text(label))
		if n == current {
			link = A(Href(fmt.Sprintf("/day/%d", n)), Aria("current", "page"), g.Text(label))
		}
		items = append(items, Li(link))
	}
	return Nav(Ul(Class("days"), g.Group(items)))
}

type quizView struct {
	num      int
	item     day
	state    dayState
	notice   string
	unlocked int
	total    int
	today    int
	st       status
	when     string
}

func dayPage(v quizView) g.Node {
	note := ""
	if v.st == statusFinished {
		note = "The starter track is complete. Pick any day to read it again."
	} else if v.num != v.today {
		note = fmt.Sprintf("Reviewing day %d. Today is day %d.", v.num, v.today)
	} else {
		note = "A new lesson unlocks every day at 8:00am."
	}
	title := fmt.Sprintf("Day %d · %s", v.num, v.item.title)
	return shell(title, v.when, note, v.unlocked, v.num, v.total,
		Div(Class("split"),
			lessonArticle(v.item),
			quizSection(v),
		),
	)
}

func lessonArticle(d day) g.Node {
	body := []g.Node{
		P(Class("kicker"), g.Text("Lesson")),
		H1(g.Text(d.title)),
	}
	for _, b := range d.lesson {
		if b.code {
			body = append(body, Pre(Code(g.Text(b.text))))
			continue
		}
		body = append(body, P(g.Text(b.text)))
	}
	return Article(Class("lesson"), g.Group(body))
}

func quizSection(v quizView) g.Node {
	head := []g.Node{
		P(Class("kicker"), g.Text("Quiz")),
		H2(g.Text("Check yourself")),
	}
	if v.notice != "" {
		head = append(head, P(Class("notice"), g.Text(v.notice)))
	}
	var form g.Node
	if v.state.Submitted {
		correct := score(v.item, v.state.Answers)
		head = append(head, P(Class("score"), g.Text(fmt.Sprintf("%d of %d", correct, len(v.item.questions)))))
		form = Form(
			Method("post"),
			Action(fmt.Sprintf("/day/%d/reset", v.num)),
			hx(fmt.Sprintf("/day/%d/reset", v.num)),
			g.Group(questionResults(v.item, v.state.Answers)),
			Button(Type("submit"), g.Text("Try again")),
		)
	} else {
		head = append(head, P(Class("hint"), g.Text("Pick one option, or complete every blank.")))
		form = Form(
			Method("post"),
			Action(fmt.Sprintf("/day/%d/grade", v.num)),
			hx(fmt.Sprintf("/day/%d/grade", v.num)),
			g.Group(questionFields(v.item, v.state.Answers)),
			Button(Type("submit"), g.Text("Check answers")),
		)
	}
	return Section(ID("quiz"), Class("quiz"), g.Group(append(head, form)))
}

func hx(url string) g.Node {
	return g.Group([]g.Node{
		g.Attr("hx-post", url),
		g.Attr("hx-target", "#quiz"),
		g.Attr("hx-swap", "outerHTML"),
	})
}

func questionFields(d day, answers [][]string) []g.Node {
	items := make([]g.Node, 0, len(d.questions))
	for i, question := range d.questions {
		ans := answerRow(answers, i)
		body := questionLead(i, question, ans, false)
		if question.fill() {
			body = append(body, fillSnippet(i, question, ans, false))
		} else {
			body = append(body, Div(Class("choices"), g.Group(choiceInputs(i, question, ans))))
		}
		items = append(items, Li(g.Group(body)))
	}
	return []g.Node{Ol(Class("questions"), g.Group(items))}
}

func questionResults(d day, answers [][]string) []g.Node {
	items := make([]g.Node, 0, len(d.questions))
	for i, question := range d.questions {
		ans := answerRow(answers, i)
		body := questionLead(i, question, ans, true)
		if question.fill() {
			body = append(body, fillSnippet(i, question, ans, true))
		} else {
			body = append(body, Div(Class("choices"), g.Group(choiceResults(question, ans))))
		}
		body = append(body, P(Class("why"), g.Text(question.why)))
		items = append(items, Li(g.Group(body)))
	}
	return []g.Node{Ol(Class("questions"), g.Group(items))}
}

func answerRow(answers [][]string, i int) []string {
	if i < len(answers) {
		return answers[i]
	}
	return nil
}

func questionLead(i int, q question, ans []string, show bool) []g.Node {
	title := fmt.Sprintf("%d. %s", i+1, q.prompt)
	head := H3(g.Text(title))
	if show {
		tag := "Wrong"
		class := "tag bad"
		if q.matches(ans) {
			tag = "Correct"
			class = "tag"
		}
		head = H3(g.Text(title), Span(Class(class), g.Text(tag)))
	}
	nodes := []g.Node{head}
	if q.stem != "" {
		nodes = append(nodes, Pre(Code(g.Text(q.stem))))
	}
	return nodes
}

func choiceInputs(i int, q question, ans []string) []g.Node {
	picked := -1
	if len(ans) == 1 {
		if v, err := parseIndex(ans[0]); err == nil {
			picked = v
		}
	}
	opts := make([]g.Node, 0, len(q.choices))
	name := fmt.Sprintf("q%d", i)
	for c, text := range q.choices {
		input := []g.Node{
			Type("radio"),
			Name(name),
			Value(fmt.Sprintf("%d", c)),
		}
		if c == 0 {
			input = append(input, Required())
		}
		if c == picked {
			input = append(input, Checked())
		}
		opts = append(opts, Label(Class("choice"),
			Input(g.Group(input)),
			Span(g.Text(text)),
		))
	}
	return opts
}

func choiceResults(q question, ans []string) []g.Node {
	picked := -1
	if len(ans) == 1 {
		if v, err := parseIndex(ans[0]); err == nil {
			picked = v
		}
	}
	opts := make([]g.Node, 0, len(q.choices))
	for c, text := range q.choices {
		class := "choice"
		var tag g.Node = g.Text("")
		switch {
		case c == q.answer && c == picked:
			class += " is-answer"
			tag = Span(Class("tag"), g.Text("Correct"))
		case c == q.answer:
			class += " is-answer"
			tag = Span(Class("tag"), g.Text("Answer"))
		case c == picked:
			class += " is-wrong"
			tag = Span(Class("tag"), g.Text("Yours"))
		}
		opts = append(opts, Div(Class(class), Span(g.Text(text), tag)))
	}
	return opts
}

func fillSnippet(i int, q question, ans []string, show bool) g.Node {
	parts := strings.Split(q.code, "___")
	nodes := make([]g.Node, 0, len(parts)*2)
	for j, part := range parts {
		if part != "" {
			nodes = append(nodes, g.Text(part))
		}
		if j == len(parts)-1 {
			break
		}
		val := ""
		if j < len(ans) {
			val = ans[j]
		}
		if !show {
			nodes = append(nodes, Input(
				Type("text"),
				Class("blank"),
				Name(fmt.Sprintf("q%db%d", i, j)),
				Value(val),
				g.Attr("spellcheck", "false"),
				AutoComplete("off"),
				Required(),
			))
			continue
		}
		ok := j < len(q.accept) && oneOf(val, q.accept[j])
		class := "hole bad"
		if ok {
			class = "hole ok"
		}
		nodes = append(nodes, Span(Class(class), g.Text(displayBlank(val))))
		if !ok && j < len(q.accept) && len(q.accept[j]) > 0 {
			nodes = append(nodes, Span(Class("expect"), g.Text(q.accept[j][0])))
		}
	}
	return Div(Class("snippet"), g.Group(nodes))
}

func displayBlank(val string) string {
	if strings.TrimSpace(val) == "" {
		return "∅"
	}
	return val
}

func parseIndex(s string) (int, error) {
	var n int
	_, err := fmt.Sscan(strings.TrimSpace(s), &n)
	return n, err
}

func score(d day, answers [][]string) int {
	n := 0
	for i, question := range d.questions {
		if question.matches(answerRow(answers, i)) {
			n++
		}
	}
	return n
}

func waitingPage(start time.Time, total int) g.Node {
	when := dayDate(start, 1).Format("Monday, January 2")
	return shell("Lean daily", "The first lesson is not open yet", "", 0, 0, total,
		Div(Class("panel"),
			H1(g.Text("See you at 8:00am")),
			P(g.Text(fmt.Sprintf("Day 1 opens %s at 8:00am. This page is only on this Mac, at http://127.0.0.1:8787.", when))),
		),
	)
}

func finishedPage(total int) g.Node {
	return shell("Track complete", "Starter track complete", "All 21 lessons stay available above.", total, 0, total,
		Div(Class("panel"),
			H1(g.Text("You finished the starter track")),
			P(g.Text("Each day was one idea from Lean: definitions, then tactics, then induction. Open any day above to read the lesson and try the quiz again.")),
		),
	)
}

func lockedPage(n int, start time.Time, unlocked, total int) g.Node {
	when := dayDate(start, n).Format("Monday, January 2")
	return shell("Not yet", fmt.Sprintf("Day %d is locked", n), "", unlocked, 0, total,
		Div(Class("panel"),
			H1(g.Text(fmt.Sprintf("Day %d opens %s", n, when))),
			P(g.Text("Lessons unlock at 8:00am, one per day.")),
		),
	)
}
