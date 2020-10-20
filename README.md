# Marianne (le logo de la république)

Ce petit logiciel génère le logo des administrations en respectant la charte graphique de l'état de 2020.

![Exemple de logo](exemple/logo_inst.svg)
## Installation

### Version compilée

Il suffit de télécharger la version pour votre os à partir des [Realases](https://github.com/kpym/marianne/releases).

### Compilation

Après avoir cloné ce dépôt, pour compiler avec Go il faut exécuter dans le répertoire du projet :

```shell
$ go build .
```
Vous pouvez aussi utiliser [goreleaser](https://github.com/goreleaser/goreleaser/) en local pour compiler pour tous les os :

```shell
goreleaser --snapshot --skip-publish --rm-dist
```

## Utilisation

### Aide

```shell
$ ./marianne -h
marianne (version: --)

Ce programme génère le logo de l'institution.
Paramètres disponibles:

  -avec-marges
        Avec zone de protection autour du logo. Ce paramètre est compatible avec -sans-marges.
  -direction string
        Intitulé de de direction, service ou délégations interministérielles.
  -eol string
        Le passage à la ligne, en plus du EOL standard. (default "\\")
  -formats string
        Les formats parmi SVG, PDF, PNG, GIF et JPG. (default "SVG")
  -hauteurs string
        Les hauteurs pour les logos en PNG et JPG. (default "100,300,700")
  -institution string
        Le nom du ministère, ambassade, ... (default "RÉPUBLIQUE\\FRANÇAISE")
  -nom string
        Le nom du fichier. (default "logo")
  -pour-signature
        Le logo est destiné à une signature mail.
  -sans-marges
        Sans zone de protection autour du logo. Ce paramètre est compatible avec -avec-marges.
  -silence
        N'imprime rien.
```

### Exemple

```shell
$ ./marianne -institution "L'institution" -direction "Intitulé de la\\direction" -formats "svg png" -nom "logo_inst"
Création du logo ...fait.

Enregistrement avec marges :
SVG fait.
PNG de hauteur ...100...300...700...fait.

$ ls *.svg *.png
logo_inst.svg logo_inst_100.png logo_inst_300.png logo_inst_700.png

```
