# Qu'est-ce qu'un Agent ?

## L'intuition

Un Agent est un **decideur** : une entite qui prend des decisions dans un probleme sequentiel. Le mot "agent" est volontairement generique. Selon le domaine, il designe des choses tres differentes :

| Domaine | Agent | Exemple concret |
|---------|-------|-----------------|
| Jeu de plateau | Un joueur | Joueur X au morpion |
| Negociation | Une partie | L'acheteur, le vendeur |
| Diagnostic medical | Un decideur | Le medecin, le patient |
| Planification | Un acteur | Le planificateur |
| Verification | Un role | Le systeme, l'attaquant |

Ce que tous ces cas ont en commun : a chaque etape, **un agent agit**, puis le probleme passe a l'etape suivante.

## Agent dans le code

Dans le framework, `Agent` est un simple `uint8`. C'est un identifiant numerique, rien de plus :

```go
type Agent = uint8
```

Ce choix est delibere : un `uint8` est le type le plus leger possible. Il n'impose aucune structure, aucune hierarchie, aucune semantique. C'est a l'implementation de `State` de donner un sens a chaque valeur d'Agent.

## Comment le MCTS utilise les agents

Le moteur MCTS ne sait presque rien des agents. Il sait seulement deux choses :

1. **Qui doit agir maintenant ?** → `CurrentPlayer() Agent`
2. **Qui a agi pour arriver ici ?** → `PreviousPlayer() Agent`

C'est tout. Le MCTS ne connait pas le nombre d'agents, ne sait pas comment ils alternent, et ne sait pas ce que leurs identifiants signifient. Il utilise ces deux informations pour une seule chose : **crediter les victoires au bon agent** lors de la retropropagation.

```
Noeud : CurrentPlayer = 2, PreviousPlayer = 1
        → L'agent 1 a joue le coup menant ici.
        → Si le resultat final est "Agent 1 gagne",
          ce noeud recoit +1 win.
```

## La convention Agent = Result

Le framework utilise une convention simple pour relier agents et resultats :

> **Quand un agent gagne, le `Result` est egal a l'`Agent`.**

Autrement dit, `Result(a)` signifie "l'agent `a` a gagne". C'est ce que verifie la retropropagation :

```go
if result == n.state.PreviousPlayer() {
    n.wins += 1   // l'agent qui a joue ici a gagne
}
```

Cette convention fonctionne pour n'importe quel nombre d'agents. Que le jeu ait 1, 2 ou 10 agents, la comparaison `result == agent` reste valide.

## Un agent, deux agents, N agents

### 2 agents (cas de reference)

Le morpion est le cas classique. Deux agents alternent : quand l'un joue, c'est au tour de l'autre. La logique de tour est triviale : `PreviousPlayer = 3 - CurrentPlayer` (puisque `3 - 1 = 2` et `3 - 2 = 1`).

Dans ce cas, le jeu est **adversarial** : ce qui est bon pour un agent est mauvais pour l'autre. Le MCTS exploite cette propriete dans `BackpropagateValue` (chemin AlphaZero) en inversant le signe de la valeur a chaque niveau de l'arbre.

### 1 agent

Un probleme a un seul agent (planification, optimisation) est modelisable : l'unique agent explore un arbre de decisions. `CurrentPlayer()` et `PreviousPlayer()` retournent toujours le meme agent. Le MCTS explore l'arbre normalement, et la retropropagation credite toujours le meme agent.

### N agents

Pour N > 2, chaque implementation de `State` definit sa propre logique de tour via `PreviousPlayer()`. Le chemin MCTS pur (rollouts + `Backpropagate` discret) fonctionne directement. Le chemin AlphaZero (`BackpropagateValue` avec alternance de signe) est limite a 2 joueurs pour l'instant.

**Attention aux collisions** : les constantes `Draw = 3` et `Stalemat = 4` occupent les valeurs 3 et 4. Si un agent a l'identifiant 3, il est confondu avec un match nul. Pour les jeux a 3+ joueurs, utiliser des identifiants d'agents >= 5.

