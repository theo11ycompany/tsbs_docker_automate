#!/bin/bash

# Default variable values
seed="1234"
scale="400"
start_date="2023-09-23T00:00:00Z"
end_date="2023-09-25T00:00:00Z"
username="default"
password="password"
query_type="groupby-orderby-limit"
end_date_q="2023-09-25T00:00:01Z"
log_inter="30s"
queries="1000"
target="timescaledb"
host="localhost" 
user="postgres"
ssl="disable"
workers_value=10 


# Function to display script usage
usage() {
  echo "Usage: $0 [OPTIONS]"
  echo "Options:"
  echo "--seed          Specify seed for data generation"
  echo "--scale         Specify number of devices generating data"
  echo "--start_date    Start date of records"
  echo "--end_date      End date of records"
  echo "--end_date_q    End date of query"
  echo "--log_inter     How often should devices generate data"
  echo "--target        Database to target while generating data and query"
  echo "--query_type    Query type to evaluate with"
  echo "--queries       Number of queries to test against"
  echo "--password      Password of user in the database"
  echo "--username      Username of user in the database"
}

# Function to handle options and arguments
handle_options() {
  while [ $# -gt 0 ]; do
    case $1 in
      -h | --help)
        usage
        exit 0
        ;;
      --seed=*)
        seed="${1#*=}"
        shift
        ;;
      --scale=*)
        scale="${1#*=}"
        shift
        ;;
      --start_date=*)
        start_date="${1#*=}"
        shift
        ;;
      --end_date=*)
        end_date="${1#*=}"
        shift
        ;;
      --end_date_q=*)
        end_date_q="${1#*=}"
        shift
        ;;
      --log_inter=*)
        log_inter="${1#*=}"
        shift
        ;;
      --target=*)
        target="${1#*=}"
        shift
        ;;
      --query_type=*)
        query_type="${1#*=}"
        shift
        ;;
      --queries=*)
        queries="${1#*=}"
        shift
        ;;
      --username=*)
        username="${1#*=}"
        shift
        ;;
      --password=*)
        password="${1#*=}"
        shift
        ;;
      *)
        echo "Invalid option: $1" >&2
        usage
        exit 1
        ;;
    esac
  done
}

# Main script execution
handle_options "$@"



echo "Generating data first..."
./tsbs_generate_data --use-case="devops" --seed="$seed" --scale="$scale" \
    --timestamp-start="$start_date" \
    --timestamp-end="$end_date" \
    --log-interval="$log_inter" --format="$target" \
    | gzip > "/tmp/${target}-data.gz"



echo "Generating queries..."
./tsbs_generate_queries --use-case="devops" --seed="$seed" --scale="$scale" \
    --timestamp-start="$start_date" \
    --timestamp-end="$end_date_q" \
    --queries="$queries" --query-type="$query_type" --format="$target" \
    | gzip > "/tmp/${target}-query.gz"



echo "Loading data..."
gunzip -c "/tmp/${target}-data.gz" | ./tsbs_load_timescaledb load --pass "$password" --workers=$workers_value  > temp_data_insert_performance 



echo "Running queries" 
gunzip -c "/tmp/${target}-query.gz" | ./tsbs_run_queries_timescaledb --workers=$workers_value  --postgres="host=\"$host\" user=\"$user\" password=\"$password\"  sslmode=\"$ssl\""  > temp_query_running_performance
