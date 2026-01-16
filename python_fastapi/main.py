import uuid
import sqlite3
from datetime import datetime
from typing import Optional

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI()

DB_PATH = "files.db"

class CreateUserRequest(BaseModel):
    first_name: str
    last_name: str
    email: str


class CreateUserResponse(BaseModel):
    id: str
    first_name: str
    last_name: str
    email: str


@app.get("/healthz")
def health_check():
    return "OK"


@app.post("/users", response_model=CreateUserResponse)
def create_user(user: CreateUserRequest):
    user_id = str(uuid.uuid4())
    return CreateUserResponse(id=user_id, **user.dict())


class CreateFileRequest(BaseModel):
    directory_path: str
    filename: str
    file_type: str
    size: int
    checksum: str


class FileResponse(BaseModel):
    id: str
    directory_path: str
    filename: str
    file_type: str
    size: int
    checksum: str
    created_at: datetime
    updated_at: datetime


def get_db():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn


@app.on_event("startup")
def init_db():
    conn = get_db()
    conn.execute(
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

    conn.execute(
        """
        INSERT OR REPLACE INTO files (
            id, directory_path, filename, file_type,
            size, checksum, created_at, updated_at
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        """,
        (
            "b0320eab-57a6-4c45-ba6d-0b68a3501ef6",
            "cmd/server/",
            "main.go",
            "file",
            123,
            "1afb2837cb93eb1f3d68027adf777218",
            now,
            now,
        ),
    )
    conn.commit()
    conn.close()


@app.post("/files", response_model=FileResponse)
def create_file(file: CreateFileRequest):
    file_id = str(uuid.uuid4())
    now = datetime.utcnow()

    conn = get_db()
    conn.execute(
        """
        INSERT INTO files (
            id, directory_path, filename, file_type,
            size, checksum, created_at, updated_at
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        """,
        (
            file_id,
            file.directory_path,
            file.filename,
            file.file_type,
            file.size,
            file.checksum,
            now,
            now,
        ),
    )
    conn.commit()
    conn.close()

    return FileResponse(
        id=file_id,
        created_at=now,
        updated_at=now,
        **file.dict(),
    )


@app.get("/files/{file_id}", response_model=FileResponse)
def get_file(file_id: str):
    conn = get_db()
    row = conn.execute(
        "SELECT * FROM files WHERE id = ?",
        (file_id,),
    ).fetchone()
    conn.close()

    if row is None:
        raise HTTPException(status_code=404, detail="File not found")

    return FileResponse(**dict(row))