## Pourquoi ne pas utiliser une interface plus riche ?

On pourrait imaginer un type `Agent` avec des methodes (nom, couleur, strategie...). Le choix d'un simple `uint8` est delibere :

- **Pas de couplage** : le MCTS ne depend d'aucune structure d'agent
- **Performance** : les comparaisons `result == agent` sont des comparaisons d'entiers
- **Flexibilite** : n'importe quel probleme peut attribuer des identifiants comme il veut
- **Simplicite** : moins de code, moins de surface d'API, moins de bugs

La richesse semantique (nom du joueur, couleur, strategie) est dans l'implementation de `State`, pas dans le type `Agent` lui-meme.

## `board.Agent` vs agent agentique

Le mot "agent" a deux sens qui se superposent dans ce projet :

| | `board.Agent` | Agent agentique |
|---|---|---|
| **Nature** | Un identifiant (`uint8`) | Un systeme autonome |
| **Question** | "A qui le tour ?" | "Qui decide ?" |
| **Exemples** | `Player1 = 1`, `Player2 = 2` | Un humain, un MCTS, un LLM |

Un `board.Agent` est un **role** dans le probleme. Un agent agentique est le **systeme** qui occupe ce role. Le framework ne fait aucune distinction : il voit des roles qui alternent, independamment de ce qui se passe "dans la tete" de chaque role.

Au morpion par exemple :

```
Partie de morpion
│
├── Role 1 (board.Agent = 1) ← occupe par : Humain
│   └── decide seul (intuition, reflexion)
│
└── Role 2 (board.Agent = 2) ← occupe par : Agent MCTS
    │
    ├── Mecanisme de recherche : MCTS (explore l'arbre)
    │
    └── Evaluator (sous-agent)
        ├── Policy : "quels chemins sont prometteurs ?"
        └── Value : "est-on bien ou mal parti ?"
```

L'humain et le MCTS occupent chacun un `board.Agent`, mais ce sont deux agents agentiques tres differents. L'un reflechit, l'autre calcule. Le framework les traite de maniere identique.

## L'Evaluator comme sous-agent

L'agent MCTS est un **orchestrateur** : il ne "comprend" rien au probleme. Il explore systematiquement un arbre de possibilites en repetant selection, expansion, evaluation, retropropagation. L'intelligence vient de l'`Evaluator`, qui est un sous-agent specialise.

L'`Evaluator` d'AlphaZero est en realite **deux sous-agents fusionnes** dans un seul reseau de neurones a double tete :

| Sous-agent | Question | Sortie | Role dans le MCTS |
|------------|----------|--------|-------------------|
| **Policy** | "Par ou chercher ?" | Distribution de probabilites sur les coups | Guide l'exploration (priors dans PUCT) |
| **Value** | "Comment ca se presente ?" | Score dans [-1, 1] | Remplace le rollout aleatoire |

Dans AlphaZero, ces deux sous-agents partagent un reseau (pour des raisons de performance : les couches profondes sont mutualisees). Mais conceptuellement, rien n'empeche de les separer :

- Un **LLM** qui propose des pistes (policy) — "quelles sont les continuations les plus naturelles ?"
- Un **autre LLM** (ou le meme avec un prompt different) qui evalue des situations (value) — "ce raisonnement tient-il la route ?"
- Un **expert humain** qui donne son intuition (value) — "cette position me semble forte"
- Un **modele statistique** qui estime des probabilites (policy) — "historiquement, ce coup est joue 40% du temps"

L'interface `Evaluator` est le point d'injection : le MCTS fournit le mecanisme de recherche, l'Evaluator fournit l'intelligence. Changer d'Evaluator change l'intelligence sans toucher a la recherche.

## MCTS comme moteur de raisonnement

Le framework n'est pas specifique aux jeux. Si on pousse l'abstraction, il peut piloter un **systeme de raisonnement par recherche arborescente** — l'idee derriere les modeles de raisonnement modernes (Tree of Thoughts, recherche guidee par LLM).

