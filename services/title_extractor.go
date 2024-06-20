package services

import (
	"net/http"

	"golang.org/x/net/html"
)

// ExtractTitle extracts the title of a webpage given its URL.
func ExtractTitle(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	return findTitleTag(z)
}

func findTitleTag(z *html.Tokenizer) string {
	inHeader := false
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return ""

		case html.StartTagToken:
			t := z.Token()

			if t.Data == "head" {
				inHeader = true
				continue
			}

			if inHeader && t.Data == "title" {
				//drain any nested tags
				for titleType := z.Next(); titleType != html.TextToken; titleType = z.Next() {
				}
				return string(z.Text())
			}

		case html.EndTagToken:
			t := z.Token()
			if t.Data == "head" {
				return ""
			}
		}
	}
}
