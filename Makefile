.PHONY: all clean tictactoe taquin ollama-benchmark vertexai-benchmark vertexai-reasoning

BIN := bin

all: tictactoe taquin ollama-benchmark vertexai-benchmark vertexai-reasoning

tictactoe:
	go build -o $(BIN)/tictactoe ./decision/board/samples/tictactoe/cmd

taquin:
	go build -o $(BIN)/taquin ./decision/board/samples/taquin/cmd

ollama-benchmark:
	cd benchmark/ollama && go build -o ../../$(BIN)/ollama-benchmark ./cmd/benchmark

vertexai-benchmark:
	cd benchmark/vertexai && go build -o ../../$(BIN)/vertexai-benchmark ./cmd/benchmark

vertexai-reasoning:
	cd benchmark/vertexai && go build -o ../../$(BIN)/vertexai-reasoning ./cmd/reasoning

clean:
	rm -rf $(BIN)
