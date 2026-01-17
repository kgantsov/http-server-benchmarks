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

start_go_net_http:
	cd go_net_http && go run .

start_rust_actix_web:
	cd rust_actix_web && cargo run --release

start_rust_axum:
	cd rust_axum && cargo run --release

run_bench_get:
	hammerload --duration 10 --concurrency 200 http -u http://localhost:8080/healthz

run_bench_post:
	hammerload --duration 10 --concurrency 200 http -X POST -u http://localhost:8080/users -H "Content-Type: application/json" --body '{"first_name": "John", "last_name": "Doe", "email": "john.doe@gmail.com"}'

run_bench_get_file:
	hammerload --duration 10 --concurrency 200 http -X GET -u http://localhost:8080/files/b0320eab-57a6-4c45-ba6d-0b68a3501ef6

run_bench_post_file:
	hammerload --duration 10 --concurrency 200 http -X POST -u http://localhost:8080/files -H "Content-Type: application/json" -b '{"filename": "test.txt", "directory_path": "", "file_type": "file", "checksum": "checksum", "size": 0}'
