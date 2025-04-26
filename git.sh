#!/bin/bash

# Script to git add, commit, and push

# Check if a commit message is provided as an argument
#
# if [ -z "$Updated" ]; then
#   echo "Usage: $0 <commit_message>"
#   exit 1
# fi

commit_message="$Updated"

# echo "Adding all changes..."
git add .

if [ $? -ne 0 ]; then
  echo "Error during 'git add'."
  exit 1
fi

# echo "Committing changes with message: '$commit_message'"
git commit -m "$commit_message"

if [ $? -ne 0 ]; then
  echo "Error during 'git commit'."
  exit 1
fi

# echo "Pushing changes to the remote repository..."
git push

if [ $? -ne 0 ]; then
  echo "Error during 'git push'."
  exit 1
fi

# echo "Git operations completed successfully."
