package decision

// ActorID identifie un décideur dans le problème. Dans un jeu de plateau,
// c'est un joueur. Dans une négociation, c'est une partie. Dans un problème
// de planification, c'est un acteur.
//
// ActorID est aussi utilisé comme résultat de [State.Evaluate] : la valeur
// retournée est l'ActorID du gagnant, [Undecided] si le problème est en cours,
// ou [Stalemate] en cas de match nul.
type ActorID int

const (
	// Undecided indique que le problème est en cours : aucune issue n'a encore
	// été déterminée. Utilisé comme valeur de retour de [State.Evaluate] et
	// comme contenu d'une position vide sur un plateau.
	Undecided ActorID = 0
	// Stalemate indique que le problème est terminé sans gagnant (match nul,
	// impasse, absence de solution).
	Stalemate ActorID = -1
)

// State représente l'état d'un problème de décision séquentiel à un ou
// plusieurs acteurs.
//
// Tout problème implémentant cette interface peut être résolu par le moteur
// MCTS. Les jeux de plateau en sont un cas particulier : le morpion
// en est l'implémentation de référence pour deux acteurs.
//
// # Implémenter State
//
// Le struct d'état doit contenir au minimum deux informations :
//
//   - L'état du problème (plateau, configuration, etc.)
//   - L'acteur courant (qui doit agir)
//
// Astuce : utiliser un tableau fixe ([N]uint8) plutôt qu'un slice ([]uint8)
// pour le plateau simplifie [State.PossibleMoves] : l'affectation d'un
// tableau copie les données automatiquement, sans appel à copy().
//
// # Contrat
//
// Les invariants suivants doivent être respectés :
//
//   - [State.PossibleMoves] ne doit jamais modifier le récepteur. Chaque état
//     retourné doit être une copie indépendante (c'est le piège le plus courant).
//   - Chaque état retourné par [State.PossibleMoves] doit avoir
//     [State.CurrentActor] défini sur l'acteur suivant.
//   - [State.PossibleMoves] retourne un slice vide (ou nil) quand
//     [State.Evaluate] != [Undecided] : un état terminal n'a pas d'actions.
//   - [State.ID] doit être déterministe et inclure l'acteur courant.
//     Même configuration + acteur différent = ID différent.
//
// # Mapping vers d'autres domaines
//
//   - Jeu de plateau : CurrentActor = joueur, PossibleMoves = coups légaux
//   - Négociation : CurrentActor = partie, PossibleMoves = propositions/contre-offres
//   - Diagnostic : CurrentActor = optique (patient/médecin), PossibleMoves = examens/traitements
//   - Planification : CurrentActor = agent, PossibleMoves = actions possibles
type State interface {
	// CurrentActor retourne l'acteur dont c'est le tour d'agir.
	// Pour un problème à deux acteurs en alternance, l'acteur suivant est 3 - acteur.
	CurrentActor() ActorID

	// PreviousActor retourne l'acteur qui a effectué l'action menant à cet état.
	// Pour un problème à deux acteurs : 3 - CurrentActor().
	// Pour l'état initial, le comportement est défini par l'implémentation
	// (typiquement le dernier acteur dans l'ordre de jeu).
	// Le moteur MCTS utilise cette méthode pour créditer les victoires
	// au bon acteur lors de la rétropropagation.
	PreviousActor() ActorID

	// Evaluate retourne l'issue du problème :
	//   - [Undecided] (0) si le problème est en cours
	//   - [Stalemate] (-1) en cas de match nul
	//   - un ActorID positif si cet acteur a gagné
	//
	// Le moteur MCTS appelle cette méthode à chaque nœud pour détecter
	// les états terminaux. Un état où Evaluate != Undecided ne doit pas
	// avoir d'actions possibles (PossibleMoves retourne nil ou vide).
	Evaluate() ActorID

	// PossibleMoves retourne tous les états atteignables depuis l'état courant
	// en effectuant une action. C'est la méthode la plus importante à implémenter
	// correctement.
	//
	// Chaque état retourné doit :
	//   - être une copie indépendante (ne pas partager de données mutables avec le récepteur)
	//   - avoir CurrentActor() défini sur l'acteur suivant
	//
	// Si l'implémentation satisfait aussi [board.ActionRecorder], chaque état enfant
	// doit avoir LastAction() défini sur l'action qui a produit cet état.
	//
	// Retourne nil ou un slice vide si Evaluate() != [Undecided].
	//
	// Astuce : utiliser un tableau fixe ([N]uint8) au lieu d'un slice ([]uint8)
	// pour le plateau rend la copie automatique par simple affectation.
	PossibleMoves() []State

	// ID retourne un identifiant unique pour cet état. Deux états avec la même
	// configuration et le même acteur courant doivent retourner des ID identiques.
	// Inclure l'acteur courant dans l'ID : même configuration avec un acteur
	// différent constitue un état différent.
	ID() string
}

