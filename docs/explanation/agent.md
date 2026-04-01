# Qu'est-ce qu'un PlayerID ?

## L'intuition

Un `PlayerID` identifie un **décideur** : une entité qui prend des décisions dans un problème séquentiel. Le mot "joueur" est utilisé par convention, mais le concept est plus large. Selon le domaine, il désigne des choses très différentes :

| Domaine | PlayerID désigne | Exemple concret |
|---------|-----------------|-----------------|
| Jeu de plateau | Un joueur | Joueur X au morpion |
| Négociation | Une partie | L'acheteur, le vendeur |
| Diagnostic médical | Un décideur | Le médecin, le patient |
| Planification | Un acteur | Le planificateur |
| Vérification | Un rôle | Le système, l'attaquant |

Ce que tous ces cas ont en commun : à chaque étape, **un décideur agit**, puis le problème passe à l'étape suivante.

## PlayerID dans le code

Dans le framework, `PlayerID` est un type distinct basé sur `int` :

```go
type PlayerID int
```

Ce choix est délibéré :

- **Type distinct** : empêche les confusions avec d'autres `int` du code (le compilateur refuse les mélanges)
- **Valeurs négatives possibles** : `DrawResult = -1` ne peut jamais entrer en collision avec un identifiant de joueur positif
- **Pas de type Result séparé** : `Evaluate()` retourne un `PlayerID` — le gagnant, ou `NoPlayer`/`DrawResult`

Les constantes prédéfinies :

```go
const (
    NoPlayer   PlayerID = 0   // jeu en cours / case vide
    DrawResult PlayerID = -1  // match nul
    Player1    PlayerID = 1   // premier joueur
    Player2    PlayerID = 2   // second joueur
)
```

## Comment le MCTS utilise les PlayerID

Le moteur MCTS ne sait presque rien des joueurs. Il sait seulement deux choses :

1. **Qui doit agir maintenant ?** → `CurrentPlayer() PlayerID`
2. **Qui a agi pour arriver ici ?** → `PreviousPlayer() PlayerID`

C'est tout. Le MCTS ne connaît pas le nombre de joueurs, ne sait pas comment ils alternent, et ne sait pas ce que leurs identifiants signifient. Il utilise ces deux informations pour une seule chose : **créditer les victoires au bon joueur** lors de la rétropropagation.

```
Nœud : CurrentPlayer = 2, PreviousPlayer = 1
        → Le joueur 1 a joué le coup menant ici.
        → Si Evaluate() retourne Player1,
          ce nœud reçoit +1 win.
```

## La convention Evaluate = PlayerID du gagnant

Le framework utilise une convention simple pour relier joueurs et résultats :

> **Quand un joueur gagne, `Evaluate()` retourne son `PlayerID`.**

C'est ce que vérifie la rétropropagation :

```go
if result == n.state.PreviousPlayer() {
    n.wins += 1   // le joueur qui a joué ici a gagné
}
```

Cette convention fonctionne pour n'importe quel nombre de joueurs. Que le jeu ait 1, 2 ou 10 joueurs, la comparaison `result == playerID` reste valide. Et comme `DrawResult = -1`, il n'y a jamais de collision avec un identifiant de joueur.

## Un joueur, deux joueurs, N joueurs

### 2 joueurs (cas de référence)

Le morpion est le cas classique. Deux joueurs alternent : quand l'un joue, c'est au tour de l'autre. La logique de tour est triviale : `PreviousPlayer = 3 - CurrentPlayer` (puisque `3 - 1 = 2` et `3 - 2 = 1`).

Dans ce cas, le jeu est **adversarial** : ce qui est bon pour un joueur est mauvais pour l'autre. Le MCTS exploite cette propriété dans la backpropagation AlphaZero en inversant le signe de la valeur à chaque niveau de l'arbre.

### 1 joueur

Un problème à un seul joueur (planification, optimisation) est modélisable : l'unique joueur explore un arbre de décisions. `CurrentPlayer()` et `PreviousPlayer()` retournent toujours le même joueur. Le MCTS explore l'arbre normalement, et la rétropropagation crédite toujours le même joueur.

### N joueurs

