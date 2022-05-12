package helpers

import (
	"bytes"
	"fmt"
	"github.com/webview/webview"
	"html/template"
	"os"
	"regexp"
	"strings"
)

var regCss = regexp.MustCompile(`(?mi)<link rel="stylesheet" href="(.+\.css)">|<script src="(.+\.js)"></script>`)

type Page struct {
	Template *template.Template
}

func ParseHtml(pageName string, funcMap template.FuncMap) (*Page, error) {
	temp := template.New("index")
	if funcMap != nil {
		temp.Funcs(funcMap)
	}

	rawHtml, _ := os.ReadFile(fmt.Sprintf("./pages/%s/index.gohtml", pageName))
	html := string(rawHtml)
	for _, str := range regCss.FindAllStringSubmatch(html, -1) {
		if str[1] != "" {
			css, err := template.ParseFiles(fmt.Sprintf("./pages/%s/%s", pageName, str[1]))
			if err != nil {
				return nil, err
			}

			var rawCss bytes.Buffer
			err = css.Execute(&rawCss, []string{})
			if err != nil {
				return nil, err
			}

			html = strings.ReplaceAll(html, str[0], fmt.Sprintf("<style>%s</style>", rawCss.String()))
		}

		if str[2] != "" {
			js, err := template.ParseFiles(fmt.Sprintf("./pages/%s/%s", pageName, str[2]))
			if err != nil {
				return nil, err
			}

			var rawJs bytes.Buffer
			err = js.Execute(&rawJs, []string{})
			if err != nil {
				return nil, err
			}

			html = strings.ReplaceAll(html, str[0], fmt.Sprintf("<script>%s</script>", rawJs.String()))
		}
	}

	page, err := temp.Parse(html)
	if err != nil {
		return nil, err
	}

	return &Page{
		Template: page,
	}, nil
}

func (page Page) Execute(data any) (string, error) {
	var rawPage bytes.Buffer
	err := page.Template.Execute(&rawPage, data)
	return rawPage.String(), err
}

func (page Page) Navigate(w webview.WebView, data any) {
	htmlRaw, err := page.Execute(data)
	if err != nil {
		fmt.Println(err)
	}

	w.Dispatch(func() {
		w.Navigate(fmt.Sprint("data:text/html,", htmlRaw))
	})
}
