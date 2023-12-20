package rendezvous_test

import (
	"fmt"
	"os"
	"runtime/metrics"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	exitCode := m.Run()

	// Dump godebug metrics to investigate https://github.com/golang/go/issues/64649
	descs := metrics.All()
	var godebugMetrics []metrics.Sample
	for i := range descs {
		if strings.HasPrefix(descs[i].Name, "/godebug/") && strings.HasSuffix(descs[i].Name, ":events") {
			godebugMetrics = append(godebugMetrics, metrics.Sample{Name: descs[i].Name})
		}
	}
	if len(godebugMetrics) > 0 {
		metrics.Read(godebugMetrics)
		for i := range godebugMetrics {
			if godebugMetrics[i].Value.Uint64() > 0 {
				fmt.Printf("%s: %v\n", godebugMetrics[i].Name, godebugMetrics[i].Value.Uint64())
			}
		}
	}

	os.Exit(exitCode)
}