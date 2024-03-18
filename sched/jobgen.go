package sched

import "fmt"

type jobGenerations map[string]int

var jobsetGens jobGenerations = make(jobGenerations)

func (jg jobGenerations) currentGen(category, jobset string) string {
	gen := jg[jobset]
	return fmt.Sprintf("//%s/%s##%d", category, jobset, gen)
}

func (jg jobGenerations) nextGen(category, jobset string) string {
	gen := jg[jobset]
	gen += 1
	jg[jobset] = gen
	return fmt.Sprintf("//%s/%s##%d", category, jobset, gen)
}
