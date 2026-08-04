package utils

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

func StripHTML(input string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(input))
	var buffer bytes.Buffer

	for {
		tokenType := tokenizer.Next()

		switch tokenType {
		case html.ErrorToken:
			return buffer.String()
		case html.TextToken:
			token := tokenizer.Token()
			buffer.WriteString(token.Data)
		}
	}
}