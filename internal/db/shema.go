package db

var schema = map[string]string{
	"users": `CREATE TABLE users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        email TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user',
		is_banned INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`,
	"reports": `CREATE TABLE "reports" (
		"id"	INTEGER NOT NULL UNIQUE,
		"reporter_id"	INTEGER NOT NULL,
		"content_id"	INTEGER NOT NULL,
		"content_type"	TEXT NOT NULL,
		"reason"	TEXT NOT NULL,
		"status" TEXT NOT NULL DEFAULT 'pending',
		"created_at"	DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY("id" AUTOINCREMENT),
		FOREIGN KEY("reporter_id") REFERENCES "users"("id")
	);`,
	"connexions_data": `CREATE TABLE "connexions_data" (
		"id"	INTEGER NOT NULL UNIQUE,
		"user_id"	INTEGER NOT NULL,
		"ip"	TEXT,
		"connected_at"	DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY("id" AUTOINCREMENT),
		FOREIGN KEY("user_id") REFERENCES "users"("id")
	);`,
	"session": `CREATE TABLE "session" (
		"id"	INTEGER NOT NULL UNIQUE,
		"uuid"	TEXT NOT NULL UNIQUE,
		"token"	TEXT NOT NULL UNIQUE,
		"user_id"	INTEGER NOT NULL,
		"created_at"	DATETIME NOT NULL,
		"end_at"	DATETIME NOT NULL,
		"ip"	TEXT NOT NULL,
		PRIMARY KEY("id" AUTOINCREMENT),
		FOREIGN KEY("user_id") REFERENCES "users"("id")
	);`,
	"cat": `CREATE TABLE "cat" (
		"id"	INTEGER NOT NULL UNIQUE,
		"name"	TEXT NOT NULL UNIQUE,
		"desc"	TEXT NOT NULL,
		PRIMARY KEY("id" AUTOINCREMENT)
	);`,
	"post": `CREATE TABLE "post" (
		"id"	INTEGER NOT NULL UNIQUE,
		"author"	INTEGER NOT NULL,
		"category_id" INTEGER NOT NULL,
		"title"	TEXT NOT NULL,
		"text"	TEXT NOT NULL,
		"data_uuid"	TEXT NOT NULL UNIQUE,
		"timestamp"	DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY("id" AUTOINCREMENT),
		FOREIGN KEY("author") REFERENCES "users"("id"),
		FOREIGN KEY("category_id") REFERENCES "cat"("id")
	);`,
	"post_message": `CREATE TABLE "post_message" (
		"id"	INTEGER NOT NULL UNIQUE,
		"user_id"	INTEGER NOT NULL,
		"post_id"	INTEGER NOT NULL,
		"rep_id"	INTEGER,
		"text"	TEXT NOT NULL,
		"timestamp"	DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY("id" AUTOINCREMENT),
		FOREIGN KEY("post_id") REFERENCES "post"("id"),
		FOREIGN KEY("user_id") REFERENCES "users"("id")
	);`,
	"post_likes": `CREATE TABLE "post_likes" (
		"post_id" INTEGER NOT NULL,
		"user_id" INTEGER NOT NULL,
		"type" INTEGER NOT NULL, -- 1 for like, -1 for dislike
		PRIMARY KEY("post_id", "user_id"),
		FOREIGN KEY("post_id") REFERENCES "post"("id") ON DELETE CASCADE,
		FOREIGN KEY("user_id") REFERENCES "users"("id") ON DELETE CASCADE
	);`,
	"comment_likes": `CREATE TABLE "comment_likes" (
		"comment_id" INTEGER NOT NULL,
		"user_id" INTEGER NOT NULL,
		"type" INTEGER NOT NULL, -- 1 for like, -1 for dislike
		PRIMARY KEY("comment_id", "user_id"),
		FOREIGN KEY("comment_id") REFERENCES "post_message"("id") ON DELETE CASCADE,
		FOREIGN KEY("user_id") REFERENCES "users"("id") ON DELETE CASCADE
	);`,
}

var categories = map[string]string{
	"Général":                           "Discussions générales",
	"Programmation":                     "Tout sur le code",
	"IOT & Robotique":                   "Vous aimez les robotique et les objets connectés? Nous aussi!",
	"Cloud & Intelligence artificielle": "Tout sur le cloud et l'intelligence artificielle à la maison",
}
