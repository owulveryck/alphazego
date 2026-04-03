// Package reasoning implémente [decision.State] pour des problèmes de
// raisonnement par décomposition factuelle, où un LLM génère les étapes
// de raisonnement et un autre évalue leur qualité.
//
// L'état représente un contexte de raisonnement qui s'enrichit à chaque étape.
// Les branches de l'arbre MCTS sont des étapes de raisonnement candidates
// générées par un [Generator]. Un [Judge] évalue la qualité de chaque chemin
// de raisonnement et sert de base à l'[Evaluator] pour le MCTS AlphaZero.
//
// # Utilisation
//
// Le package définit deux interfaces LLM :
//
//   - [Generator] : génère N étapes de raisonnement candidates à partir
//     d'un prompt (question + étapes précédentes)
//   - [Judge] : attribue un score de qualité à un raisonnement
//
// Ces interfaces sont implémentées dans des modules séparés (ex: VertexAI,
// Ollama) avec leur propre go.mod.
//
// Le programme est générique : on fournit une question et un critère de
// succès (phrase décrivant la bonne réponse). Le raisonnement se termine
// quand une étape conclut (préfixe "CONCLUSION:") ou quand la profondeur
// maximale est atteinte.
//
// # Problème mono-acteur
//
// Comme le taquin, le raisonnement est un problème à un seul acteur.
// [State.CurrentActor] et [State.PreviousActor] retournent toujours [Player].
//
// # Compatibilité MCTS
//
// Le package fournit un [Evaluator] qui implémente [mcts.Evaluator] en
// wrappant un [Judge]. Il peut être utilisé avec [mcts.NewAlphaMCTS] pour
// guider l'exploration de l'arbre de raisonnement.
//
// Le MCTS pur (sans évaluateur) fonctionne aussi : les rollouts aléatoires
// choisissent des étapes au hasard parmi les candidats générés.
package reasoning