Pour N > 2, chaque implémentation de `State` définit sa propre logique de tour via `PreviousPlayer()`. Le chemin MCTS pur (rollouts + backpropagation discrète) fonctionne directement. Le chemin AlphaZero (backpropagation avec alternance de signe) est limité à 2 joueurs pour l'instant.

## Pourquoi pas une interface Player plus riche ?

On pourrait imaginer un type `Player` avec des méthodes (nom, couleur, stratégie...). Le choix d'un simple `int` nommé est délibéré :

- **Pas de couplage** : le MCTS ne dépend d'aucune structure de joueur
- **Performance** : les comparaisons `result == playerID` sont des comparaisons d'entiers
- **Flexibilité** : n'importe quel problème peut attribuer des identifiants comme il veut
- **Simplicité** : moins de code, moins de surface d'API, moins de bugs

La richesse sémantique (nom du joueur, couleur, stratégie) est dans l'implémentation de `State`, pas dans le type `PlayerID` lui-même.

## `board.PlayerID` vs agent agentique

Le mot "agent" a deux sens qui se superposent dans les projets de recherche en IA :

| | `board.PlayerID` | Agent agentique |
|---|---|---|
| **Nature** | Un identifiant (`int`) | Un système autonome |
| **Question** | "À qui le tour ?" | "Qui décide ?" |
| **Exemples** | `Player1 = 1`, `Player2 = 2` | Un humain, un MCTS, un LLM |

Un `PlayerID` est un **rôle** dans le problème. Un agent agentique est le **système** qui occupe ce rôle. Le framework ne fait aucune distinction : il voit des rôles qui alternent, indépendamment de ce qui se passe "dans la tête" de chaque rôle.

Au morpion par exemple :

```
Partie de morpion
│
├── Rôle 1 (PlayerID = 1) ← occupé par : Humain
│   └── décide seul (intuition, réflexion)
│
└── Rôle 2 (PlayerID = 2) ← occupé par : Agent MCTS
    │
    ├── Mécanisme de recherche : MCTS (explore l'arbre)
    │
    └── Evaluator (sous-agent)
        ├── Policy : "quels chemins sont prometteurs ?"
        └── Value : "est-on bien ou mal parti ?"
```

L'humain et le MCTS occupent chacun un `PlayerID`, mais ce sont deux agents agentiques très différents. L'un réfléchit, l'autre calcule. Le framework les traite de manière identique.

## L'Evaluator comme sous-agent

L'agent MCTS est un **orchestrateur** : il ne "comprend" rien au problème. Il explore systématiquement un arbre de possibilités en répétant sélection, expansion, évaluation, rétropropagation. L'intelligence vient de l'`Evaluator`, qui est un sous-agent spécialisé.

L'`Evaluator` d'AlphaZero est en réalité **deux sous-agents fusionnés** dans un seul réseau de neurones à double tête :

| Sous-agent | Question | Sortie | Rôle dans le MCTS |
|------------|----------|--------|-------------------|
| **Policy** | "Par où chercher ?" | Distribution de probabilités sur les coups | Guide l'exploration (priors dans PUCT) |
| **Value** | "Comment ça se présente ?" | Score dans [-1, 1] | Remplace le rollout aléatoire |

Dans AlphaZero, ces deux sous-agents partagent un réseau (pour des raisons de performance : les couches profondes sont mutualisées). Mais conceptuellement, rien n'empêche de les séparer :

- Un **LLM** qui propose des pistes (policy) — "quelles sont les continuations les plus naturelles ?"
- Un **autre LLM** (ou le même avec un prompt différent) qui évalue des situations (value) — "ce raisonnement tient-il la route ?"
- Un **expert humain** qui donne son intuition (value) — "cette position me semble forte"
- Un **modèle statistique** qui estime des probabilités (policy) — "historiquement, ce coup est joué 40% du temps"

L'interface `Evaluator` est le point d'injection : le MCTS fournit le mécanisme de recherche, l'Evaluator fournit l'intelligence. Changer d'Evaluator change l'intelligence sans toucher à la recherche.

## MCTS comme moteur de raisonnement

Le framework n'est pas spécifique aux jeux. Si on pousse l'abstraction, il peut piloter un **système de raisonnement par recherche arborescente** — l'idée derrière les modèles de raisonnement modernes (Tree of Thoughts, recherche guidée par LLM).

