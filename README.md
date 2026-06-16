# FORUM-PROJECT
this is the repo for the forum project by Romain Carrere, Thibaud Herry, Leo Gaiguant--Anquit, Joan Courty, Gaëtan 

## Choix Techniques & Contraintes

### Architecture et Bibliothèques
* **Langage :** Go (Golang) en version 1.26.3.
* **Approche sans framework :** Le projet utilise uniquement la bibliothèque standard Go pour le routage HTTP (`net/http.NewServeMux()`), renforçant la compréhension des mécanismes web bas niveau.
* **Bibliothèques tierces autorisées & utilisées :**
    * `modernc.org/sqlite` : Pilote SQLite natif en pur Go (sans CGO requis pour le build `CGO_ENABLED=0`).
    * `golang.org/x/crypto/bcrypt` : Pour le hachage sécurisé des mots de passe.
    * `github.com/google/uuid` : Pour la génération d'identifiants uniques (sessions, identifiants de posts).

### Structure du projet
L'architecture suit une séparation par domaines :
* `cmd/` : Point d'entrée de l'application (`main.go`).
* `internal/` : Cœur de la logique métier (base de données `db/`, configuration `config/`, routage et handlers `router/`, gestion des sessions `session/`).
* `pkg/` : Outils partagés et utilitaires (chargement d'environnement `openenv/`, système de logs, formatage du temps, sécurité, rendu des templates html).
* `template/` : Fichiers HTML pour les pages entières et sous-dossier `components/` pour les blocs réutilisables, chargés via un système de "Layout" automatisé en Go.
* `static/` : Fichiers statiques servis publiquement (CSS, JavaScript, images, icônes).

### Base de données (SQLite)
* Format : Fichier local `database.db` stocké dans le dossier `data/`.
* Comportement : Le fichier et les tables manquantes sont automatiquement générés au démarrage grâce au package `internal/db` (`db.Init()`). Les catégories par défaut sont injectées automatiquement.

## Schéma de la Base de Données

Le système gère les utilisateurs, le contenu textuel (articles, commentaires) et les interactions (likes, modération). 
La base se crée et s'initialise automatiquement au démarrage du serveur (`db.Init()`). Aucune commande SQL externe n'est requise.

### Description des tables
* **`users`** : Gère les membres inscrits. Contient l'ID, le pseudo, le mot de passe (haché avec bcrypt), l'email, le rôle (ex: "user", "moderator", "admin"), et un flag booléen `is_banned`.
* **`session`** : Système de maintien de connexion. Associe un token et un UUID à un ID utilisateur (`user_id`), avec une adresse IP d'origine et des dates de création et d'expiration.
* **`cat`** (Catégories) : Thématiques du forum (ex: "Programmation", "IOT & Robotique"). Auto-peuplée au premier démarrage.
* **`post`** : Les sujets de discussion. Chaque post est lié à un auteur (`users`), une catégorie (`cat`), contient un titre, un texte, et possède un `data_uuid` unique.
* **`post_message`** (Commentaires) : Les réponses aux articles. Liés à un auteur, un post cible, et incluent une notion de parenté (`rep_id`) pour gérer les réponses imbriquées (commentaires de commentaires).
* **`post_likes` & `comment_likes`** : Système de votes (Like = 1, Dislike = -1). Clef composite empêchant les votes multiples sur un même contenu par un même utilisateur.
* **`reports`** : Outil de modération. Lie un signalement à un utilisateur (`reporter_id`), identifie le type de contenu visé (`post` ou `comment`) via son `content_id`, et inclut un motif et un statut (par défaut "pending").
* **`connexions_data`** : Historique des connexions IP par utilisateur.

## Endpoints / Routes HTTP

L'application tourne sur les ports 80 (redirige automatiquement en HTTPS 301) et 443 (serveur TLS sécurisé avec certificats auto-générés). 

*Toutes les routes (sauf `/login`, `/register`, `/banned`, `/profile`, `/logout`) sont protégées par le middleware `checkBannedStatus` qui redirige les utilisateurs bannis.*

| Méthode | Route | Rôle & Protections | Retour/Action |
| :--- | :--- | :--- | :--- |
| **GET** | `/` | Accueil. Public. | Affiche catégories et les 5 posts récents. (200) |
| **GET/POST** | `/register` | Inscription. Public. | Vérifie unicité Pseudo/Email. Hache MDP. Redirige vers `/` en cas de succès. (303) |
| **GET/POST** | `/login` | Connexion. Public. | Création cookie + session en base. Redirige vers `/category/1`. (303) |
| **GET** | `/logout` | Déconnexion. Public. | Supprime la session base + expire cookie. (303) |
| **GET** | `/profile` | Profil utilisateur. **Auth requise**. | Affiche les posts créés, les likes, les commentaires postés. (200 / 303 si non auth) |
| **GET** | `/category/{id}`| Page de catégorie. Public. | Liste les posts de la catégorie. (200 / 404) |
| **GET** | `/post/{id}` | Lecture d'un article. Public. | Affiche l'article et l'arbre des commentaires. (200 / 404) |
| **GET/POST** | `/post/create` | Création d'article. **Auth requise**. | Crée un article. Redirection vers l'article publié. (303) |
| **POST** | `/post/{id}/comment`| Ajouter un commentaire. **Auth requise**.| Enregistre un commentaire ou une réponse (si `parent_id`). (303) |
| **POST** | `/like/post` | Liker un post. **Auth requise**. | Gère l'ajout, inversion, retrait d'un like/dislike. Retourne JSON des totaux. (200) |
| **POST** | `/like/comment` | Liker un comm. **Auth requise**. | Idem post. Retourne JSON des totaux. (200) |
| **POST** | `/report` | Signaler du contenu. **Auth requise**.| Crée une entrée dans la table `reports`. (303) |
| **GET** | `/banned` | Page pour bannis. Public. | Affiche le statut d'exclusion. (200) |
| **GET** | `/moderation` | Dashboard de modération. **Auth Modérateur/Admin**. | Liste les signalements en cours. (200 / 403) |
| **GET** | `/moderation/ban/{id}`| Bannir utilisateur. **Auth Modérateur/Admin**. | Bascule `is_banned` à 1. (303) |
| **GET** | `/moderation/unban/{id}`| Débannir util. **Auth Modérateur/Admin**. | Bascule `is_banned` à 0. (303) |
| **GET** | `/moderation/handle/delete/{id}`| Traiter signalement. **Auth Modérateur/Admin**. | Supprime le contenu visé (post/comment) et passe le rapport en "handled". (303) |
| **GET** | `/moderation/handle/dismiss/{id}`| Rejeter signalement. **Auth Modérateur/Admin**.| Passe le rapport en "handled" sans agir sur le contenu. (303) |
| **POST** | `/moderation/delete/post/{id}` | Sup. post arbitraire. **Auth Modérateur/Admin**.| Supprime l'article en cascade. (303) |
| **POST** | `/moderation/delete/comment/{id}`| Sup. comm arbitraire. **Auth Modérateur/Admin**.| Supprime le commentaire en cascade. (303) |

## Troubleshooting & Opérations courantes (Docker & App)

L'application est conteneurisée via Docker et stocke ses données dans des volumes montés à la racine du projet.

### Vider les volumes (Reset complet de la Base de Données)
La base de données et les certificats TLS sont persistés dans des dossiers locaux via Docker Volumes ou directement dans l'arborescence.
Pour effacer la base de données et repartir à zéro :
1. Stoppez l'application : `docker compose down`
2. Supprimez le fichier de base de données situé dans votre arborescence : 
   `rm ./data/database.db` (Linux/Mac) ou supprimez manuellement le dossier `data/`.
3. Relancez l'application (le fichier `database.db` se recréera vierge au lancement) : `docker compose up --build -d`

### Ports inaccessibles ou erreurs de démarrage
* L'application expose les ports **80** (HTTP) et **443** (HTTPS) dans le `docker-compose.yml` et le `config.json`.
* Si le lancement échoue avec une erreur type *"bind: address already in use"*, cela signifie que ces ports sont déjà occupés par un autre service (ex: Apache, Nginx, Skype). 
* **Solution :** Coupez les services concurrents.

### Lire les logs du serveur
Le projet intègre un système de logs personnalisé (`utils.Log`, `utils.LogError`). 
* En local ou via un terminal classique, les erreurs s'afficheront en rouge, et si l'application crashe (ex: fichier `.env` manquant), un log d'erreur FATAL apparaîtra.
* Si vous utilisez Docker, consultez les logs du conteneur en direct avec : 
    `docker logs -f forum_container`

### Gestion de l'environnement manquant (`.env`)
Le script `entrypoint.sh` est conçu pour parer aux oublis : si le fichier `.env` est manquant au lancement du conteneur, il copiera automatiquement le fichier `.env.exemple` pour permettre à l'application de démarrer sans crasher (`openenv` exigeant le fichier).