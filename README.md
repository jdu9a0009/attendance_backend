# How to Run the Backend

## Clone on your local  not server


- How to clone attendance_backend  [cloning steps](clone-attendnace-backend.md)

## Prerequisites

- Install Docker 
    ```terminal on your local or server
    # write commend one by one
    
     sudo apt update
     sudo upt upgrade
     sudo apt install docker.io
     ```
- Install Postgresql 
    ```terminal on your local or server
    # write commend one by one
    
     sudo apt install postgresql postgresql-contrib
     ```

## Setting Up Environment

1. **Set a Postgresql username and password Creat Database**:
To Set a Postgresql Username And Password  Write Command One by One.
    ```terminal  
sudo -i -u postgres
psql
CREATE USER your_user WITH PASSWORD 'your_password';
CREATE DATABASE IF NOT EXISTS attendances;
    ```

2. **Write Environment Variables**:
Open the `config.yaml` file. 

Write database dependencies that yours;

  - **db_username:**: "username"
  - **db_password**: "password"
 - **db_host**: "host"  //like 164.90.180.81
  - **db_name**: "database name"
  - **default_lang**: `uz`//jp or other
  - **port**: "5432" //if your port is other change it
  - **disable_tls**: "true" // recommend always true
  - **base_url**: "https://164.90.180.81:8080/api/v1" //just change host
  - **jwt_key**: "attendancePanel" // recommend dont change


3.  **Run the database migration  And  Server **:
 Open attendance_backend file VS Code or other. Open the terminal and run the following command:

    ```terminal
    make run
    ```
4.  **Local changes Update to Server**:
 If you change smth on your local  code and update it on server just push it on main branch , CI/CD pipeline Automatic update your server:

    ```terminal
  git add .
  git commit -m "Describe your changes"
  git push origin main
       ```
