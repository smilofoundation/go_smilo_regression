package go_smilo_regression

import (
	"sync"

	"go-smilo/src/blockchain/regression/src/container"
)

func WaitFor(geths []container.Ethereum, waitFn func(eth container.Ethereum, wg *sync.WaitGroup)) {
	wg := new(sync.WaitGroup)
	for _, g := range geths {
		wg.Add(1)
		go waitFn(g, wg)
	}
	wg.Wait()
}
