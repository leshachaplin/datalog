run-clickhouse:
	docker run -d -e CLICKHOUSE_DB=test_db -e CLICKHOUSE_USER=su -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 -e CLICKHOUSE_PASSWORD=su \
 	-p 18123:8123 -p 19000:9000  --name clickhouse-server --ulimit nofile=262144:262144 clickhouse/clickhouse-server:latest-alpine

stop-clickhouse:
	docker stop clickhouse-server
	docker rm clickhouse-server
