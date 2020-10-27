package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"regexp" // pour les ajustements dans le svg
	"strings"

	flag "github.com/spf13/pflag" // pour les paramètres en ligne de commande

	"github.com/markbates/pkger" // permet d'inclure la police Marianne dans l'exécutable
	"github.com/nfnt/resize"     // pour pouvoir dessiner puis rétrécir le logo (pour les petites tailles)
	"github.com/tdewolff/canvas" // la bibliothèque principale pour réaliser le logo
	"github.com/tdewolff/canvas/eps"
	"github.com/tdewolff/canvas/pdf"
	"github.com/tdewolff/canvas/rasterizer"
	"github.com/tdewolff/canvas/svg"
	"github.com/tdewolff/minify/v2"
	minsvg "github.com/tdewolff/minify/v2/svg"
)

// quelques variables globales
var (
	// la version du logiciel (remplacée lors de la compilation)
	version = "--"

	// une variable temporaire d'erreur
	err error
)

// panique en cas d'erreur
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Imprimer des messages (si pas en mode silence)
var log = func(msg ...interface{}) {
	fmt.Fprint(os.Stderr, msg...)
}

// Aide affiche l'aide d'utilisation
func Aide() {
	var out = flag.CommandLine.Output()
	fmt.Fprintf(out, "marianne (version: %s)\n\n", version)
	fmt.Fprintf(out, "Ce programme génère le logo de l'institution.\nParamètres disponibles:\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(out, "\n")
}

// les flags (pour la description voir SetParameters plus bas)
var (
	nom           string
	institution   string
	direction     string
	formats       []string
	hauteurs      []uint
	avecMarges    bool
	sansMarges    bool
	pourSignature bool
	eol           string
	jpgq          int
	col16         bool
	silence       bool
	aide          bool
)

// SetParameters récupération des paramètre à partir de la ligne de commande, puis
// retourne `formatstr` qui contient la liste des format sous la forme "svg,png..."
func SetParameters() (formatstr string) {
	// déclare les flags (c.-à-d. les paramètres de la ligne de commande)
	flag.StringVarP(&nom, "nom-du-logo", "o", "logo", "Le nom du logo = le début des noms des fichiers générés.")
	flag.StringVarP(&institution, "institution", "i", "RÉPUBLIQUE\\FRANÇAISE", "Le nom du ministère, ambassade...")
	flag.StringVarP(&direction, "direction", "d", "", "Intitulé de direction, service ou délégation interministérielles.")
	flag.StringSliceVarP(&formats, "format", "f", nil, "Le(s) format(s) parmi SVG, PDF, EPS, PNG, GIF et JPG. (par défaut SVG, ou PNG pour signature)")
	flag.UintSliceVarP(&hauteurs, "hauteur", "t", nil, "La (ou les) hauteur(s) pour les logos en PNG, GIF et JPG. (par défaut 700, ou 100 pour signature)")
	flag.BoolVarP(&avecMarges, "avec-marges", "M", false, "Avec zone de protection autour du logo. Ce paramètre est compatible avec -sans-marges.")
	flag.BoolVarP(&sansMarges, "sans-marges", "m", false, "Sans zone de protection autour du logo ('_szp' est rajouté aux noms des fichiers).")
	flag.BoolVarP(&pourSignature, "pour-signature", "g", false, "Le logo est destiné à une signature mail.")
	flag.StringVar(&eol, "eol", "\\", "Le passage à la ligne, en plus du EOL standard.")
	flag.IntVar(&jpgq, "qualite-jpg", 100, "La qualité [1-100] des jpeg.")
	flag.BoolVar(&col16, "seize-couleurs", false, "Enregistre les PNG et les GIF en 16 couleurs, sinon c'est en 8.")
	flag.BoolVarP(&silence, "silence", "q", false, "N'imprime rien.")
	flag.BoolVarP(&aide, "aide", "h", false, "Imprime ce message d'aide.")
	// garde l'ordre des paramètres dans l'aide
	flag.CommandLine.SortFlags = false
	// installe la traduction des messages en français
	flag.CommandLine.SetOutput(FrenchTranslator{flag.CommandLine.Output()})
	// le message d'aide
	flag.Usage = Aide
	// en cas d'erreur ne pas afficher l'erreur une deuxième fois
	flag.CommandLine.Init("marianne", flag.ContinueOnError)

	// récupère les flags
	err = flag.CommandLine.Parse(os.Args[1:])
	// affiche l'aide si demandé ou si erreur de paramètre
	if aide || err != nil {
		flag.Usage()
		if err != nil {
			fmt.Fprintln(flag.CommandLine.Output(), "ERREUR : ", err)
			os.Exit(2)
		} else {
			os.Exit(0)
		}
	}

	// au moins une des versions doit être présente (avec marges par défaut)
	if !sansMarges && !avecMarges {
		if pourSignature {
			sansMarges = true
		} else {
			avecMarges = true
		}
	}

	// le format par défaut
	if formats == nil {
		if pourSignature {
			formats = []string{"PNG"}
		} else {
			formats = []string{"SVG"}
		}
	}

	// la hauteur par défaut
	if hauteurs == nil {
		if pourSignature {
			hauteurs = []uint{100}
		} else {
			hauteurs = []uint{700}
		}
	}

	// normalisation des formats
	formatstr = strings.ToLower(strings.Join(formats, ","))
	// silence ?
	if silence {
		log = func(msg ...interface{}) {}
	}

	if jpgq < 1 {
		jpgq = 1
	} else if jpgq > 100 {
		jpgq = 100
	}

	return // formatstr
}

// affiche un texte multilingue dans le context ctx
// - fontFamily : la police Marianne-Bold
// - txt : le texte à afficher
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
func drawLogo(ctx *canvas.Context, institution, direction string) {

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

	// affiche la Marianne
	for i := 0; i < 3; i++ {
		p, _ := canvas.ParseSVG(logo[i])
		ctx.SetFillColor(logoColor[i])
		ctx.DrawPath(0, 0, p)
	}

	// affiche l'institution
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

// CanvasToRGBAImg transforme les chemins du canevas en image RGB
func CanvasToRGBAImg(c *canvas.Canvas, oh uint) image.Image {
	h := int(oh)
	rescale := h < 700
	if rescale {
		h = int(2 * oh)
	}
	dpmm := float64(h) / c.H
	w := int(c.W*dpmm + 0.5)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	c.Render(rasterizer.New(img, canvas.DPMM(dpmm)))
	if rescale {
		return resize.Resize(0, oh, img, resize.Lanczos3)
	}
	return img
}

// MariannePalette16 contient les 16 couleurs du logo pour PNG et GIF
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

// MariannePalette8 contient les 8 couleurs du logo pour PNG et GIF
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

// ToIndexedImg transforme une image RGBA en image de 8 ou 16 couleurs
func ToIndexedImg(rgba image.Image) (img image.Image) {
	rect := image.Rect(0, 0, rgba.Bounds().Dx(), rgba.Bounds().Dy())
	logoPalette := MariannePalette8
	if col16 {
		logoPalette = MariannePalette16
	}
	img = image.NewPaletted(rect, logoPalette)
	dimg, _ := img.(draw.Image)
	draw.Draw(dimg, rect, rgba, image.ZP, draw.Src)

	return img
}

// SaveRasterImage enregistre l'image en fonction de l'extension
func SaveRasterImage(img image.Image, name, ext string) {
	dstFile, err := os.Create(name + ext)
	check(err)
	defer dstFile.Close()
	switch ext {
	case "png":
		png.Encode(dstFile, img)
	case "gif":
		gif.Encode(dstFile, img, nil)
	case "jpg":
		jpeg.Encode(dstFile, img, &jpeg.Options{Quality: jpgq})
	}
	log(".." + ext + ".")
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
		// suppression de la partie fractale
		reNumbers := regexp.MustCompile(`(?m)([ MZLHVCSQTA+-]\d*)\.\d+`)
		buf = reNumbers.ReplaceAll(buf, []byte("$1"))
		// compression du SVG (réécriture en coordonnées relatives)
		memoryFile.Reset()
		mediatype := "image/svg+xml"
		m := minify.New()
		m.AddFunc(mediatype, minsvg.Minify)
		check(m.Minify(mediatype, memoryFile, bytes.NewReader(buf)))
		// et on enregistre finalement
		err = ioutil.WriteFile(name, memoryFile.Bytes(), 0644)
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
		// pour chaque hauteur ...
		for i := 0; i < len(hauteurs); i++ {
			log("Image de hauteur ", hauteurs[i], ".")
			// la base du nom (sans l'extension)
			name := fmt.Sprintf("%s%s_%d.", nom, zp, hauteurs[i])
			// l'image matriciel non compressé
			img := CanvasToRGBAImg(c, hauteurs[i])
			// création du JPG
			if doJPG {
				SaveRasterImage(img, name, "jpg")
			}
			// Création des PNG et GIF (en 8 couleurs)
			if doPNG || doGIF {
				img = ToIndexedImg(img)
				if doPNG {
					SaveRasterImage(img, name, "png")
				}
				if doGIF {
					SaveRasterImage(img, name, "gif")
				}
			}
			log(" Fait.\n")
		}
	}

}

func main() {
	// récpère les paramètres de l'application
	var formatstr = SetParameters()

	// le canevas et le contexte sur lesquels on va dessiner
	c := canvas.New(1, 1) // la taille sera ajustée après avec Fit()
	ctx := canvas.NewContext(c)
	log("Création du logo ...")
	drawLogo(ctx, institution, direction)
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
