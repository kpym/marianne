package main

import (
	"io"
	"strings"
)

type trad struct {
	from, to string
}

var traductions = strings.NewReplacer(
	" (default [])", "",
	"default", "par défaut",
	" strings ", "         ",
	" string ", "        ",
	" uints ", "       ",
	" int ", "     ",
	"bad flag syntax:", "mauvaise syntaxe du paramètre :",
	"unknown flag:", "paramètre inconnu :",
	"unknown shorthand flag:", "paramètre court inconnu :",
	"flag needs an argument:", "paramètre sans argument :",
	" in ", " dans ",
)

// FrenchTranslator est un io.Writer qui traduit quelques phrases d'angalsi en français et
// renvoie le résultat au w
type FrenchTranslator struct {
	w io.Writer
}

// Write traduit le message p et renvoie la traduction à fw.w
func (fw FrenchTranslator) Write(p []byte) (n int, err error) {
	return traductions.WriteString(fw.w, string(p))
}
