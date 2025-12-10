use axum::{
    Router,
    extract::Json,
    routing::{get, post},
};
use serde::{Deserialize, Serialize};
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

#[tokio::main]
async fn main() {
    let app = Router::new()
        .route("/healthz", get(health_check))
        .route("/users", post(create_user));

    let listener = TcpListener::bind("127.0.0.1:8080")
        .await
        .expect("failed to bind");

    println!("Listening on http://127.0.0.1:8080");

    axum::serve(listener, app).await.unwrap();
}
