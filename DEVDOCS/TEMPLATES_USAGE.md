## Système de Layout et Templates (Automatisé)

Le projet utilise un système de "Layout" (Gabarit) intelligent qui charge automatiquement tous les templates au démarrage.

Le principe reste le même : la structure globale (`<head>`, `<nav>`, `<footer>`) est dans `template/layout.html`. Les pages spécifiques (`index.html`, etc.) ne contiennent que leur propre contenu. La grande différence est que vous n'avez plus besoin de lister manuellement les fichiers à charger !

### Comment créer une nouvelle page ?

Imaginons que nous voulions créer une page "Contact".

#### Étape 1 : Créer le fichier HTML
Créez un fichier `template/contact.html`. Le contenu principal de votre page doit être enveloppé dans un bloc `{{define "page"}} ... {{end}}`. C'est ce bloc qui sera injecté dans le layout.

```html
<!-- template/contact.html -->
{{define "page"}}
<main>
    <h1>Nous contacter</h1>
    <p>Ceci est la page de contact.</p>
    
    <!-- Vous pouvez utiliser des variables Go -->
    <p>Message du serveur : {{.Message}}</p>

    <!-- Vous pouvez même inclure d'autres composants ! -->
    {{template "un-composant" .}}
</main>
{{end}}
```

#### Étape 2 : Créer le Handler côté Go
Dans votre dossier `internal/router/`, créez la fonction qui va générer la page. C'est maintenant beaucoup plus simple : un seul appel de fonction suffit.

```go
package router

import (
    "net/http"
    "TonProjet/pkg/utils" // Ajuster l'import
)

func ContactHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Préparer les données (optionnel)
    data := struct {
        Title   string
        Message string
    }{
        Title:   "Page de Contact",
        Message: "Bienvenue sur notre formulaire !",
    }

    // 2. Appeler le template par son nom de fichier
    utils.RenderTemplate(w, "contact.html", data)
}
```

### Comment créer un nouveau composant ?

C'est la plus grande force de ce système :
1. Créez un fichier dans `template/components/`, par exemple `mon-profil.html`.
2. Définissez votre composant avec un nom unique : `{{define "mon-profil"}} ... {{end}}`.
3. Utilisez-le où vous voulez avec `{{template "mon-profil" .}}`.

**Aucune modification du code Go n'est nécessaire !** Le composant sera détecté et chargé automatiquement au prochain démarrage.

### Fonctionnement Interne
- `utils.LoadTemplates()` : Appelée une seule fois dans `main.go`, cette fonction scanne les dossiers `template/` et `template/components/`. Pour chaque page, elle crée un "set" de templates qui inclut la page elle-même, le `layout.html`, et **tous** les composants trouvés. Ces sets sont mis en cache.
- `utils.RenderTemplate(w, name, data)` : Récupère le set de templates pré-compilé depuis le cache (via `name` qui est le nom du fichier) et l'exécute, ce qui est très rapide.
- Les anciennes fonctions `Render` et `RenderFile` n'existent plus.
