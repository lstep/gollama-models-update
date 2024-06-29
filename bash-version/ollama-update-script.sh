#!/bin/bash

# Get the list of installed models
models=$(ollama ls | awk 'NR>1 {print $1}')

# Initialize an array to store failed models
failed_models=()

# Function to format time
format_time() {
    local seconds=$1
    local minutes=$((seconds / 60))
    local remaining_seconds=$((seconds % 60))
    printf "%02d:%02d" $minutes $remaining_seconds
}

# Loop through each model and pull the latest version
for model in $models
do
    echo "Pulling latest version of $model..."
    start_time=$(date +%s)
    output=$(ollama pull "$model" 2>&1)
    exit_status=$?
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    formatted_duration=$(format_time $duration)
    
    if echo "$output" | grep -q "Error: pull model manifest: file does not exist" || [ $exit_status -ne 0 ]; then
        echo "Failed to update $model"
        failed_models+=("$model")
    else
        echo "Update of $model completed in $formatted_duration"
    fi
    echo "-------------------------"
done

echo "All model update attempts completed."

# Display the list of failed models
if [ ${#failed_models[@]} -eq 0 ]; then
    echo "All models updated successfully!"
else
    echo "The following models failed to update:"
    for model in "${failed_models[@]}"
    do
        echo "- $model"
    done
    echo "Please check these models manually."
fi
