package main

import (
	"math/rand"
	"time"
)

var generator *rand.Rand

func init() {
	generator = rand.New(rand.NewSource(time.Now().UnixNano()))
}

const alphabet = "abcdefghijklmnopqrstuvwxyz"

type Model struct {
	Name string
	Code string
}

type Machine struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Model Model
}

func InsertMachineDocument() Machine {
	var first, second, third, fourth, fifth, sixth, seventh, eighth, nineth, tenth int

	first = generator.Intn(26)
	second = generator.Intn(26)
	third = generator.Intn(26)
	fourth = generator.Intn(26)
	fifth = generator.Intn(26)
	sixth = generator.Intn(26)
	seventh = generator.Intn(26)
	eighth = generator.Intn(26)
	nineth = generator.Intn(26)
	tenth = generator.Intn(26)

	soon := []byte{alphabet[first], alphabet[second], alphabet[third], alphabet[fourth], alphabet[fifth], alphabet[sixth], alphabet[seventh], alphabet[eighth], alphabet[nineth], alphabet[tenth]}
	id := string(soon[:])
	pkey := string(soon[:3])

	machine := Machine{id, pkey, Model{"Netspresso", "13X"}}
	return machine
}

func RandIntn(n int) int {
	return generator.Intn(n)
}
