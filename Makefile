.PHONY: all clean test test-cover lint benchmark tictactoe taquin ollama-benchmark vertexai-benchmark vertexai-reasoning \
       prof-cpu prof-mem prof-trace bench-compare bench-save inlining

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

# --- Profiling ---

# Sauvegarder les résultats de benchmark dans old.txt (baseline).
bench-save:
	go test -bench=. -benchmem -count=10 ./mcts/ | tee old.txt

# Comparer les benchmarks actuels avec la baseline (old.txt).
# Nécessite : go install golang.org/x/perf/cmd/benchstat@latest
bench-compare:
	go test -bench=. -benchmem -count=10 ./mcts/ | tee new.txt
	benchstat old.txt new.txt

# Profiling CPU sur BenchmarkRunMCTS_10000 (5 secondes).
prof-cpu:
	go test -bench=BenchmarkRunMCTS_10000 -benchmem -cpuprofile=cpu.prof -benchtime=5s ./mcts/
	go tool pprof -http=:8080 cpu.prof

# Profiling mémoire (allocations) sur BenchmarkRunMCTS_10000.
prof-mem:
	go test -bench=BenchmarkRunMCTS_10000 -benchmem -memprofile=mem.prof -benchtime=5s ./mcts/
	go tool pprof -http=:8080 mem.prof

# Trace d'exécution pour analyser la concurrence et le scheduling.
prof-trace:
	go test -bench=BenchmarkRunMCTS_1000 -trace=trace.out ./mcts/
	go tool trace trace.out

# Vérifier les décisions d'inlining du compilateur Go.
inlining:
	@echo "=== mcts/ ==="
	@go build -gcflags="-m -m" ./mcts/ 2>&1 | grep -E "(can inline|cannot inline)" | sort
	@echo ""
	@echo "=== tictactoe/ ==="
	@go build -gcflags="-m -m" ./decision/board/samples/tictactoe/ 2>&1 | grep -E "(can inline|cannot inline)" | sort

clean:
	rm -rf $(BIN) cpu.prof mem.prof trace.out old.txt new.txt
