# TP : GoLog Analyzer - Analyse de Logs Distribuée

loganalyzer est un outil en ligne de commande écrit en Go qui analyse en parallèle des fichiers de logs décrits dans un fichier de configuration JSON. Le programme simule le travail d'un analyseur (latence aléatoire), remonte les erreurs d'accès aux fichiers et peut exporter un rapport JSON filtrable.

## Sommaire
- [Fonctionnalités](#fonctionnalités)
- [Prérequis](#prérequis)
- [Installation](#installation)
- [Configuration d'entrée](#configuration-dentrée)
- [Commande `analyze`](#commande-analyze)
- [Rapport JSON](#rapport-json)
- [Architecture](#architecture)
- [Gestion des erreurs](#gestion-des-erreurs)
- [Aller plus loin](#aller-plus-loin)
- [Équipe](#équipe)

## Fonctionnalités
- Analyse concurrente des logs décrits dans un fichier JSON.
- Gestion robuste des erreurs (fichier manquant, JSON invalide) via des erreurs personnalisées et `errors.Is/As`.
- Export optionnel d'un rapport JSON avec création automatique des dossiers et horodatage du fichier (`YYMMDD_nom.json`).
- Filtrage des résultats par statut (`OK` ou `FAILED`) avant affichage/export.
- Interface CLI basée sur Cobra.

## Prérequis
- Go \>= 1.21 (développé avec Go 1.24).

## Installation

### Mode développement (rapide)
Exécuter directement le projet avec Go :
```bash
go run . analyze -c config.json
```

### Compilation d’un binaire local
Compiler le projet en binaire exécutable nommé `loganizer` :
```bash
go build -o loganizer .
./loganizer analyze -c config.json -o rapports/report.json --status OK
```

### Installation globale (optionnel)
Installer dans `$GOPATH/bin` ou `$HOME/go/bin` pour l’utiliser partout :
```bash
go install .
loganizer analyze -c config.json
```

---
## Configuration d'entrée
Le fichier JSON (par défaut `config.json`) contient un tableau d'objets :

```json
[
  {
    "id": "web-server-1",
    "path": "test_logs/access.log",
    "type": "nginx-access"
  },
  {
    "id": "app-backend-2",
    "path": "test_logs/errors.log",
    "type": "custom-app"
  }
]
```

Chaque entrée doit avoir un `id`, un `path` et un `type`. Les chemins peuvent être relatifs ou absolus.

## Commande `analyze`
```bash
./loganalyzer analyze --config config.json [--output rapports/report.json] [--status OK|FAILED]
```

Options :
- `--config, -c` : chemin du fichier JSON décrivant les logs (obligatoire).
- `--output, -o` : chemin du rapport JSON à générer. Le fichier créé sera préfixé par la date du jour (`YYMMDD_`).
- `--status` : filtre les résultats par statut (`OK` ou `FAILED`).

Pendant l'exécution, un résumé est affiché pour chaque log :
```
[OK] web-server-1 (test_logs/access.log) -> Analyse terminée avec succès.
[FAILED] missing-log (test_logs/missing.log) -> Fichier introuvable. | open test_logs/missing.log: no such file or directory
```

## Rapport JSON
Lorsque l'option `--output` est fournie, un fichier JSON est généré avec la structure suivante :
```json
[
  {
    "log_id": "web-server-1",
    "file_path": "test_logs/access.log",
    "status": "OK",
    "message": "Analyse terminée avec succès."
  },
  {
    "log_id": "missing-log",
    "file_path": "test_logs/missing.log",
    "status": "FAILED",
    "message": "Fichier introuvable.",
    "error_details": "open test_logs/missing.log: no such file or directory"
  }
]
```
Les dossiers de sortie sont créés automatiquement si nécessaire.

## Architecture
- `cmd/` : commandes Cobra (`root.go`, `analyze.go`).
- `internal/config` : chargement et validation des configurations JSON.
- `internal/analyzer` : exécution concurrente des analyses, erreurs personnalisées, filtrage.
- `internal/reporter` : export des rapports JSON.
- `main.go` : point d'entrée, gestion des signaux (Ctrl+C).

## Gestion des erreurs
- `config.ParseError` (`ErrConfigParse`) détecte les problèmes de parsing ou de contenu JSON.
- `analyzer.FileAccessError` (`ErrFileAccess`) encapsule les erreurs d'accès aux logs (`errors.As`).
- Les messages retournés distinguent les cas « fichier introuvable », « accès refusé » ou erreur générique.

## Aller plus loin
1. Ajouter des stratégies d'analyse réelles selon le type de log.
2. Introduire une sous-commande `add-log` pour enrichir dynamiquement le fichier de configuration.
3. Écrire des tests unitaires couvrant les erreurs personnalisées et le filtrage des résultats.

## Équipe
- Axelle Lanca — Développement & Conception CLI
- Grace Van — Tests & Documentation
- Logan Analyzer — Mascotte de l'équipe
