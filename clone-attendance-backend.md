## How to Contribute to `apuri_kaihatsu`

### Setting up Your Local Development Environment

1. **Fork the Repository**
   ```terminal
   git fork https://github.com/jdu9a0009/attendance_backend
   ```

2. **Clone the Forked Repository**
   ```terminal
   git clone https://github.com/YOUR-USERNAME/attendance_backend
   ```

3. **Navigate to the Project Directory**
   ```terminal
   cd attendance_backend
   ```
### Getting Latest Code Changes From original Repository to Your Repository
   Fetching the latest changes from the original repository (upstream) into your forked repository is a common task in version control. Here are the steps to achieve that using Git:

1. **Add the original repository as a remote:**
   - First, navigate to your local repository.
   - Add the upstream repository URL.
   ```terminal
   git remote add upstream https://github.com/jdu9a0009/attendance_backend
   ```

2. **Fetch the latest changes from the upstream repository:**
   - This command will fetch all the branches from the upstream repository but won't merge any changes.
   ```terminal
   git fetch upstream
   ```

3. **Merge the changes into your local branch:**
   - Checkout the branch you want to update (usually `main` or `master`).
   ```terminal
   git checkout main
   ```
   - Merge the changes from the upstream branch into your local branch.
   ```terminal
   git merge upstream/main
   ```

4. **Push the changes to your forked repository:**
   - Finally, push the updated branch to your forked repository on GitHub.
   ```sh
   git push origin main
   ```

That's it! Your forked repository should now be up-to-date with the latest changes from the original repository.

### Making Changes

1. **Create a New Branch**
   ```terminal
   git checkout -b feature-branch-name
   ```

2. **Make Your Changes and Commit**
   ```terminal
   git add .
   git commit -m "Describe your changes"
   ```

### Submitting a Pull Request

1. **Push to Your Forked Repository**
   ```terminal
   git push origin feature-branch-name
   ```

2. **Create a Pull Request**
   - Go to your repository on GitHub.
   - Click on the "Compare & pull request" button.
   - Provide a clear title and description for your pull request.



By following these steps, you can contribute effectively to the `jdu9a0009/attendance_backend` repository.
