package decision

// ActorID identifie un decideur dans le probleme. Dans un jeu de plateau,
// c'est un joueur. Dans une negociation, c'est une partie. Dans un probleme
// de planification, c'est un acteur.
//
// ActorID est aussi utilise comme resultat de [State.Evaluate] : la valeur
// retournee est l'ActorID du gagnant, [NoActor] si le probleme est en cours,
// ou [DrawResult] en cas de match nul.
type ActorID int

const (
	// NoActor indique qu'aucun acteur n'est concerne. Utilise comme valeur de
	// retour de [State.Evaluate] pour indiquer que le probleme est en cours,
	// et comme contenu d'une position vide sur un plateau.
	NoActor ActorID = 0
	// DrawResult indique un match nul : le probleme est termine mais aucun
	// acteur n'a gagne.
	DrawResult ActorID = -1
	// Actor1 est le premier acteur (typiquement X au morpion).
	Actor1 ActorID = 1
	// Actor2 est le second acteur (typiquement O au morpion).
	Actor2 ActorID = 2
)

// State represente l'etat d'un probleme de decision sequentiel a un ou
// plusieurs acteurs.
//
// Tout probleme implementant cette interface peut etre resolu par le moteur
// MCTS. Les jeux de plateau en sont un cas particulier : le morpion
// en est l'implementation de reference pour deux acteurs.
//
// # Implementer State
//
// Le struct d'etat doit contenir au minimum deux informations :
//
//   - L'etat du probleme (plateau, configuration, etc.)
//   - L'acteur courant (qui doit agir)
//
// Astuce : utiliser un tableau fixe ([N]uint8) plutot qu'un slice ([]uint8)
// pour le plateau simplifie [State.PossibleMoves] : l'affectation d'un
// tableau copie les donnees automatiquement, sans appel a copy().
//
// # Contrat
//
// Les invariants suivants doivent etre respectes :
//
//   - [State.PossibleMoves] ne doit jamais modifier le recepteur. Chaque etat
//     retourne doit etre une copie independante (c'est le piege le plus courant).
//   - Chaque etat retourne par [State.PossibleMoves] doit avoir
//     [State.CurrentActor] defini sur l'acteur suivant.
//   - [State.PossibleMoves] retourne un slice vide (ou nil) quand
//     [State.Evaluate] != [NoActor] : un etat terminal n'a pas d'actions.
//   - [State.ID] doit etre deterministe et inclure l'acteur courant.
//     Meme configuration + acteur different = ID different.
//
// # Mapping vers d'autres domaines
//
//   - Jeu de plateau : CurrentActor = joueur, PossibleMoves = coups legaux
//   - Negociation : CurrentActor = partie, PossibleMoves = propositions/contre-offres
//   - Diagnostic : CurrentActor = optique (patient/medecin), PossibleMoves = examens/traitements
//   - Planification : CurrentActor = agent, PossibleMoves = actions possibles
type State interface {
	// CurrentActor retourne l'acteur dont c'est le tour d'agir.
	// Pour un probleme a deux acteurs en alternance, l'acteur suivant est 3 - acteur.
	CurrentActor() ActorID

	// PreviousActor retourne l'acteur qui a effectue l'action menant a cet etat.
	// Pour un probleme a deux acteurs : 3 - CurrentActor().
	// Pour l'etat initial, le comportement est defini par l'implementation
	// (typiquement le dernier acteur dans l'ordre de jeu).
	// Le moteur MCTS utilise cette methode pour crediter les victoires
	// au bon acteur lors de la retropropagation.
	PreviousActor() ActorID

	// Evaluate retourne l'issue du probleme :
	//   - [NoActor] (0) si le probleme est en cours
	//   - [DrawResult] (-1) en cas de match nul
	//   - un ActorID positif si cet acteur a gagne
	//
	// Le moteur MCTS appelle cette methode a chaque noeud pour detecter
	// les etats terminaux. Un etat ou Evaluate != NoActor ne doit pas
	// avoir d'actions possibles (PossibleMoves retourne nil ou vide).
	Evaluate() ActorID

	// PossibleMoves retourne tous les etats atteignables depuis l'etat courant
	// en effectuant une action. C'est la methode la plus importante a implementer
	// correctement.
	//
	// Chaque etat retourne doit :
	//   - etre une copie independante (ne pas partager de donnees mutables avec le recepteur)
	//   - avoir CurrentActor() defini sur l'acteur suivant
	//
	// Si l'implementation satisfait aussi [board.ActionRecorder], chaque etat enfant
	// doit avoir LastAction() defini sur l'action qui a produit cet etat.
	//
	// Retourne nil ou un slice vide si Evaluate() != [NoActor].
	//
	// Astuce : utiliser un tableau fixe ([N]uint8) au lieu d'un slice ([]uint8)
	// pour le plateau rend la copie automatique par simple affectation.
	PossibleMoves() []State

	// ID retourne un identifiant unique pour cet etat. Deux etats avec la meme
	// configuration et le meme acteur courant doivent retourner des ID identiques.
	// Inclure l'acteur courant dans l'ID : meme configuration avec un acteur
	// different constitue un etat different.
	ID() string
}
