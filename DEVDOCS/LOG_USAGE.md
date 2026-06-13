## Système de Logs (`pkg/utils/log.go`)

Le système de logs permet d'afficher des messages formatés et horodatés dans la console. Il gère différents niveaux de logs : information, erreur, débogage et erreur fatale.

### Fonctions disponibles

*   `utils.Log(info string)` : Affiche un message d'information standard.
*   `utils.LogError(info string, err any)` : Affiche un message d'erreur **en rouge** dans la console pour le repérer facilement.
*   `utils.LogFatal(info string, err any)` : Affiche un message d'erreur et **arrête immédiatement le programme**. À utiliser pour les erreurs critiques qui empêchent l'application de démarrer (ex: connexion à la BDD, chargement des templates).
*   `utils.Debug(info string)` : Affiche un message de débogage uniquement si le mode debug a été activé.
*   `utils.SetDebugMode(debug bool)` : Permet d'activer ou désactiver globalement les messages de type Debug.

### Exemples d'utilisation

```go
package main

import (
    "errors"
    "TonProjet/pkg/utils" // Pensez à ajuster l'import
)

func main() {
    utils.SetDebugMode(true)

    utils.Log("Le serveur démarre...")
    // Sortie: 2026-05-29 15:04:05 : Le serveur démarre...

    err := errors.New("connexion refusée")
    utils.LogError("Erreur BDD", err)
    // Sortie (en rouge): 2026-05-29 15:04:05 : Erreur BDD : connexion refusée

    // Exemple d'erreur fatale
    utils.LogFatal("Impossible de charger la configuration", err)
    // Sortie (en rouge): FATAL: 2026-05-29 15:04:05 : Impossible de charger la configuration : connexion refusée
    // (Le programme s'arrête ici)
}
```