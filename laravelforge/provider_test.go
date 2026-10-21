package laravelforge

import "testing"

func TestProviderSchemaIsValid(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("provider schema is invalid: %s", err)
	}
}
