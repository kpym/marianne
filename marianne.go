package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/markbates/pkger"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/eps"
	"github.com/tdewolff/canvas/pdf"
	"github.com/tdewolff/canvas/rasterizer"
	"github.com/tdewolff/canvas/svg"
)

// quelques variables globales
var (
	// la version du logiciel (remplacée lors du build)
	version = "--"

	// la variable temporaire d'erreur
	err error
)

// les flags (pour la description voir dans la fonction main)
var (
	nom           string
	institution   string
	direction     string
	eol           string
	hauteurs      []int
	formats       []string
	avecMarges    bool
	sansMarges    bool
	pourSignature bool
	silence       bool
	aide          bool
)

// panique en cas d'erreur
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Imprimer des messages si pas quiet
var log = func(msg ...interface{}) {
	fmt.Fprint(os.Stderr, msg...)
}

// affiche un texte multilingue dans le context ctx
// - fontFamily : la police Marianne-Bold
// - txt : la texte à afficher
// - xPos,YPos : la position en bas à gauche de la première ligne du texte
// - size : la taille de la police (plus précisément la hauteur du "A")
// - step : la distance entre les lignes
// Retour : la position en bas à droite du "bounding box"
func drawText(ctx *canvas.Context, fontFamily *canvas.FontFamily, txt string, xPos, yPos, size, step float64) (float64, float64) {
	// la coordonnées x maximale (à retourner)
	var w float64
	// La lettre A fait 70% de la taille de la police
	// et la conversion mm -> pt est 72/25.4
	// donc la constante par laquelle on multiplie la taille de la police en pt pour obtenir la taille de A en mm est
	const fontScale = 72 / 25.4 * 100 / 70

	// préparation du texte
	txt = strings.ReplaceAll(txt, eol, "\n")
	ta := strings.Split(txt, "\n")

	// affichage du texte
	ctx.SetFillColor(canvas.Black)
	face := fontFamily.Face(size*fontScale, canvas.Black, canvas.FontBold, canvas.FontNormal)
	for i := 0; i < len(ta); i++ {
		line := strings.TrimSpace(ta[i])
		if len(line) == 0 {
			continue
		}
		// transformation du texte en chemin
		p, dx := face.ToPath(line)
		if dx > w {
			w = dx
		}
		r := p.Bounds()
		yPos += size
		ctx.DrawPath(xPos-r.X, -yPos, p)
		yPos += step
	}

	return yPos - step, w
}

// la fonction qui dessine le logo avec les textes (institution, direction)
func draw(ctx *canvas.Context, institution, direction string) {

	// déclaration de la police Marianne
	fontFamily := canvas.NewFontFamily("Marianne")
	fontFamily.Use(canvas.CommonLigatures)

	// chargement de la police Marianne-Bold
	f, err := pkger.Open("/fonts/Marianne-Bold.otf")
	check(err)
	fnt, err := ioutil.ReadAll(f)
	check(err)
	err = fontFamily.LoadFont(fnt, canvas.FontBold)
	check(err)

	// chemin temporaire
	var p *canvas.Path

	// affiche de la Marianne
	for i := 0; i < 3; i++ {
		p, _ := canvas.ParseSVG(logo[i])
		ctx.SetFillColor(logo_color[i])
		ctx.DrawPath(0, 0, p)
	}

	// affiche de l'institution
	dyI, dxI := drawText(ctx, fontFamily, strings.ToUpper(institution), 0, 3*x/2, 3*x/4, x/3)

	// affiche la devise
	p, _ = canvas.ParseSVG(devise)
	ctx.DrawPath(0, -dyI-x/2, p)

	// si la direction est présente
	if len(direction) > 0 {
		// détermine les espacement horizontaux entre le trait vertical et les textes
		dx1, dx2 := x, x
		if pourSignature {
			dx1, dx2 = 3*x, x/2
		}
		// affiche l'intitulé de la direction
		dyD, _ := drawText(ctx, fontFamily, direction, dxI+dx1+dx2, 3*x/2, 11*x/20, x/3)

		// affiche le trait séparateur
		pen := x / 40 // on suppose que 500 est proche de 12pt
		rx, ry, rW, rH := dxI+dx1-pen/2, 3*x/2, pen, math.Max(dyI, dyD)-3*x/2
		ctx.DrawPath(rx, -ry, canvas.Rectangle(rW, -rH))
	}
}

// le logo est mis sur fond blanc
func onWhite(c *canvas.Canvas) *canvas.Canvas {
	cn := canvas.New(c.W, c.H)
	ctx := canvas.NewContext(cn)
	ctx.SetFillColor(canvas.White)
	ctx.DrawPath(0, 0, canvas.Rectangle(math.Ceil(c.W), math.Ceil(c.H)))
	c.Render(cn)

	return cn
}

