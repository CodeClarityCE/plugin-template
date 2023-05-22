package template

import (
	"testing"

	"codeclarity.io/template/test"
)

// func TestReceiveSymfony(t *testing.T) {
// 	var tests = []struct {
// 		value string
// 		want  error
// 	}{
// 		{"symfony_dispatcher", nil},
// 		{"sbom_dispatcher", nil},
// 	}
// 	for _, tt := range tests {
// 		testname := fmt.Sprintf(tt.value)
// 		t.Run(testname, func(t *testing.T) {
// 			// START TEST
// 			_, msg, err := openConnection(tt.value)
// 			if err != nil {
// 				t.Errorf(msg)
// 			}
// 			// END TEST
// 		})
// 	}
// }

func TestScenario1(t *testing.T) {
	connection := "symfony_dispatcher"

	// Send mock data from scenario 1
	err := test.Scenario1(connection)
	if err != nil {
		// `t.Error*` will report test failures but continue
		// executing the test. `t.Fatal*` will report test
		// failures and stop the test immediately.
		t.Errorf(err.Error())
	}

	// Listen for messages
	// d := test.ReceiveMessage(connection)

	// Test the dispatcher for this message
	// dispatch(connection, d)
}
