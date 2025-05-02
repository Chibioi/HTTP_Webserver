import os
import subprocess
# from getpass import getpass


def git_add_commit_push():
    # Navigate to your git repository
    repo_path = input(
        "Enter path to your git repository (or leave blank for current dir): "
    ).strip()
    if repo_path:
        os.chdir(repo_path)

    # Check if current directory is a git repo
    if not os.path.isdir(".git"):
        print("Error: This is not a git repository!")
        return

    # Git add
    files = input("Enter files to add (space separated) or '.' for all: ").strip()
    subprocess.run(["git", "add", files])

    # Git commit
    commit_message = input("Enter commit message: ").strip()
    if not commit_message:
        commit_message = "Auto-commit by script"
    subprocess.run(["git", "commit", "-m", commit_message])

    # Git push
    print("\nPushing to remote repository...")
    try:
        subprocess.run(["git", "push"])
        print("Successfully pushed to GitHub!")
    except Exception as e:
        print(f"Error pushing to GitHub: {e}")


if __name__ == "__main__":
    print("=== GitHub Auto-Push Script ===")
    git_add_commit_push()