### Le mapping

| Framework | Jeu de plateau | Raisonnement par LLM |
|-----------|---------------|---------------------|
| `State` | Position du plateau | Contexte partiel (prompt + raisonnement en cours) |
| `CurrentPlayer()` | Joueur actif | Le générateur (ou alternance générateur/critique) |
| `PossibleMoves()` | Coups légaux | Continuations possibles (prochaines étapes de raisonnement, appels d'outil, blocs de texte) |
| `Evaluate()` | Victoire/nul/en cours | La réponse est-elle complète et satisfaisante ? |
| `Evaluator.policy` | Priors sur les coups | "Quelles continuations ont le plus de chances de mener à une bonne réponse ?" |
| `Evaluator.value` | Estimation de la position | "Ce raisonnement partiel est-il sur la bonne voie ?" |

### Comment ca fonctionnerait

Au lieu de générer du texte token par token (de gauche à droite), on explore un **arbre de raisonnements** :

1. **Générer plusieurs continuations** — le LLM produit K complétions alternatives (via sampling avec température). C'est `PossibleMoves()`.
2. **Évaluer chacune** — un second appel au LLM (ou un modèle critique) estime la qualité du raisonnement partiel. C'est `Evaluator.value`.
3. **Choisir la plus prometteuse** — le MCTS utilise PUCT pour équilibrer exploration et exploitation.
4. **Approfondir** — on continue la génération depuis la branche choisie.
5. **Retenir le meilleur chemin** — après N itérations, le MCTS retourne la continuation la plus visitée.

### Générateur vs critique : un jeu à deux agents

L'aspect "deux joueurs" prend un sens inattendu ici. On peut modéliser le raisonnement comme un jeu entre :

- **Agent 1 — le générateur** : propose des raisonnements, essaie de construire une réponse convaincante
- **Agent 2 — le critique** : challenge les raisonnements, cherche des failles, des incohérences

Le MCTS équilibre les deux perspectives. Le générateur explore largement (exploration) ; le critique force l'approfondissement des pistes solides (exploitation). Le résultat est un raisonnement qui a survécu à un examen contradictoire.

Le LLM lui-même peut jouer tous les rôles :

| Rôle dans le framework | Le LLM l'implémente comment |
|------------------------|----------------------------|
| `PossibleMoves()` | Générer K complétions alternatives (sampling) |
| `Evaluator.policy` | Log-probabilités des tokens (quelle suite est naturelle ?) |
| `Evaluator.value` | Auto-évaluation ("ce raisonnement est-il cohérent ?") |
| `Evaluate()` | Critère d'arrêt ("la réponse est-elle finale et correcte ?") |

### Ce que le framework apporte

Le MCTS n'apporte pas d'intelligence — il apporte de la **rigueur dans l'exploration**. Un LLM seul génère un raisonnement linéaire. Un LLM guidé par MCTS explore un arbre de raisonnements, revient en arrière quand une branche est mauvaise, et converge vers le chemin le plus robuste. C'est la différence entre "penser tout haut" et "réfléchir méthodiquement".

## Résumé

```
PlayerID                Un rôle (identifiant int distinct)
NoPlayer (0)            Jeu en cours / case vide
DrawResult (-1)         Match nul (ne collisionne jamais avec un joueur)
Player1 (1), Player2(2) Joueurs prédéfinis pour les jeux à 2 joueurs
CurrentPlayer()         Quel rôle doit agir ?
PreviousPlayer()        Quel rôle a agi pour arriver ici ?
Evaluate() == PlayerID  Ce rôle a gagné
Evaluator               Sous-agent qui fournit policy + value
MCTS                    Orchestrateur qui fournit le mécanisme de recherche
```

`PlayerID` est la brique la plus simple du framework : un identifiant que le MCTS manipule sans chercher à comprendre ce qu'il représente. L'intelligence est dans l'`Evaluator`. La recherche est dans le MCTS. Le domaine est dans l'implémentation de `State`. Que le problème soit un jeu de plateau, une négociation, ou un raisonnement par LLM, le mécanisme est le même.
