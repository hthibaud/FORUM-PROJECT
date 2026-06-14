# Documentation d'utilisation de Openenv

## Rôle

Le package `openenv` est un utilitaire crucial pour la gestion des variables d'environnement au sein de l'application. Son rôle principal est de charger des configurations à partir d'un fichier `.env` situé à la racine du projet, et de les rendre accessibles de manière structurée et sécurisée dans toute l'application.

## Initialisation

Pour que les variables d'environnement soient disponibles, il est impératif d'initialiser le package `openenv` au démarrage de l'application. Cela se fait en appelant la fonction `Init()` :

```go
import "Forum/pkg/openenv"

func main() {
    openenv.Init()
    // ... reste de l'initialisation de l'application
}
```

La fonction `Init()` se charge de lire le fichier `.env`, de parser les variables et de les charger dans la structure `ENV`.

## Structure des données

Les variables d'environnement sont stockées dans une structure `envData` qui est instanciée en une variable globale et exportée `ENV`. La structure est définie dans `pkg/openenv/struct.go`.

Exemple de structure :
```go
type envData struct {
    SecretSessionKey string `env:"SECRET_SESSION_KEY"`
    Port             int    `env:"PORT"`
    DebugMode        bool   `env:"DEBUG_MODE"`
}
```

Chaque champ de la structure doit être accompagné d'un tag `env` qui correspond à la clé dans le fichier `.env`.

## Fichier `.env`

Le fichier `.env` doit être placé à la racine du projet. Sa syntaxe est simple, sous forme de paires `clé=valeur`.

Exemple de fichier `.env` :
```env
# Clé secrète pour la session
SECRET_SESSION_KEY="une_cle_secrete_tres_complexe"

# Port du serveur
PORT=8080

# Mode de débogage
DEBUG_MODE=true
```

Les commentaires (lignes commençant par `#`) et les lignes vides sont ignorés. Les guillemets (simples ou doubles) autour des valeurs sont automatiquement retirés.

## Utilisation

Une fois le module initialisé, les variables d'environnement sont directement accessibles via la variable globale `openenv.ENV`.

Exemple d'accès aux variables :
```go
import "Forum/pkg/openenv"

// ...

// Utilisation de la clé de session
key := openenv.ENV.SecretSessionKey

// Utilisation du port
port := openenv.ENV.Port
```

## Ajouter une nouvelle variable d'environnement

Pour ajouter une nouvelle variable d'environnement, suivez ces étapes :
1.  **Ajoutez la nouvelle clé et sa valeur** dans le fichier `.env`.
2.  **Modifiez la structure `envData`** dans `pkg/openenv/struct.go` en ajoutant le nouveau champ, son type et le tag `env` correspondant.

Le package `openenv` supporte nativement les types `string`, `int`, et `bool`. Assurez-vous de choisir le bon type pour votre variable.
