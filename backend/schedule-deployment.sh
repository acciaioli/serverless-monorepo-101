#!/bin/bash

# NOT USED ANYMORE

S3_BUCKET=$1
S3_PREFIX=$2

S3_URI="s3://$S3_BUCKET/$S3_PREFIX"

# constants
DIST=".dist"
PREVIOUS_DIST=".previous-dist"

# used by all functions to return values
# todo: turns out i can return an int from functions, maybe i should rework this... :rolling_eyes:
RETURN=""

# exit codes
OK_SERVICE_UPDATED=0
OK_SERVICE_NOT_UPDATED=10
USER_ERROR=20
FATAL_ERROR=50

# functions should start by setting RETURN=""
# if by the time they return the value of RETURN was not updated,
# it is assumed that something went wrong and the script exists
function catch_exception() {
    if [[ $RETURN == "" ]]
    then
      echo "[fatal] unhandled error - aborting"
      exit $FATAL_ERROR
    fi
}

function validate_inputs() {
  RETURN="true"
  if [[ $S3_BUCKET == "" ]] || [[ $S3_BUCKET == "-" ]]
  then
   echo "[error] deployment s3 bucket not provided"
   RETURN="false"
  else
    echo "[info] deployment s3 bucket: $S3_BUCKET"
  fi
  if [[ $S3_PREFIX == "" ]]
  then
   echo "[error] deployment s3 bucket prefix not provided"
   RETURN="false"
  else
    echo "[info] deployment s3 bucket prefix: $S3_PREFIX"
  fi
}

function generate_dist() {
  RETURN=""
  echo "[info] generating dist"
  sls package --package=$DIST >& /dev/null && RETURN="ok"
}

function previous_dist_exists() {
  RETURN=""
  local dist_found="false"
  aws s3 ls "$S3_URI" >& /dev/null && local dist_found="true"
  if [[ $dist_found == "true" ]]
  then
    echo "[info] previous dist found"
    RETURN="yes"
  else
    echo "[info] previous dist  not found"
    RETURN="no"
  fi
}

function download_previous_dist() {
  RETURN=""
  local download_ok="false"
  aws s3 sync "$S3_URI" "$PREVIOUS_DIST" >& /dev/null && local download_ok="true"
  if [[ $download_ok == "true" ]]
  then
    echo "[info] downloaded previous dist"
    RETURN="ok"
  else
    echo "[error] failed to download previous dist"
  fi
}

function compare_dists() {
  RETURN=""
  local dist_sha
  local previous_dist_sha
  dist_sha=$(sha1sum $DIST/* | sha1sum | awk '{ print $1 }')
  previous_dist_sha=$(sha1sum $PREVIOUS_DIST/* | sha1sum | awk '{ print $1 }')
  if [[ $dist_sha == "$previous_dist_sha" ]]
  then
   echo "[info] current and previous dists are equal"
   RETURN="equal"
  else
   echo "[info] current and previous dists are different"
   RETURN="diff"
  fi
}


function upload_dist() {
  RETURN=""
  local upload_ok="false"
  aws s3 sync "$DIST" "$S3_URI" --delete >& /dev/null  && local upload_ok="true"
  if [[ $upload_ok == "true" ]]
  then
    echo "[info] uploaded new dist"
    RETURN="ok"
  else
    echo "[error] failed to upload new dist"
  fi
}

SERVICE_WAS_UPDATED="false"


validate_inputs
catch_exception
if [[ $RETURN == "false" ]]
then
    exit $USER_ERROR
fi

generate_dist
catch_exception
if [[ $RETURN == "false" ]]
then
    exit $USER_ERROR
fi

previous_dist_exists
catch_exception
if [[ $RETURN == "no" ]]
then
    SERVICE_WAS_UPDATED="true"
else
  download_previous_dist
  catch_exception
  if [[ $RETURN == "ok" ]]
  then
    compare_dists
    catch_exception
    if [[ $RETURN == "diff" ]]
    then
      SERVICE_WAS_UPDATED="true"
    fi
  fi
fi

if [[ $SERVICE_WAS_UPDATED == "true" ]]
then
  upload_dist
  catch_exception
  exit  $OK_SERVICE_UPDATED
else
  exit  $OK_SERVICE_NOT_UPDATED
fi
