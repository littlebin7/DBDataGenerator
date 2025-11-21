# 现有内容保持不变...

# 集成测试相关命令
.PHONY: test-integration test-integration-mysql test-integration-postgres test-integration-sqlite

# 运行集成测试（需要设置环境变量）
test-integration:
	@echo "运行集成测试..."
	@if [ -z "$$TEST_DB_ENABLED" ]; then \
		echo "错误: 请先设置 TEST_DB_ENABLED=true"; \
		exit 1; \
	fi
	go test ./... -v -run TestIntegration

# MySQL 集成测试
test-integration-mysql:
	TEST_DB_ENABLED=true \
	TEST_DB_TYPE=mysql \
	TEST_DB_HOST=localhost \
	TEST_DB_PORT=3306 \
	TEST_DB_USER=root \
	TEST_DB_PASSWORD=$$TEST_DB_PASSWORD \
	TEST_DB_NAME=test_dbdatagenerator \
	go test ./... -v -run TestIntegration

# PostgreSQL 集成测试
test-integration-postgres:
	TEST_DB_ENABLED=true \
	TEST_DB_TYPE=postgres \
	TEST_DB_HOST=localhost \
	TEST_DB_PORT=5432 \
	TEST_DB_USER=postgres \
	TEST_DB_PASSWORD=$$TEST_DB_PASSWORD \
	TEST_DB_NAME=test_dbdatagenerator \
	go test ./... -v -run TestIntegration

# SQLite 集成测试（不需要连接信息）
test-integration-sqlite:
	TEST_DB_ENABLED=true \
	TEST_DB_TYPE=sqlite \
	TEST_DB_NAME=test.db \
	go test ./... -v -run TestIntegration
