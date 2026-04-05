.PHONY: all clean test test-cover lint benchmark tictactoe taquin ollama-benchmark vertexai-benchmark vertexai-reasoning

BIN := bin

all: tictactoe taquin ollama-benchmark vertexai-benchmark vertexai-reasoning

test:
	go test ./...

test-cover:
	go test -cover ./...

lint:
	go vet ./...
	golangci-lint run ./...

benchmark:
	go test -bench=. -benchmem ./mcts/ ./decision/board/samples/tictactoe/ ./decision/board/samples/taquin/

tictactoe:
	go build -o $(BIN)/tictactoe ./decision/board/samples/tictactoe/cmd

taquin:
	go build -o $(BIN)/taquin ./decision/board/samples/taquin/cmd

ollama-benchmark:
	cd exp/benchmark/ollama && go build -o ../../../$(BIN)/ollama-benchmark ./cmd/benchmark

vertexai-benchmark:
	cd exp/benchmark/vertexai && go build -o ../../../$(BIN)/vertexai-benchmark ./cmd/benchmark

vertexai-reasoning:
	cd exp/benchmark/vertexai && go build -o ../../../$(BIN)/vertexai-reasoning ./cmd/reasoning

clean:
	rm -rf $(BIN)
