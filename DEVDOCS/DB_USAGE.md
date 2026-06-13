# 📖 Guide d'Utilisation de la Base de Données

Ce document explique comment interagir avec la base de données du forum en utilisant les fonctions fournies par le package `internal/db`.

**Règle d'or :** Ne jamais écrire de requêtes SQL directement dans les handlers. Toutes les interactions avec la base de données doivent passer par les fonctions de ce package.

---

## 🚀 1. Initialisation

La connexion à la base de données est gérée automatiquement. Le "pool" de connexions est initialisé au démarrage de l'application dans `cmd/main.go` et fermé proprement à l'arrêt du serveur.

```go
// cmd/main.go
func main() {
    // ...
    db.Init()
    defer db.Close() // La connexion est fermée à la fin du programme
    // ...
}
```

---

## 👤 2. Fonctions Utilisateurs (`User`)

Ces fonctions gèrent la création et la récupération des utilisateurs.

### `db.CreateUser(username, password, email string) error`

Crée un nouvel utilisateur dans la base de données. Le mot de passe fourni est **automatiquement haché** avant d'être stocké.

**Exemple :**
```go
err := db.CreateUser("JohnDoe", "son_mot_de_passe_123", "john.doe@example.com")
if err != nil {
    // Gérer l'erreur (par exemple, l'utilisateur existe déjà)
    utils.LogError("Impossible de créer l'utilisateur", err)
}
```

### `db.GetUserByUsername(username string) (*User, error)`

Récupère un utilisateur par son nom d'utilisateur. Retourne `nil, nil` si aucun utilisateur n'est trouvé.

**Exemple :**
```go
user, err := db.GetUserByUsername("JohnDoe")
if err != nil {
    // Gérer une erreur de base de données
    return
}
if user == nil {
    // L'utilisateur n'existe pas
    return
}
// Utiliser l'objet user...
fmt.Println(user.Email)
```

### `db.GetUserByID(id int) (*User, error)`

Récupère un utilisateur par son ID. Retourne `nil, nil` si aucun utilisateur n'est trouvé.

---

## 🔑 3. Sécurité et Mots de Passe

La comparaison des mots de passe se fait avec une fonction utilitaire.

### `utils.CheckPasswordHash(password, hash string) bool`

Compare un mot de passe en clair avec le hash stocké dans la base de données.

**Exemple de workflow de connexion :**
```go
user, _ := db.GetUserByUsername("JohnDoe")
if user != nil && utils.CheckPasswordHash("son_mot_de_passe_123", user.Password) {
    // Le mot de passe est correct, l'utilisateur est authentifié.
    // On peut maintenant créer une session.
} else {
    // Identifiants incorrects.
}
```

---

## 🍪 4. Fonctions de Session (`Session`)

Gère les sessions pour maintenir les utilisateurs connectés.

### `db.CreateSession(uuid, token, userID, duration, ip)`
Crée et stocke une nouvelle session pour un utilisateur.

### `db.GetSessionByUUID(uuid string) (*Session, error)`
Récupère les informations d'une session à partir de son UUID (généralement stocké dans un cookie). Utile pour vérifier si un utilisateur est authentifié.

### `db.DeleteSessionByUUID(uuid string) error`
Supprime une session de la base de données, ce qui correspond à une déconnexion.

---

## 📝 5. Fonctions des Posts (`Post`)

Gère la création et la récupération des articles du forum.

### `db.CreatePost(authorID, categoryID int, title, text string) error`
Crée un nouvel article dans une catégorie donnée.

### `db.GetPostByID(id int) (*Post, error)`
Récupère un article spécifique avec les informations de son auteur.

### `db.GetPostsByCategory(categoryID int) ([]Post, error)`
Récupère la liste de tous les articles pour une catégorie donnée, triés par date (du plus récent au plus ancien).

### `db.GetRecentPosts(limit int) ([]Post, error)`
Récupère les `limit` articles les plus récents, toutes catégories confondues.

---

## 💬 6. Fonctions des Commentaires (`Comment`)

Gère les commentaires sur les articles.

### `db.CreateComment(userID, postID int, repID sql.NullInt64, text string) error`
Ajoute un nouveau commentaire à un article. `repID` peut être utilisé si le commentaire est une réponse à un autre commentaire.

**Exemple pour un commentaire de premier niveau :**
```go
// repID est nul car ce n'est pas une réponse
err := db.CreateComment(userID, postID, sql.NullInt64{Valid: false}, "Mon super commentaire !")
```

### `db.GetCommentsByPostID(postID int) ([]Comment, error)`
Récupère la liste de tous les commentaires pour un article donné, triés par date (du plus ancien au plus récent).

---

## 📂 7. Fonctions des Catégories (`Category`)

### `db.GetCategories() ([]Category, error)`
Récupère la liste de toutes les catégories du forum.

**Exemple :**
```go
categories, err := db.GetCategories()
if err != nil {
    // Gérer l'erreur
}
for _, cat := range categories {
    fmt.Println(cat.Name)
}
```
