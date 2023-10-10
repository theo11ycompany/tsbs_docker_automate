#!/bin/bash

# Default variable values
verbose_mode=false
output_file=""

# Function to display script usage
usage() {
 echo "Usage: $0 [OPTIONS]"
 echo "Options:"
 echo "--seed          Specificy seed for data generation "
 echo "--scale         Specificy number of devices generating data"
 echo "--start_date    Start date of records"
 echo "--end_date      End date of records"
 echo "--end_date_q    End date of query"
 echo "--log_inter     How often should devices generate data"
 echo "--target        Database to target while generating data and query"
 echo "--query_type    Query type to evaluate with"
 echo "--queries       Number of queries to test against"
 echo "--password      Password of user in database"
 echo "--username      Username of user in database"
}

has_argument() {
    [[ ("$1" == *=* && -n ${1#*=}) || ( ! -z "$2" && "$2" != -*)  ]];
}

extract_argument() {
  echo "${2:-${1#*=}}"
}

seed="1234"
scale="400"
start_date="2023-09-23T00:00:00Z"
end_date="2023-09-25T00:00:00Z"
username="default"
password=""
query_type="groupby-orderby-limit"
end_date_q="2023-09-25T00:00:01Z"
log_inter="30s"
queries="1000"
target="clickhouse"
# Function to handle options and arguments
handle_options() {
  while [ $# -gt 0 ]; do
    case $1 in
      -h | --help)
        usage
        exit 0
        ;;
      --seed*)
        if ! has_argument $@; then
          echo "Seed not mentioned, default 1234" >&2
          continue
        fi

        seed=$(extract_argument $@)

        shift
        ;;
      --scale*)
        if ! has_argument $@; then
          echo "Scale not mentioned, default 400" >&2
          continue
        fi

        scale=$(extract_argument $@)

        shift
        ;;
      --start_date*)
        if ! has_argument $@; then
          echo "start_date not mentioned, defaulting" >&2
          continue
        fi

        start_date=$(extract_argument $@)

        shift
        ;;
      --end_date*)
        if ! has_argument $@; then
          echo "end_date not mentioned, defaulting" >&2
          continue
        fi 

        end_date=$(extract_argument $@)

        shift
        ;;
      --end_date_q*)
        if ! has_argument $@; then
          echo "end_date_q not mentioned, defaulting" >&2
          continue
        fi 

        end_date_q=$(extract_argument $@)

        shift
        ;;
      --log_inter*)
        if ! has_argument $@; then
          echo "log_inter not mentioned, defaulting to 30s" >&2
          continue
        fi 

        log_inter=$(extract_argument $@)

        shift
        ;;
      --target*)
        if ! has_argument $@; then
          echo "log_inter not mentioned, defaulting clickhouse" >&2
          continue
        fi 

        target=$(extract_argument $@)

        shift
        ;;

      --query_type*)
        if ! has_argument $@; then
          echo "query_type not mentioned, defaulting to groupby-orderby-limit" >&2
          continue
        fi 

        query_type=$(extract_argument $@)

        shift
        ;;

      --queries*)
        if ! has_argument $@; then
          echo "queries not mentioned, defaulting to 1000" >&2
          continue
        fi 

        queries=$(extract_argument $@)

        shift
        ;;
      --username*)
        if ! has_argument $@; then
          echo "username not mentioned, defaulting to default" >&2
          continue
        fi 

        username=$(extract_argument $@)

        shift
        ;;

      --password*)
        if ! has_argument $@; then
          echo "passoword not mentioned, empty default" >&2
          continue
        fi 

        password=$(extract_argument $@)

        shift
        ;;
      *)
        echo "Invalid option: $1" >&2
        usage
        exit 1
        ;;
    esac
    shift
  done
}

# Main script execution
handle_options "$@"

echo "Generating data first..."
./tsbs_generate_data --use-case="devops" --seed=$seed --scale=$scale \
    --timestamp-start=$start_date \
    --timestamp-end=$end_date \
    --log-interval=$log_inter --format=$target \
    | gzip > /tmp/`echo "$target-data"`.gz

echo "generating queries..."

./tsbs_generate_queries --use-case="devops" --seed=$seed --scale=$scale \
    --timestamp-start= $start_date\
    --timestamp-end= $end_date_q\
    --queries=$queries --query-type=$query_type --format=$target \
    | gzip > /tmp/`echo "$target-query"`.gz

# echo "Loading data..."
#
# cat /tmp/`echo "$target-data"`.gz | gunzip |  ./tsbs_load_clickhouse `[ -z "$password" ] && echo "" || echo "--password $password"` > temp_data_insert_performance 
#
# echo "Running queries" 
#
# cat /tmp/`echo "$target-query"`.gz | gunzip |  ./tsbs_run_queries_clickhouse `[ -z "$password" ] && echo "" || echo "--password $password"` > temp_query_running_performance
