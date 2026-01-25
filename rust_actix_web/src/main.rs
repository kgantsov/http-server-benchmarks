use actix_web::{get, post, web, App, HttpResponse, HttpServer, Responder};
use chrono::{DateTime, Utc};
use rusqlite::{params, Connection};
use serde::{Deserialize, Serialize};
use serde_json::json;
use std::sync::Mutex;
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

#[get("/healthz")]
async fn health_check() -> impl Responder {
    HttpResponse::Ok().body("OK")
}

#[post("/users")]
async fn create_user(info: web::Json<CreateUserRequest>) -> impl Responder {
    HttpResponse::Ok().json(json!(&CreateUserResponse {
        id: Uuid::new_v4().to_string(),
        first_name: info.first_name.clone(),
        last_name: info.last_name.clone(),
        email: info.email.clone(),
    }))
}

#[derive(Serialize, Deserialize)]
pub struct FileMetadata {
    pub id: String,
    pub directory_path: String,
    pub filename: String,
    pub file_type: String,
    pub size: i64,
    pub checksum: String,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Deserialize)]
pub struct CreateFileRequest {
    pub directory_path: String,
    pub filename: String,
    pub file_type: String,
    pub size: i64,
    pub checksum: String,
}

#[post("/files")]
async fn create_file(
    db: web::Data<Mutex<Connection>>,
    req: web::Json<CreateFileRequest>,
) -> impl Responder {
    let id = Uuid::new_v4().to_string();
    let now: DateTime<Utc> = Utc::now();

    let conn = db.lock().unwrap();
    conn.execute(
        r#"
        INSERT INTO files (
            id, directory_path, filename, file_type,
            size, checksum, created_at, updated_at
        ) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8)
        "#,
        params![
            id,
            req.directory_path,
            req.filename,
            req.file_type,
            req.size,
            req.checksum,
            now,
            now
        ],
    )
    .unwrap();

    HttpResponse::Ok().json(FileMetadata {
        id,
        directory_path: req.directory_path.clone(),
        filename: req.filename.clone(),
        file_type: req.file_type.clone(),
        size: req.size,
        checksum: req.checksum.clone(),
        created_at: now,
        updated_at: now,
    })
}

#[get("/files/{id}")]
async fn get_file(db: web::Data<Mutex<Connection>>, file_id: web::Path<String>) -> impl Responder {
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

    let result = stmt.query_row(params![file_id.into_inner()], |row| {
        Ok(FileMetadata {
            id: row.get(0)?,
            directory_path: row.get(1)?,
            filename: row.get(2)?,
            file_type: row.get(3)?,
            size: row.get(4)?,
            checksum: row.get(5)?,
            created_at: row.get(6)?,
            updated_at: row.get(7)?,
        })
    });

    match result {
        Ok(file) => HttpResponse::Ok().json(file),
        Err(_) => HttpResponse::NotFound().body("File not found"),
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let conn = Connection::open("files.db").unwrap();

    conn.execute_batch("PRAGMA journal_mode=WAL;").unwrap();

    conn.execute(
        r#"
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
        "#,
        [],
    )
    .unwrap();

    let now: DateTime<Utc> = Utc::now();
    conn.execute(
        r#"
        INSERT OR REPLACE INTO files (
            id, directory_path, filename, file_type,
            size, checksum, created_at, updated_at
        ) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8)
        "#,
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
    )
    .unwrap();

    let db = web::Data::new(Mutex::new(conn));

    HttpServer::new(move || {
        App::new()
            .app_data(db.clone())
            .service(health_check)
            .service(create_user)
            .service(create_file)
            .service(get_file)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}
