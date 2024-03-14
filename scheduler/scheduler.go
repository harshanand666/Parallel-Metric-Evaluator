package scheduler

import (
	"encoding/csv"
	"os"
	"parallel_model_evaluator/metrics"
)

type Config struct {
	Mode      string
	ThreadNum int
	File      string
}

// Reads the input file and returns a slice of rows
func readAllCsv(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}
	return records
}

// Runs the correct version based on the config
func Schedule(config Config) {

	records := readAllCsv(config.File)

	// Thresholds to calculate precision and recall for
	thresholds := []float32{0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.55, 0.6, 0.65, 0.7, 0.75}
	optimalMetrics := metrics.InitOptimalMetrics()

	if config.Mode == "s" {
		RunSequential(config, records, thresholds, optimalMetrics)
	} else {
		if config.Mode == "p-normal" {
			RunParallelNormal(config, records, thresholds, optimalMetrics)
		} else if config.Mode == "p-nosteal" {
			RunParallelSteal(config, records, thresholds, optimalMetrics, false)
		} else if config.Mode == "p-steal" {
			RunParallelSteal(config, records, thresholds, optimalMetrics, true)
		}
	}
}
