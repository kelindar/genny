// Code generated with https://github.com/kelindar/genny DO NOT EDIT.
// Any changes will be lost if this file is regenerated.

package main

import "github.com/kelindar/genny/examples/user-defined-types/person"
import "github.com/kelindar/genny/examples/user-defined-types/pet"

type PairPersonDog struct {
	First  person.Person
	Second pet.Dog
}

func (p PairPersonDog) Left() person.Person {
	return p.First
}

func (p PairPersonDog) Right() pet.Dog {
	return p.Second
}
