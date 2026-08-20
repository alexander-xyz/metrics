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