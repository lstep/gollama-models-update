#!/bin/bash

# Get the list of installed models
models=$(ollama ls | awk 'NR>1 {print $1}')

# Loop through each model and pull the latest version
for model in $models
do
    echo "Pulling latest version of $model..."
    ollama pull "$model"
    echo "-------------------------"
done

echo "All models updated successfully!"
