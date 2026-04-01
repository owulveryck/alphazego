// Package decision definit les interfaces generiques pour les problemes de
// decision sequentielle a un ou plusieurs acteurs.
//
// L'abstraction centrale est [State], qui represente un instantane d'un
// probleme de decision sequentiel. Les jeux de plateau (morpion, Go, echecs)
// en sont l'exemple naturel, mais les interfaces s'appliquent a tout probleme
// modelisable comme un arbre de decisions : negociations, planification
// adversariale, diagnostics medicaux, composition creative, etc.
//
// Le nombre d'acteurs n'est pas fixe par le framework. L'interface [State]
// expose [State.PreviousActor] pour que le moteur MCTS determine qui a
// effectue la derniere action sans connaitre la logique de tour. Ainsi, un
// probleme a 2 acteurs (morpion), a 1 acteur (planification) ou a N acteurs
// (negociation multilaterale) peut etre resolu par le meme moteur.
//
// # Types
//
// [ActorID] identifie un decideur (joueur, partie, acteur). C'est un type
// distinct base sur int. La methode [State.Evaluate] retourne directement un
// ActorID : le gagnant si le probleme est resolu, [NoActor] si il est en
// cours, ou [DrawResult] en cas de match nul. Il n'y a pas de type Result
// separe : le resultat EST l'identifiant du gagnant.
//
// Pour utiliser ce framework, implementer [State] avec la logique
// specifique au probleme. Le sous-package board fournit des interfaces
// supplementaires pour les jeux de plateau ([board.Boarder]).
package decision
