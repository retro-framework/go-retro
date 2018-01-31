package identifier

import (
	"bytes"
	"strings"
	"unicode"
)

func New(name string) *identifier {
	return NewIdentifier(name)
}

func NewIdentifier(name string) *identifier {
	return &identifier{
		name: name,
		splitFn: func(r rune) bool {
			return unicode.IsUpper(r) || unicode.IsPunct(r) || unicode.IsSpace(r)
		},
	}
}

type identifier struct {
	name    string
	splitFn func(rune) bool
}

func (i *identifier) TypeName() string {
	var b bytes.Buffer
	for _, w := range i.parts() {
		b.WriteString(strings.Title(w))
	}
	return b.String()
}

func (i *identifier) Path() string {
	var b bytes.Buffer
	for _, w := range i.parts() {
		b.WriteString(strings.ToLower(w))
	}
	return b.String()
}

func (i *identifier) Abbr() string {
	var b bytes.Buffer
	for _, w := range i.parts() {
		b.WriteByte(w[0])
	}
	return b.String()
}

func (i *identifier) CamelCase() string {
	var b bytes.Buffer
	for _, w := range i.parts() {
		b.WriteString(strings.Title(w))
	}
	return b.String()
}

func (i *identifier) parts() []string {
	var words []string
	l := 0
	for s := i.name; s != ""; s = s[l:] {
		l = strings.IndexFunc(s[1:], i.splitFn) + 1
		if l <= 0 {
			l = len(s)
		}
		word := strings.Trim(strings.ToLower(s[:l]), "_-./|!' ")
		if word != "" {
			words = append(words, word)
		}
	}
	return words
}
