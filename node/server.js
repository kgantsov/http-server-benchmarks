import app from "./app.js";

const HOST = "0.0.0.0";
const PORT = 8080;

app.listen(PORT, HOST, () => {
  console.log(`Server running on http://${HOST}:${PORT}`);
});
