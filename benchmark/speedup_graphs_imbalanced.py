import subprocess
import matplotlib.pyplot as plt
import csv

csv_filename = f"speedup_data_imbalanced.csv"  # Store data in CSV

THREADS = ["2", "4", "6", "8", "12"]
RUNS = 5
baselines = []

with open(csv_filename, "w", newline="") as csv_file:
    csv_writer = csv.writer(csv_file)
    csv_writer.writerow(["mode", "threads", "avg_time", "speedup"])

    for i in range(RUNS):
        op = (
            subprocess.check_output(
                ["go", "run", "editor.go", "imbalanced", "s"],
                cwd="../editor",
                text=True,
            )
            .strip()
            .split("\n")[-1]
        )
        baselines.append(float(op.strip()))
        baseline = sum(baselines) / RUNS

    modes = ["p-normal", "p-nosteal", "p-steal"]
    for mode in modes:
        speedups = []
        for thread in THREADS:
            times = []
            for i in range(RUNS):
                op = (
                    (
                        subprocess.check_output(
                            ["go", "run", "editor.go", "imbalanced", mode, thread],
                            cwd="../editor",
                            text=True,
                        )
                    )
                    .strip()
                    .split("\n")[-1]
                )
                time = float(op.strip())
                times.append(time)  # Min runtime of parallel file

            speedup = baseline / (sum(times) / RUNS)
            speedups.append(speedup)
            csv_writer.writerow([mode, thread, sum(times) / RUNS, speedup])

        plt.plot(THREADS, speedups, label=mode)  # Plot speedup graph

# Add labels and legend to the plot
plt.xlabel("Number of Threads")
plt.ylabel("Speedup")
plt.title(f"Speedup vs Threads - Imbalanced")
plt.legend()
plt.grid(True)

plt.savefig("./speedup-images-imbalanced.png")
