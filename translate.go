package main

import (
	"fmt"
	"io"
	"strings"
)

type trad struct {
	from, to string
}

var traductions = []trad{
	{"default", "par défaut"},
	{"strings", "       "},
	{"string", "      "},
	{"ints", "    "},
	{"bad flag syntax:", "mauvaise syntaxe du paramètre :"},
	{"unknown flag:", "paramètre inconnu :"},
	{"unknown shorthand flag:", "paramètre court inconnu :"},
	{"flag needs an argument:", "Le paramètre nécessite d'un argument :"},
	{" in ", " dans "},
}

type FrenchTranslator struct {
	w io.Writer
}

func (fw FrenchTranslator) Write(p []byte) (n int, err error) {
	out := string(p)

	for _, tr := range traductions {
		out = strings.ReplaceAll(out, tr.from, tr.to)
	}

	fmt.Fprintf(fw.w, out)
	return len(p), nil
}
