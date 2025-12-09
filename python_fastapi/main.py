import uuid

from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()

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
