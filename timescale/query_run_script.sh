#!/bin/bash

#### NOTE : Be careful of prints in this script as this stdout is fed into golang processed and into a csv 



# Function to display script usage
usage() {
 echo "Usage: $0 [OPTIONS]"
 echo "Options:"
 echo "--workers       Number of workers to do operations with"
 echo "--password      Password of user in database"
 echo "--username      Username of user in database"
}

has_argument() {
    [[ ("$1" == *=* && -n ${1#*=}) || ( ! -z "$2" && "$2" != -*)  ]];
}

extract_argument() {
  echo "${2:-${1#*=}}"
}

username="default"
password="password"
workers_value=10 

handle_options() {
  while [ $# -gt 0 ]; do
    case $1 in
      -h | --help)
        usage
        exit 0
        ;;
      --workers*)
        if ! has_argument $@; then
          continue
        fi

        workers=$(extract_argument $@)

        shift
        ;;
      --username*)
        if ! has_argument $@; then
          continue
        fi 

        username=$(extract_argument $@)

        shift
        ;;

      --password*)
        if ! has_argument $@; then
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

handle_options "$@"


echo "Running queries" 
cat /tmp/`echo "$target-data"`.gz | gunzip | ./tsbs_run_queries_timescaledb --workers=$workers_value  --postgres="host=\"$host\" user=\"$user\" password=\"$password\"  sslmode=\"$ssl\"" 
