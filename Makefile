start_node:
	cd node && pm2 start server.js -i 8 --name node-bench

stop_node:
	cd node && pm2 stop node-bench

start_python_flask:
	cd python_flask && gunicorn -w 8 'main:app' -b 0.0.0.0:8080

start_python_fastapi:
	cd python_fastapi && uvicorn main:app --host 0.0.0.0 --port 8080 --workers 8 --log-level warning

start_go_fiber:
	cd go_fiber && go run .

start_rust_actix_web:
	cd rust_actix_web && cargo run --release

run_bench_get:
	hammerload --duration 10 --concurrency 200 http -u http://localhost:8080/healthz

run_bench_post:
	hammerload --duration 10 --concurrency 200 http -X POST -u http://localhost:8080/users -H "Content-Type: application/json" --body '{"first_name": "John", "last_name": "Doe", "email": "john.doe@gmail.com"}'
