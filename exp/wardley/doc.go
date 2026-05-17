// Package wardley implémente l'exploration stratégique de cartes Wardley via MCTS.
//
// Il combine le moteur MCTS d'alphazego avec le modèle de cartes Wardley de
// wardleyToGo pour explorer l'espace des décisions stratégiques : évolutions
// de composants et applications de gameplays.
//
// Un LLM (Gemini via Vertex AI) évalue la qualité stratégique de chaque état
// de la carte, guidant la recherche MCTS vers les séquences de moves optimales.
//
// L'outil traite la carte Wardley comme un puzzle mono-acteur : le décideur
// unique explore les moves possibles, et le MCTS sélectionne la meilleure
// séquence via AlphaMCTS (mode AlphaZero avec évaluation LLM).
package wardley
