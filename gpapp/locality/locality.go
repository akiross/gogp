package main

import (
	"github.com/akiross/gogp"
	"github.com/akiross/gpapp"
)

// Locality test

func main() {
	// For each representation
	//   Build N individuals (e.g. 10000)
	//   For each individual i
	//     Mutate i M times (e.g. 50), yielding j_k
	//     Compute avg (over k) distance between i and j_k

	const (
		N = 10000
		M = 50
	)

	// Select mutation to use
	// Where do I get the functionals for each representation?
	// FIXME devo ri-impacchettare tutto in modo che i funzionali siano
	// esposti per ogni rappresentazione e che quindi io possa usarli qui?
	// Oppure li ricostruisco?
	// Forse conviene muovere il pacchetto locality fuori da gpapp?
	// Forse conviene fare un bel disegno per vedere le dipendenze?
	mutate = gogp.MakeTreeMutation()

	// Build random individuals
	randomIndividuals := make([]gpapp.Individual, N)
	// For each individual
	for _, i := range randomIndividuals {
		// Initialize it
		i.Initialize()
		// Copy the individual
		j := i.Copy()
		// Mutate the individual
	}
}