// les 16 couleurs du logo pour PNG et GIF (non utilisé pour le moment)
var MariannePalette16 = color.Palette{
	color.RGBA{0x0c, 0x0c, 0x0c, 0xff},
	color.RGBA{0xff, 0xff, 0xff, 0xff},
	color.RGBA{0x79, 0x79, 0x7c, 0xff},
	color.RGBA{0xa3, 0xa3, 0xa4, 0xff},
	color.RGBA{0x09, 0x09, 0x94, 0xff},
	color.RGBA{0xd7, 0xd7, 0xd9, 0xff},
	color.RGBA{0xe1, 0x04, 0x11, 0xff},
	color.RGBA{0xf2, 0xf2, 0xf2, 0xff},
	color.RGBA{0xf9, 0xcb, 0xce, 0xff},
	color.RGBA{0xc1, 0xc0, 0xc1, 0xff},
	color.RGBA{0x47, 0x47, 0x47, 0xff},
	color.RGBA{0xed, 0x5f, 0x70, 0xff},
	color.RGBA{0xfb, 0xdd, 0xdf, 0xff},
	color.RGBA{0xf0, 0xf0, 0xf9, 0xff},
	color.RGBA{0xe7, 0xe7, 0xe7, 0xff},
	color.RGBA{0xfa, 0xfa, 0xfa, 0xff},
}

// les 8 couleurs du logo pour PNG et GIF
var MariannePalette8 = color.Palette{
	color.NRGBA{0xff, 0xff, 0xff, 0xff}, // blanc
	color.NRGBA{0x05, 0x05, 0x05, 0xff}, // noir
	color.NRGBA{0x80, 0x80, 0x82, 0xff}, // gris
	color.NRGBA{0xb3, 0xb2, 0xb3, 0xff}, // gris
	color.NRGBA{0x00, 0x00, 0x91, 0xff}, // bleu
	color.NRGBA{0xe1, 0x00, 0x0f, 0xff}, // rouge
	color.NRGBA{0xdb, 0xdb, 0xdb, 0xff}, // gris
	color.NRGBA{0xea, 0x65, 0x67, 0xff}, // rouge pale
}

type MarianneQuantizer struct {
}

func (q *MarianneQuantizer) Quantize(p color.Palette, m image.Image) color.Palette {
	return MariannePalette8
}

// transforme les chemins du canevas en image 8 ou 16 couleurs
func CanvasToIndexedImg(c *canvas.Canvas, h int, p color.Palette) image.Image {
	img := image.NewPaletted(image.Rect(0, 0, int(c.W*float64(h)/c.H+0.5), h), p)
	c.Render(rasterizer.New(img, canvas.DPMM(float64(h)/c.H)))
	return img
}

// Créer les fichiers : svg, pdf, eps, png, gif, jpg
// - c : le canvas contenant l'image
// - zp : chaîne "sans zone de protection" a rajouter au nom ou pas
func writeImages(c *canvas.Canvas, zp, formats string) {
	var name string

	// Création du SVG
	if strings.Contains(formats, "svg") {
		name = fmt.Sprintf("%s%s.svg", nom, zp)
		// au lieu de fichier on utilise un Buffer
		var memoryFile = new(bytes.Buffer)
		err = svg.Writer(memoryFile, c)
		check(err)
		var buf = memoryFile.Bytes()
		// et on supprime 'width' et 'height'
		reWidth := regexp.MustCompile(`(?m)(width\s*=\s*\"[^"]*\"\s*)`)
		buf = reWidth.ReplaceAll(buf, []byte{})
		reHeight := regexp.MustCompile(`(?m)(height\s*=\s*\"[^"]*\"\s*)`)
		buf = reHeight.ReplaceAll(buf, []byte{})
		// et on enregistre finalement
		err = ioutil.WriteFile(name, buf, 0644)
		check(err)
		log("SVG fait.\n")
	}

	// Création du PDF
	if strings.Contains(formats, "pdf") {
		name := fmt.Sprintf("%s%s.pdf", nom, zp)
		c.WriteFile(name, pdf.Writer)
		log("PDF fait.\n")
	}

	// Création du EPS
	if strings.Contains(formats, "eps") {
		name := fmt.Sprintf("%s%s.eps", nom, zp)
		c.WriteFile(name, eps.Writer)
		log("EPS fait.\n")
	}

	doPNG := strings.Contains(formats, "png")
	doGIF := strings.Contains(formats, "gif")
	doJPG := strings.Contains(formats, "jpg") || strings.Contains(formats, "jpeg")

	if doPNG || doGIF || doJPG {
		// les hauteurs des images

		// pour chaque hauteur ...
		for i := 0; i < len(hauteurs); i++ {
			log("Image de hauteur ", hauteurs[i], ".")
			// la base du nom (sans l'extension)
			name := fmt.Sprintf("%s%s_%d.", nom, zp, hauteurs[i])

			// Création PNG et GIF (en 8 couleurs)
			if doPNG || doGIF {
				img := CanvasToIndexedImg(c, hauteurs[i], MariannePalette8)
				if doPNG {
					dstFile, err := os.Create(name + "png")
					check(err)
					defer dstFile.Close()
					png.Encode(dstFile, img)
					log("..png.")
				}
				if doGIF {
					dstFile, err := os.Create(name + "gif")
					check(err)
					defer dstFile.Close()
					gif.Encode(dstFile, img, nil)
					log("..gif.")
				}
			}

			// création JPG
			if doJPG {
				c.WriteFile(name+"jpg", rasterizer.JPGWriter(canvas.DPMM(float64(hauteurs[i])/c.H), nil))
				log("..jpg.")
			}

			log(" Fait.\n")
		}
	}

}

