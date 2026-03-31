# Documentation AlphaZeGo

Cette documentation suit la [structure Divio](https://docs.divio.com/documentation-system/) :

## Comprendre (Explanation)

- [Un framework generique de decision](explanation/framework-generique.md) -- Pourquoi State/PossibleMoves/Evaluate ne sont pas specifiques aux jeux
- [Qu'est-ce qu'un PlayerID ?](explanation/agent.md) -- Le decideur generique : identifiant, convention Evaluate=PlayerID, et support 1/2/N joueurs
- [L'algorithme MCTS](explanation/mcts.md) -- Monte Carlo Tree Search : intuition, phases, et pourquoi il fonctionne
- [AlphaGo et les reseaux de neurones](explanation/alphago-et-reseaux-de-neurones.md) -- Comment AlphaGo utilise la convolution pour guider la recherche
- [De MCTS a AlphaZero](explanation/de-mcts-a-alphazero.md) -- PUCT, suppression des rollouts, et integration du reseau dans l'arbre

## Reference

- [PlayerID](reference/agent-et-result.md) -- Type, constantes, conventions et exemples d'implementation
- [Formules mathematiques](reference/formules.md) -- UCB1, PUCT, UCT : definitions et derivations
- [Architecture du reseau de neurones](reference/architecture-reseau.md) -- Tenseurs d'entree, blocs residuels, tetes policy/value
- [Interfaces Go pour le reseau](reference/interfaces-evaluator.md) -- Specification des interfaces `State`, `Evaluator` et `Tensorizable`

## Guides pratiques (How-to)

- [Integrer un reseau de neurones dans le MCTS](how-to/integrer-reseau-neurones.md) -- Les modifications concretes a apporter au code
- [Implementer un Evaluator](how-to/implementer-evaluator.md) -- Comment creer un evaluateur (uniforme, rollout, ONNX)

## Tutoriels

*(a venir)*
