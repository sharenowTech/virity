package main

import (
	"testing"
)

func TestGenerateID(t *testing.T) {
	agentID := generateID(int64(9))

	t.Logf(agentID)
	if agentID != "MIMM" {
		t.Errorf("Wrong AgentID")
	}
}

/*func TestGenerateUID(t *testing.T) {

	_, err := generateUID("myhost")
	if err != nil {
		t.Error(err)
	}
}*/
