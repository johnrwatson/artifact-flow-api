package validation

import (
	artifacts "artifactflow.com/m/v2/cmd/artifacts"
	"fmt"
	//"log"
	"strings"
)

//type Limit struct {
//	Type  string       `json:"type"`
//	Value *interface{} `json:"value,omitempty"`
//}

// compile-time interface check
var _ Constraint = RuleLimit{}

//func (lim RuleLimit) String() string {
//	return fmt.Sprintf("type: %s, value: %s", lim.Type, lim.Value)
//}

// Evaluate implements Constraint
func (lim RuleLimit) Evaluate(artifact artifacts.Artifact, ruleKey string) *ConstraintViolation {
	switch lim.Type {
	case "min", "max", "equal", "set":
		return lim.checkLimitValues(artifact, ruleKey)
	default:
		return &ConstraintViolation{
			//Key:      ruleKey,
			Problems: []string{fmt.Sprintf("Unsupported Limit Type, set to %T with value %d, supported values are one of equal|min|max|set", lim.Type, *lim.Value)},
		}
	}
}

func (lim RuleLimit) checkLimitValues(artifact artifacts.Artifact, ruleKey string) (err *ConstraintViolation) {
	fmt.Println("Info: checking contents", artifact, "for limit", ruleKey, "of type", lim.Type, "with max", lim.Value)
	problems := make([]string, 0, 1)
	defer func() {
		if len(problems) == 0 {
			err = nil
		}
		err = &ConstraintViolation{
			//Key:      ruleKey,
			Problems: problems,
		}
	}()

	keys := strings.Split(ruleKey, ".")

	//fmt.Println("Checking where the keys are ", keys, "within", artifact, "for", ruleKey)
	// Is the key within the infinitely nested map object?

	value, ok := lookup(artifact, artifact.ArtifactMetadata, keys)
	if !ok {
		problems = append(problems, fmt.Sprintf("specified key not found: %v", keys))
		return
	}

	// Variable Setup to Handle Cases
	var intArtifactValue int
	var strArtifactValue string
	var intLimitValue int
	var strLimitValue string

	switch artifactValue := (value).(type) {
	case string:
		strArtifactValue = artifactValue
	case int, int32, float64:
		//fmt.Println("Artifact value found to be:", artifactValue)
		if n, ok := artifactValue.(int); ok {
			//fmt.Println("Setting intArtifactValue from int", n)
			intArtifactValue = n
		} else if f64, ok := artifactValue.(float64); ok {
			//fmt.Println("Setting intArtifactValue from float64", f64)
			intArtifactValue = int(f64)
		} else if i32, ok := artifactValue.(int32); ok {
			//fmt.Println("Setting intArtifactValue from int32", i32)
			intArtifactValue = int(i32)
		} else {
			problems = append(problems, fmt.Sprintf("artifact value is not an acceptable int: %v", artifactValue))
			return
		}
	default:
		problems = append(problems, fmt.Sprintf("Unsupported Artifact Value type, set to %T with value %d", artifactValue, artifactValue))
		return
	}

	//fmt.Println("limValue is set to", *lim.Value)
	// Need something in here that checks the two variables are of the same type, otherwise they can't be compared
	switch limValue := (*lim.Value).(type) {
	case string:
		strLimitValue = limValue
		if lim.Type == "equal" && strArtifactValue != strLimitValue {
			problems = append(problems, fmt.Sprintf("%s is not equal to %s", strArtifactValue, strLimitValue))
		}
	case int, int32, float64:
		//fmt.Println("Limit value found to be:", limValue)
		if n, ok := limValue.(int); ok {
			//fmt.Println("Setting intLimValue from int", n)
			intLimitValue = n
		} else if f64, ok := limValue.(float64); ok {
			//fmt.Println("Setting intLimValue from float64", f64)
			intLimitValue = int(f64)
		} else if i32, ok := limValue.(int32); ok {
			//fmt.Println("Setting intLimValue from int32", i32)
			intLimitValue = int(i32)
		} else {
			problems = append(problems, fmt.Sprintf("value is not an int: %v", limValue))
			return
		}
		// limit     record
		//fmt.Println(intLimitValue, intArtifactValue)
		//log.Printf("The type of intLimValue is %T with limit type %d", intLimitValue, lim.Type)
		//log.Printf("The type of intArtifactValue is %T", intArtifactValue)
		// min / max / equal / set (set doesn't need a case, would've errored earlier if it didn't exist)

		if lim.Type == "equal" && intArtifactValue != intLimitValue {
			problems = append(problems, fmt.Sprintf("%d is not equal to %d", intArtifactValue, intLimitValue))
		} else if lim.Type == "min" && lim.Value != nil && intArtifactValue < intLimitValue {
			problems = append(problems, fmt.Sprintf("%d is less than %d", intArtifactValue, intLimitValue))
		} else if lim.Type == "max" && lim.Value != nil && intArtifactValue > intLimitValue {
			problems = append(problems, fmt.Sprintf("%d is greater than %d", intArtifactValue, intLimitValue))
		} 

	default:
		problems = append(problems, fmt.Sprintf("Unsupported Limit Value type, set to %T with value %d", *lim.Value, *lim.Value))
		return
	}

	return
}