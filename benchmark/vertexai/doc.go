// Package vertexai fournit des implémentations de [reasoning.Generator] et
// [reasoning.Judge] utilisant Google Vertex AI (Gemini).
//
// Ce package fait partie du répertoire benchmark/ qui explore de manière
// expérimentale le couplage entre MCTS et IA générative. Il n'est pas
// destiné à un usage en production.
//
// # Modèles
//
// Deux modèles Gemini sont utilisés avec des rôles distincts :
//
//   - [GeneratorModel] (gemini-3.1-pro-preview) : génère les étapes de
//     raisonnement candidates. C'est le « gros modèle » qui produit des
//     réponses de haute qualité. Température 0.8 pour favoriser la
//     diversité entre les candidats.
//   - [JudgeModel] (gemini-3.1-flash-lite-preview) : évalue la qualité
//     des raisonnements. C'est le « petit modèle » rapide, utilisé avec
//     thinking level "low" et température 0.0 pour la reproductibilité.
//
// Le benchmark compare 4 configurations :
//
//	| Config | Modèle     | Méthode  | Description                      |
//	|--------|------------|----------|----------------------------------|
//	| A      | flash-lite | One-shot | Petit modèle, un seul appel      |
//	| B      | flash-lite | MCTS     | Petit modèle + raisonnement MCTS |
//	| C      | pro        | One-shot | Gros modèle, un seul appel       |
//	| D      | pro        | MCTS     | Gros modèle + raisonnement MCTS  |
//
// La comparaison clé est A vs B : le MCTS aide-t-il un petit modèle ?
// La comparaison secondaire est B vs C : un petit modèle + MCTS fait-il
// aussi bien qu'un gros modèle seul ?
//
// # Configuration
//
// Le package nécessite un projet Google Cloud avec l'API Vertex AI activée.
// L'authentification se fait via les credentials par défaut :
//
//	export GCP_PROJECT=mon-projet
//	export GCP_REGION=us-central1
//	gcloud auth application-default login
//
// # Commandes
//
// Deux commandes sont disponibles :
//
//   - cmd/benchmark : exécute le benchmark complet (10 problèmes × 4 configs)
//     avec scoring LLM-as-Judge, timing et comptage de tokens.
//   - cmd/reasoning : CLI de raisonnement interactif pour poser une question
//     libre et observer la décomposition MCTS étape par étape.
//
// # Exemple
//
//	client, err := vertexai.NewClient(ctx, project, region)
//	gen := vertexai.NewGenerator(client)
//	judge := vertexai.NewJudge(client)
//
//	state := reasoning.New(ctx, question, criterion, gen,
//	    reasoning.WithMaxDepth(5),
//	    reasoning.WithBranchFactor(3),
//	)
//	eval := reasoning.NewEvaluator(ctx, judge)
//	m := mcts.NewAlphaMCTS(eval, 1.5)
//	best := m.RunMCTS(state, 15)
package vertexai
