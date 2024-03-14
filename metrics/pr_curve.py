#!/usr/local/bin/python

import matplotlib.pyplot as plt

fig = plt.figure()
ax = fig.add_subplot(111)

thresholds = []
precisions = []
recalls = []

# Read data from text file
with open("../results/PrecisionRecall.txt") as f:
    txt = f.readlines()

# Extract and populate data from each line
for line in txt:
    data = line.strip().split(",")
    precision = float(data[1].split(": ")[1])
    recall = float(data[2].split(": ")[1])
    if precision == 0 and recall == 0:
        # No need to plot this entry
        continue
    thresholds.append(float(data[0].split(": ")[1]))
    precisions.append(precision)
    recalls.append(recall)

# Generate and save the plot
plt.plot(precisions, recalls)

plt.title("Precision - Recall Curve")
plt.xlabel("Precision")
plt.ylabel("Recall")
plt.savefig("../results/Precision-Recall Curve.png")
