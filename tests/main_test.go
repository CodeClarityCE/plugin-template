package main

import (
	"log"
	"testing"

	plugin "github.com/CodeClarityCE/plugin-NAME/src"
	dbhelper "github.com/CodeClarityCE/utility-dbhelper/helper"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	knowledge, err := dbhelper.GetDatabase(dbhelper.Config.Database.Knowledge)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	out := plugin.Start(knowledge)

	// Assert the expected values
	assert.NotNil(t, out)
}

func BenchmarkCreate(b *testing.B) {
	knowledge, err := dbhelper.GetDatabase(dbhelper.Config.Database.Knowledge)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	out := plugin.Start(knowledge)

	// Assert the expected values
	assert.NotNil(b, out)
}
