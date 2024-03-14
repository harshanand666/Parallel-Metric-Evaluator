package scheduler

import "parallel_model_evaluator/metrics"

// Run the sequential version of the program
func RunSequential(config Config, records [][]string, thresholds []float32, optimalMetrics *metrics.OptimalMetrics) {

	thresholdScores := make(map[float32]metrics.Metrics)
	for _, threshold := range thresholds {
		curMetrics := &metrics.Metrics{}
		// For each threshold, calculate metrics for the entire file as a single chunk
		metrics.ChunkMetrics(&records, 0, len(records), curMetrics, threshold)
		if curMetrics.TotalP == 0 {
			curMetrics.Precision = 0
		} else {
			curMetrics.Precision = float32(curMetrics.CorrectP) / float32(curMetrics.TotalP)
		}
		curMetrics.Recall = float32(curMetrics.CorrectR) / float32(curMetrics.TotalR)
		thresholdScores[threshold] = *curMetrics
		metrics.UpdateOptimalMetrics(curMetrics, optimalMetrics, threshold)
	}
	// Generate output
	metrics.PrintMetrics(&thresholdScores, optimalMetrics)
	metrics.PRCurve(&thresholdScores, thresholds)
}
