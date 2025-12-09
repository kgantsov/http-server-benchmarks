use actix_web::{get, post, web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};
use serde_json::json;
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

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| App::new().service(health_check).service(create_user))
        .bind(("127.0.0.1", 8080))?
        .run()
        .await
}
