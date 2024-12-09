## How to Contribute to `attendance_backend`

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

3. **Configure Docker and Server Settings for GitHub Actions**
 
## To enable secure and efficient CI/CD pipelines, sensitive variables like Docker credentials and server access must be set up as secrets in GitHub Actions. Follow these steps to configure the required secrets:

1.  Open GitHub Actions Secrets
   # Go to your repository on GitHub.
   # Click on Settings > Secrets and variables > Actions.

2. Add the Required Secrets
Set up the following secrets with the appropriate values:
```
# Secret Name	Description
DOCKER_USERNAME	Your Docker Hub username
DOCKER_PASSWORD	Your Docker Hub password
SERVER_HOST	The IP or hostname of the target server
SERVER_USER	The username for accessing the server
SERVER_PASSWORD	The password for the server user
```
# Steps to Add Secrets:

# Click "New repository secret".
Enter the secret name (e.g., DOCKER_USERNAME).
Paste the corresponding value into the input field.
# Click "Add secret".
Repeat these steps for each secret listed above.

3. Use the Secrets in GitHub Actions
Once the secrets are added, they can be referenced in your workflow files. Below is an example configuration:

```

  DOCKER_USERNAME: ${{ secrets.DOCKER_USERNAME }}
  DOCKER_PASSWORD: ${{ secrets.DOCKER_PASSWORD }}
  SERVER_HOST: ${{ secrets.SERVER_HOST }}
  SERVER_USER: ${{ secrets.SERVER_USER }}
  SERVER_PASSWORD: ${{ secrets.SERVER_PASSWORD }}
```
# After that change ,when u push your code to github it automatically update server
By following these steps, you can contribute effectively to the `jdu9a0009/attendance_backend` repository.
