package mcts

import "math"

// logTableSize est la taille de la table de lookup pour math.Log.
// La plupart des nœuds ont moins de 1024 visites ; seule la racine
// dépasse pour les recherches à plus de 1024 itérations.
const logTableSize = 1025

// logTable contient les valeurs précalculées de math.Log(i) pour i ∈ [0, 1024].
// logTable[0] = -Inf (log(0)), jamais utilisé en pratique car visits ≥ 1.
var logTable [logTableSize]float64

// sqrtTable contient les valeurs précalculées de math.Sqrt(i) pour i ∈ [0, 1024].
var sqrtTable [logTableSize]float64

func init() {
	logTable[0] = math.Inf(-1)
	sqrtTable[0] = 0
	for i := 1; i < logTableSize; i++ {
		logTable[i] = math.Log(float64(i))
		sqrtTable[i] = math.Sqrt(float64(i))
	}
}

// fastLog retourne math.Log(v) en utilisant la table de lookup si v est
// un entier ≤ 1024, sinon calcule via math.Log.
func fastLog(v float64) float64 {
	iv := int(v)
	if iv < logTableSize {
		return logTable[iv]
	}
	return math.Log(v)
}

// fastSqrt retourne math.Sqrt(v) en utilisant la table de lookup si v est
// un entier ≤ 1024, sinon calcule via math.Sqrt.
func fastSqrt(v float64) float64 {
	iv := int(v)
	if iv < logTableSize {
		return sqrtTable[iv]
	}
	return math.Sqrt(v)
}
