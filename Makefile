DATABASE_DSN ?= postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable

build:
	go build -o cmd/server/server ./cmd/server
	go build -o cmd/agent/agent ./cmd/agent

1: build
	./.tools/metricstest -test.v -test.run=^TestIteration1$$ \
		-binary-path=cmd/server/server

2: build
	./.tools/metricstest -test.v -test.run=^TestIteration2[AB]*$$ \
                  -source-path=. \
                  -agent-binary-path=cmd/agent/agent

3: build
	./.tools/metricstest -test.v -test.run=^TestIteration3[AB]*$$ \
                  -source-path=. \
                  -agent-binary-path=cmd/agent/agent \
                  -binary-path=cmd/server/server
4: build
	SERVER_PORT=$$(./.tools/random unused-port) && \
	ADDRESS="localhost:$$SERVER_PORT" && \
	TEMP_FILE=$$(./.tools/random tempfile) && \
	./.tools/metricstest -test.v -test.run=^TestIteration4$$ \
	   -agent-binary-path=cmd/agent/agent \
	   -binary-path=cmd/server/server \
	   -server-port=$$SERVER_PORT \
	   -source-path=.

5: build
	SERVER_PORT=$$(./.tools/random unused-port) && \
	ADDRESS="localhost:$$SERVER_PORT" && \
	TEMP_FILE=$$(./.tools/random tempfile) && \
	./.tools/metricstest -test.v -test.run=^TestIteration5$$ \
	   -agent-binary-path=cmd/agent/agent \
	   -binary-path=cmd/server/server \
	   -server-port=$$SERVER_PORT \
	   -source-path=.

6: build
	SERVER_PORT=$$(./.tools/random unused-port) && \
	ADDRESS="localhost:$$SERVER_PORT" && \
	TEMP_FILE=$$(./.tools/random tempfile) && \
	./.tools/metricstest -test.v -test.run=^TestIteration6$$ \
	   -agent-binary-path=cmd/agent/agent \
	   -binary-path=cmd/server/server \
	   -server-port=$$SERVER_PORT \
	   -source-path=.

7: build
	SERVER_PORT=$$(./.tools/random unused-port) && \
	ADDRESS="localhost:$$SERVER_PORT" && \
	TEMP_FILE=$$(./.tools/random tempfile) && \
	./.tools/metricstest -test.v -test.run=^TestIteration7$$ \
	   -agent-binary-path=cmd/agent/agent \
	   -binary-path=cmd/server/server \
	   -server-port=$$SERVER_PORT \
	   -source-path=.

8: build
	SERVER_PORT=$$(./.tools/random unused-port) && \
	ADDRESS="localhost:$$SERVER_PORT" && \
	TEMP_FILE=$$(./.tools/random tempfile) && \
	./.tools/metricstest -test.v -test.run=^TestIteration8$$ \
	   -agent-binary-path=cmd/agent/agent \
	   -binary-path=cmd/server/server \
	   -server-port=$$SERVER_PORT \
	   -source-path=.

9: build
	SERVER_PORT=$$(./.tools/random unused-port) && \
	ADDRESS="localhost:$$SERVER_PORT" && \
	TEMP_FILE=$$(./.tools/random tempfile) && \
	./.tools/metricstest -test.v -test.run=^TestIteration9$$ \
	   -agent-binary-path=cmd/agent/agent \
	   -binary-path=cmd/server/server \
	   -file-storage-path=$$TEMP_FILE \
	   -server-port=$$SERVER_PORT \
	   -source-path=.

10: build
	SERVER_PORT=$$(./.tools/random unused-port) && \
	ADDRESS="localhost:$$SERVER_PORT" && \
	TEMP_FILE=$$(./.tools/random tempfile) && \
	./.tools/metricstest -test.v -test.run=^TestIteration10[AB]$$ \
	   -agent-binary-path=cmd/agent/agent \
	   -binary-path=cmd/server/server \
	   -database-dsn='$(DATABASE_DSN)' \
	   -server-port=$$SERVER_PORT \
	   -source-path=.

11: build
	SERVER_PORT=$$(./.tools/random unused-port) && \
	ADDRESS="localhost:$$SERVER_PORT" && \
	TEMP_FILE=$$(./.tools/random tempfile) && \
	./.tools/metricstest -test.v -test.run=^TestIteration11$$ \
	   -agent-binary-path=cmd/agent/agent \
	   -binary-path=cmd/server/server \
	   -database-dsn='$(DATABASE_DSN)' \
	   -server-port=$$SERVER_PORT \
	   -source-path=.

12: build
	SERVER_PORT=$$(./.tools/random unused-port) && \
	ADDRESS="localhost:$$SERVER_PORT" && \
	TEMP_FILE=$$(./.tools/random tempfile) && \
	./.tools/metricstest -test.v -test.run=^TestIteration12$$ \
	   -agent-binary-path=cmd/agent/agent \
	   -binary-path=cmd/server/server \
	   -database-dsn='$(DATABASE_DSN)' \
	   -server-port=$$SERVER_PORT \
	   -source-path=.

13: build
	SERVER_PORT=$$(./.tools/random unused-port) && \
	ADDRESS="localhost:$$SERVER_PORT" && \
	TEMP_FILE=$$(./.tools/random tempfile) && \
	./.tools/metricstest -test.v -test.run=^TestIteration13$$ \
	   -agent-binary-path=cmd/agent/agent \
	   -binary-path=cmd/server/server \
	   -database-dsn='$(DATABASE_DSN)' \
	   -server-port=$$SERVER_PORT \
	   -source-path=.

14: build
	SERVER_PORT=$$(./.tools/random unused-port) && \
	ADDRESS="localhost:$$SERVER_PORT" && \
	TEMP_FILE=$$(./.tools/random tempfile) && \
	./.tools/metricstest -test.v -test.run=^TestIteration14$$ \
	   -agent-binary-path=cmd/agent/agent \
	   -binary-path=cmd/server/server \
	   -database-dsn='$(DATABASE_DSN)' \
	   -key="$$TEMP_FILE" \
	   -server-port=$$SERVER_PORT \
	   -source-path=.

statictest:
	go vet -vettool=$$(pwd)/.tools/statictest ./...

cover40:
	./.tools/covertest -test.v -test.run=^TestCoverage40$$

fmt:
	gofmt -l -w .
