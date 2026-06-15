package db

var schema = map[string]string{
	"users": `CREATE TABLE users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        email TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
}

var categories = map[string]string{
	"Général":       "Discussions générales",
	"Programmation": "Tout sur le code",
	"IOT & Robotique": "Vous aimez les robotique et les objets connectés? Nous aussi!",
	"Cloud & Intelligence artificielle": "Tout sur le cloud et l'intelligence artificielle à la maison",
}