var traductions = map[string]string{
	"default": "par défaut",
	"strings": "       ",
	"string":  "      ",
	"ints":    "    ",
	// "bad flag syntax:":        "mauvaise syntaxe du paramètre :",
	// "unknown flag:":           "paramètre inconnu :",
	// "unknown shorthand flag:": "paramètre court inconnu :",
	// " in ":                    " dans ",
	// "flag needs an argument:": "Le paramètre nécessite d'un argument :",
}

// Affiche l'aide d'utilisation
// C'est un peut plus compliqué que ce que ça devrait être
// car on doit remplacer "default" avec "par défaut"
// voir : https://github.com/golang/go/issues/42124
func Aide() {

	fmt.Fprintf(os.Stderr, "marianne (version: %s)\n\n", version)
	fmt.Fprintf(os.Stderr, "Ce programme génère le logo de l'institution.\nParamètres disponibles:\n\n")

	// Remplace "default" avec "par défaut" dans `flag.PrintDefaults`
	var buf = new(bytes.Buffer)
	// on redirige la sortie vers buf et on affiche les flags dedans
	flag.CommandLine.SetOutput(buf)
	defer flag.CommandLine.SetOutput(os.Stderr)
	flag.PrintDefaults()
	// le message avant les ajustements
	msg := string(buf.Bytes())
	// traduction en français
	for from, to := range traductions {
		msg = strings.ReplaceAll(msg, from, to)
	}
	// on affiche finalement le message d'aide
	fmt.Fprintf(os.Stderr, msg)
	fmt.Fprintf(os.Stderr, "\n")
}

func main() {
	// déclare les flags
	flag.StringVarP(&nom, "nom-du-logo", "o", "logo", "Le nom du logo = le début des noms des fichiers générés.")
	flag.StringVarP(&institution, "institution", "i", "RÉPUBLIQUE\\FRANÇAISE", "Le nom du ministère, ambassade...")
	flag.StringVarP(&direction, "direction", "d", "", "Intitulé de direction, service ou délégation interministérielles.")
	flag.StringSliceVarP(&formats, "format", "f", []string{"SVG"}, "Le(s) format(s) parmi SVG, PDF, EPS, PNG, GIF et JPG.")
	flag.IntSliceVarP(&hauteurs, "hauteur", "t", []int{100, 300, 700}, "La (ou les) hauteur(s) pour les logos en PNG, GIF et JPG.")
	flag.BoolVarP(&avecMarges, "avec-marges", "M", false, "Avec zone de protection autour du logo. Ce paramètre est compatible avec -sans-marges.")
	flag.BoolVarP(&sansMarges, "sans-marges", "m", false, "Sans zone de protection autour du logo ('_szp' est rajouté aux noms des fichiers).")
	flag.BoolVarP(&pourSignature, "pour-signature", "g", false, "Le logo est destiné à une signature mail.")
	flag.StringVar(&eol, "eol", "\\", "Le passage à la ligne, en plus du EOL standard.")
	flag.BoolVarP(&silence, "silence", "q", false, "N'imprime rien.")
	flag.BoolVarP(&aide, "aide", "h", false, "Imprime ce message d'aide.")
	// Message d'aide
	flag.CommandLine.SortFlags = false
	flag.Usage = Aide
	// récupère les flags
	flag.Parse()
	// affiche l'aide si demandé
	if aide {
		flag.Usage()
		os.Exit(0)
	}
	// au moins une des versions doit être présente (avec marges par défaut)
	if !sansMarges {
		avecMarges = true
	}
	// normalisation des formats
	var formatstr = strings.ToLower(strings.Join(formats, ","))
	// silence ?
	if silence {
		log = func(msg ...interface{}) {}
	}

	// le canevas et le contexte sur lesquels on va dessiner
	c := canvas.New(1, 1) // la taille sera ajustée après avec Fit()
	ctx := canvas.NewContext(c)
	log("Création du logo ...")
	draw(ctx, institution, direction)
	log("fait.\n")

	if sansMarges {
		c.Fit(0.0)
		log("\nEnregistrement sans marges :\n")
		writeImages(onWhite(c), "_szp", formatstr)
	}
	if avecMarges {
		c.Fit(x)
		log("\nEnregistrement avec marges :\n")
		writeImages(onWhite(c), "", formatstr)
	}
}
