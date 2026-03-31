# Qu'est-ce qu'un PlayerID ?

## L'intuition

Un `PlayerID` identifie un **decideur** : une entite qui prend des decisions dans un probleme sequentiel. Le mot "joueur" est utilise par convention, mais le concept est plus large. Selon le domaine, il designe des choses tres differentes :

| Domaine | PlayerID designe | Exemple concret |
|---------|-----------------|-----------------|
| Jeu de plateau | Un joueur | Joueur X au morpion |
| Negociation | Une partie | L'acheteur, le vendeur |
| Diagnostic medical | Un decideur | Le medecin, le patient |
| Planification | Un acteur | Le planificateur |
| Verification | Un role | Le systeme, l'attaquant |

Ce que tous ces cas ont en commun : a chaque etape, **un decideur agit**, puis le probleme passe a l'etape suivante.

## PlayerID dans le code

Dans le framework, `PlayerID` est un type distinct base sur `int` :

```go
type PlayerID int
```

Ce choix est delibere :

- **Type distinct** : empeche les confusions avec d'autres `int` du code (le compilateur refuse les melanges)
- **Valeurs negatives possibles** : `DrawResult = -1` ne peut jamais entrer en collision avec un identifiant de joueur positif
- **Pas de type Result separe** : `Evaluate()` retourne un `PlayerID` — le gagnant, ou `NoPlayer`/`DrawResult`

Les constantes predefinies :

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

C'est tout. Le MCTS ne connait pas le nombre de joueurs, ne sait pas comment ils alternent, et ne sait pas ce que leurs identifiants signifient. Il utilise ces deux informations pour une seule chose : **crediter les victoires au bon joueur** lors de la retropropagation.

```
Noeud : CurrentPlayer = 2, PreviousPlayer = 1
        → Le joueur 1 a joue le coup menant ici.
        → Si Evaluate() retourne Player1,
          ce noeud recoit +1 win.
```

## La convention Evaluate = PlayerID du gagnant

Le framework utilise une convention simple pour relier joueurs et resultats :

> **Quand un joueur gagne, `Evaluate()` retourne son `PlayerID`.**

C'est ce que verifie la retropropagation :

```go
if result == n.state.PreviousPlayer() {
    n.wins += 1   // le joueur qui a joue ici a gagne
}
```

Cette convention fonctionne pour n'importe quel nombre de joueurs. Que le jeu ait 1, 2 ou 10 joueurs, la comparaison `result == playerID` reste valide. Et comme `DrawResult = -1`, il n'y a jamais de collision avec un identifiant de joueur.

## Un joueur, deux joueurs, N joueurs

### 2 joueurs (cas de reference)

Le morpion est le cas classique. Deux joueurs alternent : quand l'un joue, c'est au tour de l'autre. La logique de tour est triviale : `PreviousPlayer = 3 - CurrentPlayer` (puisque `3 - 1 = 2` et `3 - 2 = 1`).

Dans ce cas, le jeu est **adversarial** : ce qui est bon pour un joueur est mauvais pour l'autre. Le MCTS exploite cette propriete dans `BackpropagateValue` (chemin AlphaZero) en inversant le signe de la valeur a chaque niveau de l'arbre.

### 1 joueur

Un probleme a un seul joueur (planification, optimisation) est modelisable : l'unique joueur explore un arbre de decisions. `CurrentPlayer()` et `PreviousPlayer()` retournent toujours le meme joueur. Le MCTS explore l'arbre normalement, et la retropropagation credite toujours le meme joueur.

### N joueurs

Pour N > 2, chaque implementation de `State` definit sa propre logique de tour via `PreviousPlayer()`. Le chemin MCTS pur (rollouts + `Backpropagate` discret) fonctionne directement. Le chemin AlphaZero (`BackpropagateValue` avec alternance de signe) est limite a 2 joueurs pour l'instant.

## Pourquoi pas une interface Player plus riche ?

On pourrait imaginer un type `Player` avec des methodes (nom, couleur, strategie...). Le choix d'un simple `int` nomme est delibere :

- **Pas de couplage** : le MCTS ne depend d'aucune structure de joueur
- **Performance** : les comparaisons `result == playerID` sont des comparaisons d'entiers
- **Flexibilite** : n'importe quel probleme peut attribuer des identifiants comme il veut
- **Simplicite** : moins de code, moins de surface d'API, moins de bugs

La richesse semantique (nom du joueur, couleur, strategie) est dans l'implementation de `State`, pas dans le type `PlayerID` lui-meme.

## `board.PlayerID` vs agent agentique

Le mot "agent" a deux sens qui se superposent dans les projets de recherche en IA :

| | `board.PlayerID` | Agent agentique |
|---|---|---|
| **Nature** | Un identifiant (`int`) | Un systeme autonome |
| **Question** | "A qui le tour ?" | "Qui decide ?" |
| **Exemples** | `Player1 = 1`, `Player2 = 2` | Un humain, un MCTS, un LLM |

Un `PlayerID` est un **role** dans le probleme. Un agent agentique est le **systeme** qui occupe ce role. Le framework ne fait aucune distinction : il voit des roles qui alternent, independamment de ce qui se passe "dans la tete" de chaque role.

Au morpion par exemple :

```
Partie de morpion
│
├── Role 1 (PlayerID = 1) ← occupe par : Humain
│   └── decide seul (intuition, reflexion)
│
└── Role 2 (PlayerID = 2) ← occupe par : Agent MCTS
    │
    ├── Mecanisme de recherche : MCTS (explore l'arbre)
    │
    └── Evaluator (sous-agent)
        ├── Policy : "quels chemins sont prometteurs ?"
        └── Value : "est-on bien ou mal parti ?"
```

L'humain et le MCTS occupent chacun un `PlayerID`, mais ce sont deux agents agentiques tres differents. L'un reflechit, l'autre calcule. Le framework les traite de maniere identique.

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
PlayerID                Un role (identifiant int distinct)
NoPlayer (0)            Jeu en cours / case vide
DrawResult (-1)         Match nul (ne collisionne jamais avec un joueur)
Player1 (1), Player2(2) Joueurs predefinis pour les jeux a 2 joueurs
CurrentPlayer()         Quel role doit agir ?
PreviousPlayer()        Quel role a agi pour arriver ici ?
Evaluate() == PlayerID  Ce role a gagne
Evaluator               Sous-agent qui fournit policy + value
MCTS                    Orchestrateur qui fournit le mecanisme de recherche
```

`PlayerID` est la brique la plus simple du framework : un identifiant que le MCTS manipule sans chercher a comprendre ce qu'il represente. L'intelligence est dans l'`Evaluator`. La recherche est dans le MCTS. Le domaine est dans l'implementation de `State`. Que le probleme soit un jeu de plateau, une negociation, ou un raisonnement par LLM, le mecanisme est le meme.
