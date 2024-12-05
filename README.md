# How to Run the Backend

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

1. **Set a Postgresql usernama and password**:
To Set a Postgresql Username And Password  Write Command One by One.
    ```terminal  
sudo -i -u postgres
psql
    ```

1. **Create a Postgresql Database**:
Create a Postgresql database .
    ```terminal firstly  navigate  your postgresql, then write this query to create database
    
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

3.  **Run the database migration**:
Migrate the `database.sql` file. Open git bash or terminal and navigate to the backend directory. Run the following command:

    ```terminal
    mysql -u root -p < backend/database.sql
    ```

4.  **Run the database migration  And Run Server **:
Migrate the `database.sql` file. Open git bash or terminal and navigate to the backend directory. Run the following command:

    ```terminal
    make run
    ```
