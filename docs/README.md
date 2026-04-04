# Documentation AlphaZeGo

Cette documentation suit la [structure Divio](https://docs.divio.com/documentation-system/) :

## Comprendre (Explanation)

- [Un framework générique de décision](explanation/framework-generique.md) -- Pourquoi State/PossibleMoves/Evaluate ne sont pas spécifiques aux jeux
- [Qu'est-ce qu'un ActorID ?](explanation/agent.md) -- Le décideur générique : identifiant, convention Evaluate=ActorID, et support 1/2/N acteurs
- [L'algorithme MCTS](explanation/mcts.md) -- Monte Carlo Tree Search : intuition, phases, et pourquoi il fonctionne
- [AlphaGo et les réseaux de neurones](explanation/alphago-et-reseaux-de-neurones.md) -- Comment AlphaGo utilise la convolution pour guider la recherche
- [De MCTS à AlphaZero](explanation/de-mcts-a-alphazero.md) -- PUCT, suppression des rollouts, et intégration du réseau dans l'arbre
- [MCTS + IA générative](explanation/mcts-genai.md) -- Coupler MCTS et LLM pour améliorer le raisonnement des petits modèles (expérimental)
- [Les problèmes du benchmark](explanation/problemes-benchmark.md) -- Les 10 problèmes d'ordonnancement, leur structure, leur chemin critique et leur intérêt

## Référence

- [ActorID](reference/agent-et-result.md) -- Type, constantes, conventions et exemples d'implémentation
- [Formules mathématiques](reference/formules.md) -- UCB1, PUCT, UCT : définitions et dérivations
- [Architecture du réseau de neurones](reference/architecture-reseau.md) -- Tenseurs d'entrée, blocs résiduels, têtes policy/value
- [Interfaces Go pour le réseau](reference/interfaces-evaluator.md) -- Spécification des interfaces `State`, `Evaluator` et `Tensorizable`

## Guides pratiques (How-to)

- [Intégrer un réseau de neurones dans le MCTS](how-to/integrer-reseau-neurones.md) -- Les modifications concrètes à apporter au code
- [Implémenter un Evaluator](how-to/implementer-evaluator.md) -- Comment créer un évaluateur (uniforme, rollout, ONNX)
- [Implémenter un nouveau jeu](how-to/implementer-un-jeu.md) -- Comment implémenter `board.State` pour connecter un jeu au MCTS
- [Ajouter un problème au benchmark](how-to/ajouter-probleme-benchmark.md) -- Comment créer et intégrer un nouveau problème d'ordonnancement

## Tutoriels

- [Le morpion pas à pas](tutorials/morpion-pas-a-pas.md) -- Construire un morpion jouable contre une IA MCTS, de zéro
