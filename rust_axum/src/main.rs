use axum::{
    Router,
    extract::{Json, Path, State},
    routing::{get, post},
};
use chrono::{DateTime, Utc};
use rusqlite::{Connection, params};
use serde::{Deserialize, Serialize};
use std::sync::{Arc, Mutex};
use tokio::net::TcpListener;
use uuid::Uuid;
use validator::Validate;

#[derive(Serialize, Deserialize, Debug, Validate)]
pub struct CreateUserRequest {
    pub first_name: String,
    pub last_name: String,
    pub email: String,
}

#[derive(Serialize)]
pub struct CreateUserResponse {
    pub id: String,
    pub first_name: String,
    pub last_name: String,
    pub email: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct File {
    pub id: String,
    pub directory_path: String,
    pub filename: String,
    pub file_type: String,
    pub size: u64,
    pub checksum: String,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Deserialize, Validate)]
pub struct CreateFileRequest {
    pub directory_path: String,
    pub filename: String,
    pub file_type: String,
    pub size: u64,
    pub checksum: String,
}

type AppState = Arc<Mutex<Connection>>;

async fn health_check() -> &'static str {
    "OK"
}

async fn create_user(Json(info): Json<CreateUserRequest>) -> Json<CreateUserResponse> {
    Json(CreateUserResponse {
        id: Uuid::new_v4().to_string(),
        first_name: info.first_name,
        last_name: info.last_name,
        email: info.email,
    })
}

async fn get_file(
    Path(file_id): Path<String>,
    State(db): State<AppState>,
) -> axum::response::Result<Json<File>> {
    let conn = db.lock().unwrap();
    let mut stmt = conn
        .prepare(
            r#"
            SELECT id, directory_path, filename, file_type,
                   size, checksum, created_at, updated_at
            FROM files WHERE id = ?1
            "#,
        )
        .unwrap();

    let result = stmt.query_row(params![file_id], |row| {
        Ok(File {
            id: row.get::<_, String>(0)?,
            directory_path: row.get::<_, String>(1)?,
            filename: row.get::<_, String>(2)?,
            file_type: row.get::<_, String>(3)?,
            size: row.get::<_, i64>(4)? as u64,
            checksum: row.get::<_, String>(5)?,
            created_at: row.get::<_, String>(6)?,
            updated_at: row.get::<_, String>(7)?,
        })
    });

    match result {
        Ok(file) => Ok(Json(file)),
        Err(_) => Err(axum::http::StatusCode::NOT_FOUND.into()),
    }
}

async fn create_file(
    State(db): State<AppState>,
    Json(info): Json<CreateFileRequest>,
) -> Json<File> {
    let conn = db.lock().unwrap();
    let id = Uuid::new_v4().to_string();
    let now = chrono::Utc::now().to_rfc3339();

    conn.execute(
        "INSERT INTO files (id, directory_path, filename, file_type, size, checksum, created_at, updated_at)
         VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8)",
        params![
            id,
            info.directory_path,
            info.filename,
            info.file_type,
            info.size as i64,
            info.checksum,
            now,
            now
        ],
    ).unwrap();

    Json(File {
        id,
        directory_path: info.directory_path,
        filename: info.filename,
        file_type: info.file_type,
        size: info.size,
        checksum: info.checksum,
        created_at: now.clone(),
        updated_at: now,
    })
}

#[tokio::main]
async fn main() {
    let conn = Connection::open("files.db").unwrap();
    conn.execute(
        "CREATE TABLE IF NOT EXISTS files (
            id TEXT PRIMARY KEY,
            directory_path TEXT NOT NULL,
            filename TEXT NOT NULL,
            file_type TEXT NOT NULL,
            size INTEGER NOT NULL,
            checksum TEXT NOT NULL,
            created_at TEXT NOT NULL,
            updated_at TEXT NOT NULL
        )",
        [],
    )
    .unwrap();

    let now: DateTime<Utc> = Utc::now();
    conn.execute(
        "INSERT OR REPLACE INTO files (id, directory_path, filename, file_type, size, checksum, created_at, updated_at)
         VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8)",
        params![
            "b0320eab-57a6-4c45-ba6d-0b68a3501ef6",
            "cmd/server/",
            "main.go",
            "file",
            123,
            "1afb2837cb93eb1f3d68027adf777218",
            now,
            now
        ],
    ).unwrap();

    let db = Arc::new(Mutex::new(conn));

    let app = Router::new()
        .route("/healthz", get(health_check))
        .route("/users", post(create_user))
        .route("/files/{id}", get(get_file))
        .route("/files", post(create_file))
        .with_state(db);

    let listener = TcpListener::bind("127.0.0.1:8080")
        .await
        .expect("failed to bind");

    println!("Listening on http://127.0.0.1:8080");

    axum::serve(listener, app).await.unwrap();
}