// RandomMover est une interface optionnelle qu'un [State] peut implémenter
// pour accélérer les rollouts aléatoires du moteur MCTS.
//
// # Motivation
//
// Durant la phase de simulation (rollout), le MCTS choisit un coup aléatoire
// à chaque étape en appelant [State.PossibleMoves], qui retourne TOUS les
// états successeurs. Or, un seul est utilisé — les autres sont immédiatement
// abandonnés. Sur un morpion avec 5 cases vides, cela signifie 5 allocations
// de struct + 1 slice pour n'en garder qu'une.
//
// RandomMover permet de court-circuiter ce schéma en générant directement
// un seul état successeur choisi aléatoirement, réduisant les allocations
// de N à 1 par étape de rollout.
//
// # Utilisation par le moteur MCTS
//
// Le moteur MCTS détecte cette interface par type assertion dans simulate().
// Si l'état implémente RandomMover, RandomMove est appelé à la place de
// PossibleMoves. Sinon, le comportement existant est conservé (fallback sur
// PossibleMoves + sélection aléatoire).
//
// # Contrat
//
//   - RandomMove ne doit jamais modifier le récepteur (même contrat que
//     [State.PossibleMoves]).
//   - L'état retourné doit être une copie indépendante avec
//     [State.CurrentActor] défini sur l'acteur suivant.
//   - Le paramètre rng est une fonction qui retourne un entier aléatoire
//     dans [0, n). L'implémentation doit l'utiliser (et non math/rand)
//     pour que le moteur MCTS puisse contrôler la reproductibilité via
//     sa graine.
//   - RandomMove ne doit être appelé que sur un état non terminal
//     ([State.Evaluate] == [Undecided]). Le comportement sur un état
//     terminal est indéfini.
//
// # Quand implémenter RandomMover
//
// Cette interface est bénéfique lorsque :
//
//   - L'état a un nombre significatif de coups possibles (>2)
//   - La création d'un état successeur est coûteuse en allocations
//   - Le MCTS pur (avec rollouts) est utilisé (pas le mode AlphaZero)
//
// En mode AlphaZero (avec evaluator), simulate() n'est jamais appelé,
// donc RandomMover n'a aucun effet.
//
// # Exemple
//
// Pour un morpion, RandomMover compte les cases vides, choisit un index
// aléatoire parmi elles, et crée un seul nouvel état :
//
//	func (t *TicTacToe) RandomMove(rng func(int) int) decision.State {
//	    count := 0
//	    for i := 0; i < BoardSize; i++ {
//	        if t.board[i] == 0 { count++ }
//	    }
//	    target := rng(count)
//	    for i := 0; i < BoardSize; i++ {
//	        if t.board[i] == 0 {
//	            if target == 0 {
//	                game := t.board
//	                game[i] = uint8(t.actorTurn)
//	                return &TicTacToe{board: game, actorTurn: 3 - t.actorTurn, lastAction: i}
//	            }
//	            target--
//	        }
//	    }
//	    return nil // unreachable
//	}
type RandomMover interface {
	// RandomMove retourne un unique état successeur choisi aléatoirement
	// parmi tous les coups possibles. Le paramètre rng(n) retourne un
	// entier uniforme dans [0, n).
	RandomMove(rng func(int) int) State
}
