package decision

// ActorID identifie un décideur dans le problème. Dans un jeu de plateau,
// c'est un joueur. Dans une négociation, c'est une partie. Dans un problème
// de planification, c'est un acteur.
//
// ActorID est aussi utilisé comme résultat de [State.Evaluate] : la valeur
// retournée est l'ActorID du gagnant, [NoActor] si le problème est en cours,
// ou [DrawResult] en cas de match nul.
type ActorID int

const (
	// NoActor indique qu'aucun acteur n'est concerné. Utilisé comme valeur de
	// retour de [State.Evaluate] pour indiquer que le problème est en cours,
	// et comme contenu d'une position vide sur un plateau.
	NoActor ActorID = 0
	// DrawResult indique un match nul : le problème est terminé mais aucun
	// acteur n'a gagné.
	DrawResult ActorID = -1
	// Actor1 est le premier acteur (typiquement X au morpion).
	Actor1 ActorID = 1
	// Actor2 est le second acteur (typiquement O au morpion).
	Actor2 ActorID = 2
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
//     [State.Evaluate] != [NoActor] : un état terminal n'a pas d'actions.
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
	//   - [NoActor] (0) si le problème est en cours
	//   - [DrawResult] (-1) en cas de match nul
	//   - un ActorID positif si cet acteur a gagné
	//
	// Le moteur MCTS appelle cette méthode à chaque nœud pour détecter
	// les états terminaux. Un état où Evaluate != NoActor ne doit pas
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
	// Retourne nil ou un slice vide si Evaluate() != [NoActor].
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
