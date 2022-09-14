package pact

import (
	"testing"
)

func TestExecPactPublish(t *testing.T) {
	t.Parallel()

	t.Run("success - default", func(t *testing.T){

	})

	t.Run("failure - pact file parsing", func(t *testing.T){
		
	})

	t.Run("failure - invalid naming of contract", func(t *testing.T){
		
	})

	t.Run("failure - publishing", func(t *testing.T){
		
	})

	t.Run("failure - reporting", func(t *testing.T){
		
	})
}

func TestPublishPact(t *testing.T) {
	t.Parallel()

	t.Run("success - publising", func(t *testing.T){

	})


	t.Run("success - already published", func(t *testing.T){

	})

	t.Run("failure - executable not found", func(t *testing.T){

	})

	t.Run("failure - publishing", func(t *testing.T){

	})
}

func TestEnforceNaming(t *testing.T) {
	t.Run("naming OK", func(t *testing.T){

	})

	t.Run("nconsumer name invalid", func(t *testing.T){

	})

	t.Run("provider name invalid", func(t *testing.T){

	})
}