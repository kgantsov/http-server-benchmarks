import uuid

from json import JSONDecodeError

from flask import Flask, request, jsonify

app = Flask(__name__)

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