### Le mapping

| Framework | Jeu de plateau | Raisonnement par LLM |
|-----------|---------------|---------------------|
| `State` | Position du plateau | Contexte partiel (prompt + raisonnement en cours) |
| `CurrentPlayer()` | Joueur actif | Le generateur (ou alternance generateur/critique) |
| `PossibleMoves()` | Coups legaux | Continuations possibles (prochaines etapes de raisonnement, appels d'outil, blocs de texte) |
| `Evaluate()` | Victoire/nul/en cours | La reponse est-elle complete et satisfaisante ? |
| `Evaluator.policy` | Priors sur les coups | "Quelles continuations ont le plus de chances de mener a une bonne reponse ?" |
| `Evaluator.value` | Estimation de la position | "Ce raisonnement partiel est-il sur la bonne voie ?" |

### Comment ca fonctionnerait

Au lieu de generer du texte token par token (de gauche a droite), on explore un **arbre de raisonnements** :

1. **Generer plusieurs continuations** — le LLM produit K completions alternatives (via sampling avec temperature). C'est `PossibleMoves()`.
2. **Evaluer chacune** — un second appel au LLM (ou un modele critique) estime la qualite du raisonnement partiel. C'est `Evaluator.value`.
3. **Choisir la plus prometteuse** — le MCTS utilise PUCT pour equilibrer exploration et exploitation.
4. **Approfondir** — on continue la generation depuis la branche choisie.
5. **Retenir le meilleur chemin** — apres N iterations, `SelectBestMove` retourne la continuation la plus visitee.

### Generateur vs critique : un jeu a deux agents

L'aspect "deux joueurs" prend un sens inattendu ici. On peut modeliser le raisonnement comme un jeu entre :

- **Agent 1 — le generateur** : propose des raisonnements, essaie de construire une reponse convaincante
- **Agent 2 — le critique** : challenge les raisonnements, cherche des failles, des incoherences

Le MCTS equilibre les deux perspectives. Le generateur explore largement (exploration) ; le critique force l'approfondissement des pistes solides (exploitation). Le resultat est un raisonnement qui a survecu a un examen contradictoire.

Le LLM lui-meme peut jouer tous les roles :

| Role dans le framework | Le LLM l'implemente comment |
|------------------------|----------------------------|
| `PossibleMoves()` | Generer K completions alternatives (sampling) |
| `Evaluator.policy` | Log-probabilites des tokens (quelle suite est naturelle ?) |
| `Evaluator.value` | Auto-evaluation ("ce raisonnement est-il coherent ?") |
| `Evaluate()` | Critere d'arret ("la reponse est-elle finale et correcte ?") |

### Ce que le framework apporte

Le MCTS n'apporte pas d'intelligence — il apporte de la **rigueur dans l'exploration**. Un LLM seul genere un raisonnement lineaire. Un LLM guide par MCTS explore un arbre de raisonnements, revient en arriere quand une branche est mauvaise, et converge vers le chemin le plus robuste. C'est la difference entre "penser tout haut" et "reflechir methodiquement".

## Resume

```
board.Agent             Un role (identifiant uint8)
Agent agentique         Un systeme qui occupe un role (humain, MCTS, LLM)
CurrentPlayer()         Quel role doit agir ?
PreviousPlayer()        Quel role a agi pour arriver ici ?
Result(agent)           Convention : ce role a gagne
Evaluator               Sous-agent qui fournit policy + value
MCTS                    Orchestrateur qui fournit le mecanisme de recherche
```

`board.Agent` est la brique la plus simple du framework : un identifiant que le MCTS manipule sans chercher a comprendre ce qu'il represente. L'intelligence est dans l'`Evaluator`. La recherche est dans le MCTS. Le domaine est dans l'implementation de `State`. Que le probleme soit un jeu de plateau, une negociation, ou un raisonnement par LLM, le mecanisme est le meme.
