package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestGetLines(t *testing.T) {

	multilineString := `TaskID
	9e4f495a-a387-4aa1-a1e3-ff2718da0ccf
	LastEdited
	2022-06-30T09:06:00Z`

	stringList := splitMultiLineString(multilineString)
	log.Println(stringList)
	for i, str := range stringList {
		log.Println(fmt.Sprintf("INDEX: %v\n%v", i, str))
	}
	assert.True(t, len(stringList) == 4)
	assert.Equal(t, "2022-06-30T09:06:00Z", stringList[len(stringList)-1])

}
