import uuid
import sqlite3
from datetime import datetime
from json import JSONDecodeError

from flask import Flask, request, jsonify, g

app = Flask(__name__)

DB_PATH = "files.db"


def get_db():
    if "db" not in g:
        g.db = sqlite3.connect(DB_PATH)
        g.db.row_factory = sqlite3.Row
    return g.db


@app.teardown_appcontext
def close_db(exception):
    db = g.pop("db", None)
    if db is not None:
        db.close()


def init_db():
    db = sqlite3.connect(DB_PATH)
    db.execute(
        """
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
        """
    )

    now = datetime.utcnow()
    record = {
        "id": "b0320eab-57a6-4c45-ba6d-0b68a3501ef6",
        "directory_path": "cmd/server/",
        "filename": "main.go",
        "file_type": "file",
        "size": 123,
        "checksum": "1afb2837cb93eb1f3d68027adf777218",
        "created_at": now,
        "updated_at": now,
    }

    db.execute(
        """
        INSERT OR REPLACE INTO files (
            id, directory_path, filename, file_type,
            size, checksum, created_at, updated_at
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        """,
        tuple(record.values())
    )
    db.commit()
    db.commit()
    db.close()


init_db()


@app.route("/healthz")
def health_check():
    return "OK"


@app.route("/users", methods=["POST"])
def create_user():
    try:
        data = request.json
    except JSONDecodeError:
        return jsonify({"error": "Invalid JSON"}), 400

    return jsonify({
        "id": str(uuid.uuid4()),
        "first_name": data["first_name"],
        "last_name": data["last_name"],
        "email": data["email"],
    })


@app.route("/files/<file_id>", methods=["GET"])
def get_file(file_id):
    db = get_db()
    row = db.execute(
        "SELECT * FROM files WHERE id = ?",
        (file_id,)
    ).fetchone()

    if row is None:
        return jsonify({"error": "File not found"}), 404

    return jsonify(dict(row))


@app.route("/files", methods=["POST"])
def create_file():
    try:
        data = request.json
    except JSONDecodeError:
        return jsonify({"error": "Invalid JSON"}), 400

    now = datetime.utcnow()
    file_id = str(uuid.uuid4())

    record = {
        "id": file_id,
        "directory_path": data["directory_path"],
        "filename": data["filename"],
        "file_type": data["file_type"],
        "size": data["size"],
        "checksum": data["checksum"],
        "created_at": now,
        "updated_at": now,
    }

    db = get_db()
    db.execute(
        """
        INSERT INTO files (
            id, directory_path, filename, file_type,
            size, checksum, created_at, updated_at
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        """,
        tuple(record.values())
    )
    db.commit()

    return jsonify(record), 201
