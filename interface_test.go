package bgo

import (
	"bytes"
	"testing"
)

var testFileSignature = []byte{'T', 'E', 'S', 'T', 'V', '1', '.', '0'}
var failFileSignature = []byte{'F', 'A', 'I', 'L', 'O', 'M', 'G', '!'}

func TestSniff(t *testing.T) {
	RegisterFormat("test", "TESTV1.0", nil)

	testData := bytes.NewReader(testFileSignature)
	failData := bytes.NewReader(failFileSignature)

	testFormat := sniff(testData)
	t.Logf("Testformat: %v\n", testFormat)

	if testFormat.name != "test" {
		t.Errorf("sniff(TESTV1.0) failed to find test format")
	}

	failFormat := sniff(failData)
	t.Logf("Failformat: %v\n", failFormat)
	if failFormat.name != "" {
		t.Errorf("sniff(FAILOMG!) did not return an empty format")
	}
}
