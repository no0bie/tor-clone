// Module to calculate Diffie-Hellman

package main

import (
	"crypto/rand"
	"math/big"
)

func calc_dh(g, G string) (big.Int, big.Int, big.Int) {
	max_min, min := big.NewInt(10000-3), big.NewInt(3)

	big_generator := new(big.Int)
	big_generator.SetString(g, 10)

	big_group := new(big.Int)
	big_group.SetString(G, 10)

	b, _ := rand.Int(rand.Reader, max_min)
	b.Add(b, min)

	Y := new(big.Int)
	Y.Exp(big_generator, b, big_group) // (g^a) mod G

	return *Y, *b, *big_group
}
