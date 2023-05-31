run-clickhouse:
	docker run -d -e CLICKHOUSE_DB=test_db -e CLICKHOUSE_USER=su -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 -e CLICKHOUSE_PASSWORD=su \
 	-p 18123:8123 -p 19000:9000  --name clickhouse-server --ulimit nofile=262144:262144 clickhouse/clickhouse-server:latest-alpine

stop-clickhouse:
	docker stop clickhouse-server
	docker rm clickhouse-server

#
#migrate-up:
#	migrate -path /Users/alexeychaplin/GolandProjects/datalog/assets/migrations -database 'clickhouse://localhost:18123/test_db?username=su&password=su' -verbose up
#
#migrate-down:
