package server

import (
	htmlTemplate "html/template"
	"net/http"
	"net/http/httptest"
	textTemplate "text/template"

	. "gopkg.in/check.v1"
)

var _ = Suite(&suiteNegotiation{})

type suiteNegotiation struct {
	html *htmlTemplate.Template
	text *textTemplate.Template
}

func acceptHeader(v string) http.Header {
	hdr := http.Header{}
	if v != "" {
		hdr.Set("Accept", v)
	}
	return hdr
}

func (s *suiteNegotiation) TestPrefersMarkdown(c *C) {
	cases := map[string]bool{
		"":           false,
		"*/*":        false,
		"text/*":     false,
		"text/plain": false,
		"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8": false,
		"text/markdown":                              true,
		"TEXT/Markdown":                              true,
		"text/markdown, text/html, */*":              true,
		"text/markdown, text/plain;q=0.9, */*;q=0.5": true,
		"text/html, text/markdown":                   false,
		"text/html;q=0.5, text/markdown":             true,
		"text/html, text/markdown;q=0.5":             false,
		"text/markdown;q=0":                          false,
		"text/markdown, text/html;q=0":               true,
	}

	for accept, want := range cases {
		c.Check(prefersMarkdown(acceptHeader(accept)), Equals, want, Commentf("Accept: %q", accept))
	}
}

func (s *suiteNegotiation) SetUpTest(c *C) {
	s.html, s.text = htmlTemplates, textTemplates
	htmlTemplates = initHTMLTemplates()
	textTemplates = initTextTemplates()
	_, err := htmlTemplates.New("index.html").Parse("html {{.WebAddress}}")
	c.Assert(err, IsNil)
	_, err = textTemplates.New("index.txt").Parse("text {{.WebAddress}}")
	c.Assert(err, IsNil)
}

func (s *suiteNegotiation) TearDownTest(c *C) {
	htmlTemplates, textTemplates = s.html, s.text
}

func view(accept string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "http://transfer.test/", nil)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	w := httptest.NewRecorder()
	(&Server{}).viewHandler(w, req)
	return w
}

func (s *suiteNegotiation) TestViewWithoutMarkdownTemplate(c *C) {
	c.Check(view("text/html,*/*;q=0.8").Body.String(), Equals, "html http://transfer.test/")
	c.Check(view("").Body.String(), Equals, "text http://transfer.test/")
	// no index.md: markdown clients fall back to the text page, not HTML
	c.Check(view("text/markdown, text/html, */*").Body.String(), Equals, "text http://transfer.test/")
}

func (s *suiteNegotiation) TestViewWithMarkdownTemplate(c *C) {
	_, err := textTemplates.New("index.md").Parse("# markdown {{.WebAddress}}")
	c.Assert(err, IsNil)

	w := view("text/markdown, text/html, */*")
	c.Check(w.Body.String(), Equals, "# markdown http://transfer.test/")
	c.Check(w.Header().Get("Content-Type"), Equals, "text/markdown; charset=utf-8")
	c.Check(w.Header().Get("Vary"), Equals, "Accept")

	c.Check(view("text/html,application/xhtml+xml,*/*;q=0.8").Body.String(), Equals, "html http://transfer.test/")
	c.Check(view("text/html, text/markdown").Body.String(), Equals, "html http://transfer.test/")
	c.Check(view("*/*").Body.String(), Equals, "text http://transfer.test/")
}

func (s *suiteNegotiation) TestPreviewMatcher(c *C) {
	srv := &Server{}
	preview := func(accept, referer string) bool {
		req := httptest.NewRequest("GET", "http://transfer.test/abc/file.txt", nil)
		req.Header.Set("Accept", accept)
		if referer != "" {
			req.Header.Set("Referer", referer)
		}
		return srv.previewMatcher(req, nil)
	}

	c.Check(preview("text/html,*/*;q=0.8", ""), Equals, true)
	c.Check(preview("text/html,*/*;q=0.8", "http://other.test/"), Equals, true)
	c.Check(preview("text/html,*/*;q=0.8", "http://transfer.test/abc/file.txt"), Equals, false)
	c.Check(preview("*/*", ""), Equals, false)
	c.Check(preview("text/markdown, text/html, */*", ""), Equals, false)
}
