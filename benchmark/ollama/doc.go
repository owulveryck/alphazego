// Package ollama fournit des implémentations de [reasoning.Generator] et
// [reasoning.Judge] utilisant un modèle local via Ollama.
//
// Ce package fait partie du répertoire benchmark/ qui explore de manière
// expérimentale le couplage entre MCTS et IA générative. Il permet de
// tester l'hypothèse sur des modèles locaux sans dépendre du cloud.
//
// # API Ollama
//
// Le package communique avec le serveur Ollama via son API REST :
//
//	POST http://localhost:11434/api/generate
//	{
//	    "model": "qwen2.5:7b",
//	    "prompt": "...",
//	    "stream": false,
//	    "options": {"temperature": 0.8}
//	}
//
// La réponse contient le texte généré et les compteurs de tokens
// (prompt_eval_count, eval_count) pour le suivi de la consommation.
//
// Les appels au [Generator] sont séquentiels (pas de parallélisme) car
// le GPU local est le goulot d'étranglement. La parallélisation des
// requêtes ne ferait que les mettre en file d'attente.
//
// # Benchmark
//
// Le benchmark compare 2 configurations avec le même modèle local :
//
//	| Config | Méthode  | Description                         |
//	|--------|----------|-------------------------------------|
//	| E      | One-shot | Modèle local, un seul appel         |
//	| F      | MCTS     | Modèle local + raisonnement MCTS    |
//
// Le même modèle sert aussi de juge (LLM-as-Judge) pour évaluer les
// solutions. C'est une limitation connue : un petit modèle peut être
// un juge peu fiable. Les résultats sont donc à interpréter avec
// précaution.
//
// # Prérequis
//
//   - Ollama installé et en cours d'exécution (https://ollama.com)
//   - Un modèle téléchargé : ollama pull qwen2.5:7b
//   - L'URL par défaut est http://localhost:11434
//
// # Utilisation
//
//	gen := ollama.NewGenerator("http://localhost:11434", "qwen2.5:7b")
//	judge := ollama.NewJudge("http://localhost:11434", "qwen2.5:7b")
//
//	state := reasoning.New(ctx, question, criterion, gen)
//	eval := reasoning.NewEvaluator(ctx, judge)
//	m := mcts.NewAlphaMCTS(eval, 1.5)
//	best := m.RunMCTS(state, 15)
package ollama
