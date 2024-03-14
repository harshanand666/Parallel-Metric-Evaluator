package main

import (
	"fmt"
	"os"
	"parallel_model_evaluator/scheduler"
	"strconv"
	"time"
)

const usage = "Usage: editor data_type mode [number of threads] \n" +
	"data_type = (balanced) uses the dataset as is, (imbalanced) uses imbalanced dataset sorted by length. \n" +
	"mode = (s) run sequentially, (p-normal) run normal parallel implementation, (p-nosteal) run parallel with dequeues, (p-steal) run parallel with work stealing \n" +
	"[number of threads] = Runs the parallel version of the program with the specified number of threads."

// Create a config object based on user's input and run the scheduler
func main() {
	config := scheduler.Config{}
	if (len(os.Args) != 4) && os.Args[2] != "s" {
		fmt.Println(usage)
		return
	}
	config.Mode = os.Args[2]

	if config.Mode == "p-normal" || config.Mode == "p-steal" || config.Mode == "p-nosteal" {
		config.ThreadNum, _ = strconv.Atoi(os.Args[3])
	} else if config.Mode != "s" {
		fmt.Println(usage)
		return
	}

	if os.Args[1] == "balanced" {
		config.File = "../data/balanced.csv"
	} else if os.Args[1] == "imbalanced" {
		config.File = "../data/imbalanced.csv"
	} else {
		fmt.Println(usage)
		return
	}

	start := time.Now()
	scheduler.Schedule(config)
	end := time.Since(start).Seconds()
	fmt.Printf("%.2f\n", end)

}
