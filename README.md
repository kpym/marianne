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

  -o, --nom-du-logo          Le nom du logo = le début des noms des fichiers générés. (par défaut "logo")
  -i, --institution          Le nom du ministère, ambassade... (par défaut "RÉPUBLIQUE\\FRANÇAISE")
  -d, --direction            Intitulé de direction, service ou délégation interministérielles.
  -f, --format               Le(s) format(s) parmi SVG, PDF, EPS, PNG, GIF et JPG. (par défaut SVG, ou PNG pour signature)
  -t, --hauteur              La (ou les) hauteur(s) pour les logos en PNG, GIF et JPG. (par défaut 700, ou 100 pour signature)
  -M, --avec-marges          Avec zone de protection autour du logo. Ce paramètre est compatible avec -sans-marges.
  -m, --sans-marges          Sans zone de protection autour du logo ('_szp' est rajouté aux noms des fichiers).
  -g, --pour-signature       Le logo est destiné à une signature mail.
      --eol                  Le passage à la ligne, en plus du EOL standard. (par défaut "\\")
      --qualite-jpg          La qualité [1-100] des jpeg. (par défaut 100)
      --seize-couleurs       Enregistre les PNG et les GIF en 16 couleurs, sinon c'est en 8.
  -q, --silence              N'imprime rien.
  -h, --aide                 Imprime ce message d'aide.
```

### Exemple

```shell
$ ./marianne -i "L'institution" -d "Intitulé de la\\direction" -f svg -f png -t 100,300,700 -o "logo_inst"
Création du logo ...fait.

Enregistrement avec marges :
SVG fait.
Image de hauteur 100...png. Fait.
Image de hauteur 300...png. Fait.
Image de hauteur 700...png. Fait.

$ ls *.svg *.png
logo_inst.svg logo_inst_100.png logo_inst_300.png logo_inst_700.png
```
