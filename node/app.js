import express from "express";
import { randomUUID } from "crypto";

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

export default app;
