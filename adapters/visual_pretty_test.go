package adapters_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/joergreinhardt/gatwd/adapters"
	"github.com/joergreinhardt/gatwd/dsltypes"
	"github.com/joergreinhardt/gatwd/visualizer"
)

func TestVisualPretty_ScannerStream(t *testing.T) {
	logData := "INFO service_started\nWARN high_memory_usage\nERROR connection_reset"
	scanner := bufio.NewScanner(strings.NewReader(logData))
	flag := dsltypes.BitFlag(1)

	seq := adapters.ScannerToSeq(scanner, flag)

	type LogRecord struct {
		Level   string
		Message string
	}

	opts := visualizer.ParseOptions("--curved --border")
	t.Log("=== PRETTY-PRINTED STREAM ADAPTER LOGS ===")
	for node := range seq {
		val, _ := node.Value.UnwrapRight()
		line := string(val.([]byte))
		parts := strings.SplitN(line, " ", 2)
		rec := LogRecord{Level: parts[0], Message: parts[1]}
		card := visualizer.Format(rec, opts)
		t.Logf("\n%s", card)
	}
}
