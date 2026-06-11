# 📚 Guide de Développement

Ce document explique comment utiliser le système de logs interne et comment créer de nouvelles pages en utilisant le système de layout (templates HTML) du projet.

---

## 🛠️ 1. Système de Logs (`pkg/utils/log.go`)

Le système de logs permet d'afficher des messages formatés et horodatés dans la console. Il gère différents niveaux de logs : information, erreur et débogage.

### Fonctions disponibles

*   `utils.Log(info string)` : Affiche un message d'information standard.
*   `utils.LogError(info string, err any)` : Affiche un message d'erreur **en rouge** dans la console pour le repérer facilement. Prend un message de contexte et l'erreur en question.
*   `utils.Debug(info string)` : Affiche un message de débogage uniquement si le mode debug a été activé.
*   `utils.SetDebugMode(debug bool)` : Permet d'activer ou désactiver globalement les messages de type Debug (à appeler au démarrage dans `main.go`).

### Exemples d'utilisation

```go
package main

import (
    "errors"
    "TonProjet/pkg/utils" // Pensez à ajuster l'import selon le nom de module
)

func main() {
    // Activer le mode debug (optionnel)
    utils.SetDebugMode(true)

    // Log classique
    utils.Log("Le serveur a démarré sur le port 8080")
    // Sortie: 2026-05-29 15:04:05 : Le serveur a démarré sur le port 8080

    // Log de debug
    utils.Debug("Ceci est un test de variable")
    // Sortie: (DEBUG) 2026-05-29 15:04:05 : Ceci est un test de variable

    // Log d'erreur
    err := errors.New("connexion refusée")
    utils.LogError("Erreur lors de la connexion à la BDD", err)
    // Sortie (en rouge): 2026-05-29 15:04:05 : Erreur lors de la connexion à la BDD : connexion refusée
}
```

---

## 🎨 2. Système de Layout et Templates (`pkg/utils/layout.go`)

Le projet utilise un système de "Layout" (Gabarit). 
Cela signifie que la structure globale du site (le `<head>`, la `<nav>`, le `<footer>`...) est définie à un seul endroit : dans le fichier `template/layout.html`. 

Les autres pages HTML (comme `index.html`) ne doivent contenir **que le contenu spécifique** à la page (généralement enveloppé dans une balise `<main>`).

### Comment créer une nouvelle page ? (Guide pas à pas)

Imaginons que nous voulions créer une page "Contact".

#### Étape 1 : Créer le fichier HTML
Crée un fichier `template/contact.html`. Ne mettez **pas** de balises `<html>`, `<head>` ou `<body>`. Mettez uniquement le contenu de votre page.

```html
<!-- template/contact.html -->
<main>
    <h1>Nous contacter</h1>
    <p>Ceci est la page de contact.</p>
    
    <!-- Vous pouvez utiliser des variables Go si besoin -->
    <p>Message du serveur : {{.Message}}</p>
</main>
```

#### Étape 2 : Créer le Handler côté Go
Dans votre dossier `internal/router/` (ou l'endroit où vous gérez vos routes), créez la fonction qui va générer la page.

Le rendu d'une page se fait en **deux étapes** :
1. Rendre le contenu spécifique (`contact.html`) en texte HTML avec `utils.Render()`.
2. Injecter ce contenu dans le layout global (`layout.html`) et l'envoyer au client avec `utils.RenderFile()`.

```go
package router

import (
    "net/http"
    "TonProjet/pkg/utils" // Ajuster l'import
)

func ContactHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Préparer les données dynamiques spécifiques à la page (optionnel)
    data := struct {
        Message string
    }{
        Message: "Bienvenue sur notre formulaire !",
    }

    // 2. Transformer le fichier `contact.html` en une chaîne de caractères HTML
    // Le premier paramètre est le nom du fichier SANS l'extension .html
    contentHTML := utils.Render("contact", data)

    // 3. Envoyer la page complète au client avec le Layout
    // Paramètres : Titre de la page, Contenu HTML généré à l'étape 2, ResponseWriter
    utils.RenderFile("Page de Contact", contentHTML, w)
}
```

### Fonctionnement Interne (pour les curieux)
- `utils.Render(fileName, data)` : Parse un template précis, exécute les variables (`{{.Data}}`), et retourne le tout sous forme de `string`. En cas d'erreur, il log l'erreur via `LogError` et retourne un message d'erreur.
- `utils.RenderFile(title, content, w)` : Instancie une structure `PageData` contenant votre Titre et votre Contenu HTML. Il appelle ensuite le fichier `layout.html` et écrit la réponse directement dans le navigateur du client.
