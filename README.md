# Kumparan
Kumparan backend test service repository

## Main Features
- Post Articles
- Get a list of articles

## Tech Stack  
- **Language:** Go  
- **Database:** PostgreSQL  
- **Search Engine:** Elasticsearch  
- **Configuration:** YAML or `.env` file  
- **Build Tool:** Native Go build  
- **Migration Support:** Built-in via application flag `--migrate`  

## Available Endpoints
| Method | Endpoint           | Description                                               |
| ------ | ------------------ | --------------------------------------------------------- |
| GET    | `/healthcheck`     | Returns a simple status to confirm the service is alive |
| POST   | `/api/v1/articles` | Create a new article                                      |
| GET    | `/api/v1/articles` | Retrieve a list of articles (supports pagination)         |


## Running Services
### 1. Build the Binary
Run the following command to compile the Go application into a binary:
```
go build -o ./bin/kumparan-be-test -v ./cmd/main.go
```
### 2. Configure the Service
Create a configuration file in the `./bin/conf` directory. The application supports both `.yaml` and `.env` formats. If the configuration file is not specified, the program will search for the configuration in the OS environment variables.

#### Example YAML Configuration
```
service_data:
address: 8080
log_level: "debug"

source_data:
postgresdb_server: localhost
postgresdb_port: 5432
postgresdb_name: dbname
postgresdb_username: dbusername
postgresdb_password: dbpass
postgresdb_timeout: 10
postgresdb_max_conns: 10
postgresdb_min_conns: 2
postgresdb_max_conn_lifetime: 3600
postgresdb_max_conn_idle_time: 1800
elasticsearch_url: http://localhost:9200
```
#### Using `.env`
Alternatively, you can use an environment file. You may copy and customize the provided example:
```
cp .env.example ./bin/conf/cfg.env
```
⚠️ Make sure to update cfg.env to match the actual name of your config file.
### 3. Run Database Migration
Execute the following command to run database migrations:
```
./bin/kumparan-be-test --migrate --config "./bin/conf/cfg.env"
```
### 4. Start the Service
Once the configuration is set and migrations have completed successfully, start the service using:
```
./bin/kumparan-be-test --config "./conf/cfg.env"
```

If you prefer to use environment variables directly (without a config file), omit the --config flag:
```
./bin/kumparan-be-test
```