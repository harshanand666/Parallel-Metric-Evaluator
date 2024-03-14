package metrics

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/grassmudhorses/vader-go/lexicon"
	"github.com/grassmudhorses/vader-go/sentitext"
)

// Stores metrics for a particular run
type Metrics struct {
	CorrectP  int
	CorrectR  int
	TotalP    int
	TotalR    int
	Precision float32
	Recall    float32
}

// Stores optimal metrics along with their thresholds
type OptimalMetrics struct {
	MaxPrec    float32
	MaxRec     float32
	PrecThresh float32
	RecThresh  float32
}

// Returns a pointer to a new OptimalMetrics object
func InitOptimalMetrics() *OptimalMetrics {
	return &OptimalMetrics{MaxPrec: 0, MaxRec: 0, PrecThresh: 0, RecThresh: 0}
}

// Calculates and updates metrics based on the label and score of the input
func CalculateMetrics(score float32, label string, threshold float32, metrics *Metrics) {
	if score >= threshold {
		metrics.TotalP++
		if label == "0" {
			metrics.CorrectP++
		}
	}
	if label == "0" {
		metrics.TotalR++
		if score >= threshold {
			metrics.CorrectR++
		}
	}
}

// Processes a specified chunk of records and updates the metrics
func ChunkMetrics(records *[][]string, chunkStart int, chunkEnd int, curMetrics *Metrics, threshold float32) {
	for i := chunkStart; i < chunkEnd; i++ {
		record := (*records)[i]
		text, label := record[1], record[0]
		parsedtext := sentitext.Parse(text, lexicon.DefaultLexicon)
		negativeScore := float32(sentitext.PolarityScore(parsedtext).Negative)
		CalculateMetrics(negativeScore, label, threshold, curMetrics)
	}
}

// Updates OptimalMetrics based on curMetrics values
func UpdateOptimalMetrics(curMetrics *Metrics, optimalMetrics *OptimalMetrics, threshold float32) {
	if curMetrics.Precision > optimalMetrics.MaxPrec {
		optimalMetrics.MaxPrec = curMetrics.Precision
		optimalMetrics.PrecThresh = threshold
	}
	if curMetrics.Recall > optimalMetrics.MaxRec {
		optimalMetrics.MaxRec = curMetrics.Recall
		optimalMetrics.RecThresh = threshold
	}
}

// Prints final metrics to the console
func PrintMetrics(thresholdScores *map[float32]Metrics, optimalMetrics *OptimalMetrics) {
	fmt.Println("Max Precision = ", optimalMetrics.MaxPrec, "at threshold ", optimalMetrics.PrecThresh)
	fmt.Println("Max Recall = ", optimalMetrics.MaxRec, "at threshold ", optimalMetrics.RecThresh)
	fmt.Println(*thresholdScores)
}

// Writes the final results into a text file, and launches a python script to plot the P-R curve
func PRCurve(thresholdScores *map[float32]Metrics, thresholds []float32) {

	// Open the file for writing
	file, err := os.Create("../results/PrecisionRecall.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Iterate over thresholds and write data to file
	for _, threshold := range thresholds {
		metric := (*thresholdScores)[threshold]
		_, err := fmt.Fprintf(file, "Threshold: %f, Precision: %f, Recall: %f\n", threshold, metric.Precision, metric.Recall)
		if err != nil {
			panic(err)
		}
	}

	// Run python script to generate the P-R curve
	// https://pkg.go.dev/os/exec#Cmd.Run
	exec.Command("python", "../metrics/pr_curve.py")
}
