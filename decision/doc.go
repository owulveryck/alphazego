// Package decision définit les interfaces génériques pour les problèmes de
// décision séquentielle à un ou plusieurs acteurs.
//
// L'abstraction centrale est [State], qui représente un instantané d'un
// problème de décision séquentiel. Les jeux de plateau (morpion, Go, échecs)
// en sont l'exemple naturel, mais les interfaces s'appliquent à tout problème
// modélisable comme un arbre de décisions : négociations, planification
// adversariale, diagnostics médicaux, composition créative, etc.
//
// Le nombre d'acteurs n'est pas fixé par le framework. L'interface [State]
// expose [State.PreviousActor] pour que le moteur MCTS détermine qui a
// effectué la dernière action sans connaître la logique de tour. Ainsi, un
// problème à 2 acteurs (morpion), à 1 acteur (planification) ou à N acteurs
// (négociation multilatérale) peut être résolu par le même moteur.
//
// # Types
//
// [ActorID] identifie un décideur (joueur, partie, acteur). C'est un type
// distinct basé sur int. La méthode [State.Evaluate] retourne directement un
// ActorID : le gagnant si le problème est résolu, [Undecided] s'il est en
// cours, ou [Stalemate] en cas d'impasse. Il n'y a pas de type Result
// séparé : le résultat EST l'identifiant du gagnant.
//
// Pour utiliser ce framework, implémenter [State] avec la logique
// spécifique au problème. Le sous-package board fournit des interfaces
// supplémentaires pour les jeux de plateau ([board.Boarder]), avec deux
// implémentations d'exemple : le morpion (deux acteurs) et le taquin
// (un seul acteur).
package decision
