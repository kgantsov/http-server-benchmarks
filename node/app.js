import express from "express";
import { randomUUID } from "crypto";
import sqlite3 from "sqlite3";
import { open } from "sqlite";

const app = express();
app.use(express.json());

// Health check
app.get("/healthz", (req, res) => {
  res.send("OK");
});

// Create user
app.post("/users", (req, res) => {
  const { first_name, last_name, email } = req.body;

  const user = {
    id: randomUUID(),
    first_name,
    last_name,
    email,
  };

  res.json(user);
});

let db;

async function initDb() {
  db = await open({
    filename: "./files.db",
    driver: sqlite3.Database,
  });

  await db.exec(`PRAGMA journal_mode=WAL;`);

  await db.exec(`
    CREATE TABLE IF NOT EXISTS files (
      id TEXT PRIMARY KEY,
      directory_path TEXT NOT NULL,
      filename TEXT NOT NULL,
      file_type TEXT NOT NULL,
      size INTEGER NOT NULL,
      checksum TEXT NOT NULL,
      created_at DATETIME NOT NULL,
      updated_at DATETIME NOT NULL
    )
  `);

  // const now = new Date().toISOString();
  const now = new Date();

  await db.run(
    `
    INSERT OR REPLACE INTO files (
      id, directory_path, filename, file_type,
      size, checksum, created_at, updated_at
    )
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `,
    [
      "b0320eab-57a6-4c45-ba6d-0b68a3501ef6",
      "cmd/server/",
      "main.go",
      "file",
      123,
      "1afb2837cb93eb1f3d68027adf777218",
      now,
      now,
    ],
  );
}

initDb().catch((err) => {
  console.error("Failed to initialize database:", err);
  process.exit(1);
});

app.get("/files/:id", async (req, res) => {
  try {
    const file = await db.get(
      "SELECT * FROM files WHERE id = ?",
      req.params.id,
    );

    if (!file) {
      return res.status(404).json({ error: "File not found" });
    }

    res.json(file);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

app.post("/files", async (req, res) => {
  try {
    const { directory_path, filename, file_type, size, checksum } = req.body;

    const now = new Date().toISOString();
    const id = randomUUID();

    await db.run(
      `
      INSERT INTO files (
        id, directory_path, filename, file_type,
        size, checksum, created_at, updated_at
      )
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
      `,
      [id, directory_path, filename, file_type, size, checksum, now, now],
    );

    res.status(201).json({
      id,
      directory_path,
      filename,
      file_type,
      size,
      checksum,
      created_at: now,
      updated_at: now,
    });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

export default app;
