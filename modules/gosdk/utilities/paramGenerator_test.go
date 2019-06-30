package utilities

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestParamGenerator(t *testing.T)  {
	fmt.Printf("Get literal parameter \n")
	literParam := []string{ "literal~~~invoke","literal~~~put" }
	args, err := GetComplexArgs(literParam,0)
	fmt.Printf("literal args: %s \n",args)
	assert.NoError(t, err)

	fmt.Printf("Get stringPattern parameter \n")
	spParam := []string{ "stringPattern~~~Account[0-9]+","stringPattern~~~Account[0-9]+" }
	args, err = GetComplexArgs(spParam,0)
	fmt.Printf("stringPattern args: %s \n",args)
	assert.NoError(t, err)

	fmt.Printf("Get sequentialString parameter \n")
	sequentialStringParam := []string{ "sequentialString~~~marbles"}
	args, err = GetComplexArgs(sequentialStringParam,0)
	fmt.Printf("sequentialString args: %s \n",args)
	assert.NoError(t, err)

	fmt.Printf("Get intergerRange parameter \n")
	integerRangeParam := []string{ "intRange~~~10~~~100","intRange~~~100~~~1000" }
	args, err = GetComplexArgs(integerRangeParam,0)
	fmt.Printf("integerRange args: %s \n",args)
	assert.NoError(t, err)

	fmt.Printf("Get payloadRanger parameter \n")
	payloadRangerParam := []string{ "payloadRange~~~10~~~11" }
	args, err = GetComplexArgs(payloadRangerParam,0)
	fmt.Printf("integerRange args: %d \n",[]byte(args[0]))
	assert.NoError(t, err)
}
