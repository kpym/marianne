package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

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
	nom           *string
	institution   *string
	direction     *string
	eol           *string
	hauteurs      *string
	formats       *string
	avecMarges    *bool
	sansMarges    *bool
	pourSignature *bool
	silence       *bool
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

// conversion des hauteurs en liste d'entiers
func stoi(h string) []int {
	var n int

	r := make([]int, 0, 7)
	ah := strings.Split(h, ",")
	for i := 0; i < len(ah); i++ {
		n, err = strconv.Atoi(ah[i])
		check(err)
		if n > 0 {
			r = append(r, n)
		}
	}

	return r
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
	// constante expérimentale pour voir la bonne taille des lettres
	const fontScale = 4.050632911392405

	// préparation du texte
	txt = strings.ReplaceAll(txt, *eol, "\n")
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
		if *pourSignature {
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

func main() {
	// déclare les flags
	nom = flag.String("nom", "logo", "Le nom du fichier.")
	institution = flag.String("institution", "RÉPUBLIQUE\\FRANÇAISE", "Le nom du ministère, ambassade, ...")
	direction = flag.String("direction", "", "Intitulé de de direction, service ou délégations interministérielles.")
	eol = flag.String("eol", "\\", "Le passage à la ligne, en plus du EOL standard.")
	hauteurs = flag.String("hauteurs", "100,300,700", "Les hauteurs pour les logos en PNG et JPG.")
	formats = flag.String("formats", "SVG", "Les format parmi SVG, PDF, PNG, GIF et JPG.")
	avecMarges = flag.Bool("avec-marges", false, "Avec zone de protection autour du logo. Ce paramètre est compatible avec -sans-marges.")
	sansMarges = flag.Bool("sans-marges", false, "Sans zone de protection autour du logo. Ce paramètre est compatible avec -avec-marges.")
	pourSignature = flag.Bool("pour-signature", false, "Le logo est destiné à une signature mail.")
	silence = flag.Bool("silence", false, "N'imprime rien.")
	// Message d'aide
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "marianne (version: %s)\n\n", version)
		fmt.Fprintf(os.Stderr, "Ce programme génère le logo de l'institution.\nParamètres disponibles:\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}
	// récupère les flags
	flag.Parse()
	// au moins une des versions doit être présente (avec marges par défaut)
	if !*sansMarges {
		*avecMarges = true
	}
	// silence ?
	if *silence {
		log = func(msg ...interface{}) {}
	}

	// le canevas et le contexte sur lesquels on va dessiner
	c := canvas.New(1, 1) // la taille sera ajustée après avec Fit()
	ctx := canvas.NewContext(c)
	log("Création du logo ...")
	draw(ctx, *institution, *direction)
	log("fait.\n")

	if *sansMarges {
		c.Fit(0.0)
		log("\nEnregistrement sans marges :\n")
		writeImages(onWhite(c), "_szp", strings.ToLower(*formats))
	}
	if *avecMarges {
		c.Fit(x)
		log("\nEnregistrement avec marges :\n")
		writeImages(onWhite(c), "", strings.ToLower(*formats))
	}
}

func onWhite(c *canvas.Canvas) *canvas.Canvas {
	cn := canvas.New(c.W, c.H)
	ctx := canvas.NewContext(cn)
	ctx.SetFillColor(canvas.White)
	ctx.DrawPath(0, 0, canvas.Rectangle(math.Ceil(c.W), math.Ceil(c.H)))
	c.Render(cn)

	return cn
}

// Créer les fichiers : svg, pdf, png, ...
// - c : le canvas contenant l'image
// - zp : chaîne "sans zone de protection" a rajouter au nom ou pas
func writeImages(c *canvas.Canvas, zp, formats string) {
	var name string

	// Création du SVG
	if strings.Contains(formats, "svg") {
		name = fmt.Sprintf("%s%s.svg", *nom, zp)
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
		name := fmt.Sprintf("%s%s.pdf", *nom, zp)
		c.WriteFile(name, pdf.Writer)
		log("PDF fait.\n")
	}

	// Création du EPS
	if strings.Contains(formats, "eps") {
		name := fmt.Sprintf("%s%s.eps", *nom, zp)
		c.WriteFile(name, eps.Writer)
		log("EPS fait.\n")
	}

	heights := stoi(*hauteurs)
	// création du PNG
	if strings.Contains(formats, "png") {
		log("PNG de hauteur ...")
		for i := 0; i < len(heights); i++ {
			name := fmt.Sprintf("%s%s_%d.png", *nom, zp, heights[i])
			c.WriteFile(name, rasterizer.PNGWriter(canvas.DPMM(float64(heights[i])/c.H)))
			log(heights[i], "...")
		}
		log("fait.\n")
	}

	// création du JPG
	if strings.Contains(formats, "jpg") || strings.Contains(formats, "jpeg") {
		log("JPG de hauteur ...")
		for i := 0; i < len(heights); i++ {
			name := fmt.Sprintf("%s%s_%d.jpg", *nom, zp, heights[i])
			c.WriteFile(name, rasterizer.JPGWriter(canvas.DPMM(float64(heights[i])/c.H), nil))
			log(heights[i], "...")
		}
		log("fait.\n")
	}

	// création du JPG
	if strings.Contains(formats, "gif") {
		log("GIF de hauteur ...")
		for i := 0; i < len(heights); i++ {
			name := fmt.Sprintf("%s%s_%d.gif", *nom, zp, heights[i])
			c.WriteFile(name, rasterizer.GIFWriter(canvas.DPMM(float64(heights[i])/c.H), nil))
			log(heights[i], "...")
		}
		log("fait.\n")
	}
}
