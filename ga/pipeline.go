package ga

import (
	"sync"
)

var sticacci string = `
TODO quindi facciamo così, che mi sembra un buon compromesso:
questi generatori sono gli utilizzatori di livello più alto delle funzioni di mutazione e crossover
quindi usiamo queste qui:

oltre a ritornare un channel pieno di individui, forniamo i dati contenenti le fitness:
possiamo usare dei chan Fitness come secondo return value, che contengono la differenza tra
la fitness dell'individuo PRIMA - DOPO, oppure ritornare direttamente la variazione relativa (DOPO/PRIMA)
così possiamo calcolare la fitness di prima partendo da quella dopo.
Avendo un canale in uscita, dovremmo essere tranquilli per la concorrenza, permettendo di scalare comunque
senza pericoli di sezioni critiche. Infone, ritornando dei valori, non dobbiamo nemmeno preoccuparci delle dipendenze
da altre parti di codice: lavoriamo direttamente con le fitness (che fose non possiamo calcolare usando ga.Individual,
ma basta includere Fitness() nell'interfaccia), così teniamo facilmente le statistiche
`

type PipelineIndividual struct {
	Ind              Individual
	InitialFitness   Fitness
	CrossoverFitness Fitness
	MutationFitness  Fitness
}

// This is a version of Select that is a stage in a pipeline. Will provide pointers to NEW individuals
func GenSelect(pop Population, selectionSize int, generation float32) <-chan PipelineIndividual {
	// A channel for output individuals
	out := make(chan PipelineIndividual, selectionSize)
	go func() {
		sel, _ := pop.Select(selectionSize, generation)
		for i := range sel {
			var ind PipelineIndividual
			ind.Ind = sel[i]
			ind.InitialFitness = sel[i].Evaluate()
			out <- ind
		}
		/*
			for i := 0; i < selectionSize; i++ {
				// Pick an initial (pointer to) random individual
				best := pop.Get(rand.Intn(pop.Size()))
				// Select other players and select the best
				for j := 1; j < tournSize; j++ {
					maybe := pop.Get(rand.Intn(pop.Size()))
					if pop.IndividualCompare(maybe, best) {
						best = maybe
					}
				}
				sel := best.Copy() //&Individual{best.node.Copy(), best.fitness, best.fitIsValid}
				out <- sel
			}*/
		close(out)
	}()
	return out
}

func GenCrossover(in <-chan PipelineIndividual, pCross float64) <-chan PipelineIndividual {
	out := make(chan PipelineIndividual)
	go func() {
		// Continue forever
		for {
			// Take one item and if we got a real item
			i1, ok := <-in
			if ok {
				i2, ok := <-in
				if ok {
					// We got two items! Crossover
					i1.Ind.Crossover(pCross, i2.Ind)
					i1.CrossoverFitness = i1.Ind.Evaluate()
					i2.CrossoverFitness = i2.Ind.Evaluate()
					out <- i1
					out <- i2
				} else {
					// Can't crossover a single item
					out <- i1
				}
			} else {
				close(out)
				break
			}
		}
	}()
	return out
}

func GenMutate(in <-chan PipelineIndividual, pMut float64) <-chan PipelineIndividual {
	out := make(chan PipelineIndividual)
	go func() {
		for ind := range in {
			ind.Ind.Mutate(pMut)
			ind.MutationFitness = ind.Ind.Evaluate()
			out <- ind
		}
		close(out)
	}()
	return out
}

// Fan-in channels of individuals
func FanIn(in ...<-chan PipelineIndividual) <-chan PipelineIndividual {
	out := make(chan PipelineIndividual)
	// Used to wait that a set of goroutines finish
	var wg sync.WaitGroup
	// Function that copy from one channel to out
	emitter := func(c <-chan PipelineIndividual) {
		for i := range c {
			out <- i
		}
		wg.Done() // Signal the group
	}
	// How many goroutines to wait for
	wg.Add(len(in))
	// Start the routines
	for _, c := range in {
		go emitter(c)
	}

	// Wait for the emitters to finish, then close
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Collects all the individuals from a channel and return a slice. Size is a hint for performances
func Collector(in <-chan PipelineIndividual, size int) []PipelineIndividual {
	pop := make([]PipelineIndividual, 0, size)
	for ind := range in {
		pop = append(pop, ind)
	}
	return pop
}
