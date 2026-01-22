import asyncio
import uuid
from contextlib import asynccontextmanager
from datetime import datetime
from typing import Optional

import aiosqlite
from fastapi import FastAPI, HTTPException, Depends
from pydantic import BaseModel

DB_PATH = "files.db"
POOL_SIZE = 10


class ConnectionPool:
    def __init__(self, db_path: str, pool_size: int):
        self.db_path = db_path
        self.pool_size = pool_size
        self.pool: asyncio.Queue = asyncio.Queue(maxsize=pool_size)

    async def init(self):
        for _ in range(self.pool_size):
            conn = await aiosqlite.connect(self.db_path)
            conn.row_factory = aiosqlite.Row
            # Enable WAL mode for better concurrency
            await conn.execute("PRAGMA journal_mode=WAL")
            await conn.commit()
            await self.pool.put(conn)

    async def close(self):
        while not self.pool.empty():
            conn = await self.pool.get()
            await conn.close()

    @asynccontextmanager
    async def get_connection(self):
        conn = await self.pool.get()
        try:
            yield conn
        finally:
            await self.pool.put(conn)


pool: Optional[ConnectionPool] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    global pool
    pool = ConnectionPool(DB_PATH, POOL_SIZE)
    await pool.init()

    # Initialize database schema and seed data
    async with pool.get_connection() as conn:
        await conn.execute(
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

        await conn.execute(
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

        await conn.commit()

    yield

    # Shutdown
    await pool.close()


app = FastAPI(lifespan=lifespan)


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
async def health_check():
    return "OK"


@app.post("/users", response_model=CreateUserResponse)
async def create_user(user: CreateUserRequest):
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


@app.post("/files", response_model=FileResponse)
async def create_file(file: CreateFileRequest):
    file_id = str(uuid.uuid4())
    now = datetime.utcnow()

    async with pool.get_connection() as conn:
        await conn.execute(
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
        await conn.commit()

    return FileResponse(
        id=file_id,
        created_at=now,
        updated_at=now,
        **file.dict(),
    )


@app.get("/files/{file_id}", response_model=FileResponse)
async def get_file(file_id: str):
    async with pool.get_connection() as conn:
        cursor = await conn.execute(
            "SELECT * FROM files WHERE id = ?",
            (file_id,),
        )
        row = await cursor.fetchone()

    if row is None:
        raise HTTPException(status_code=404, detail="File not found")

    return FileResponse(**dict(row))
