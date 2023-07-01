package validation

import (
	artifacts "artifactflow.com/m/v2/cmd/artifacts"
	"fmt"
)

type Constraint interface {
	Evaluate(artifacts.Artifact, string) *ConstraintViolation
}

type ConstraintViolation struct {
	//Key      string
	Problems []string
}

// compile-time interface check
var _ error = ConstraintViolation{}

// Error implements error
func (cv ConstraintViolation) Error() string {
	return fmt.Sprintf("%v", cv.Problems)
}
