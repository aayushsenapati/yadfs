#!/bin/bash

# Check if at least one argument is provided
if [ $# -lt 1 ]; then
    echo "Usage: $0 <directory_name>"
    exit 1
fi

# Get the first argument as the directory name
directory_name="$1"

# Check if the directory already exists
if [ -d "$directory_name" ]; then
    echo "Directory '$directory_name' already exists."
    exit 1
fi

# Create the directory
mkdir "$directory_name"

# Check if the directory creation was successful
if [ $? -eq 0 ]; then
    echo "Directory '$directory_name' created successfully."

    # Create symbolic links
    ln -s _dn/maintenance "$directory_name/maintenance"
    ln -s _dn/filemanager "$directory_name/filemanager"
    ln -s _dn/main.go "$directory_name/main.go"
    ln -s _dn/go.mod "$directory_name/go.mod"

    echo "Symbolic links created successfully."
else
    echo "Failed to create directory '$directory_name'."
fi