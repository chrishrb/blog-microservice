#!/bin/bash

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

if ! command_exists "openssl"; then
  echo "openssl is not installed. Please install it and try again."
  exit 1
fi

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

openssl ecparam -name prime256v1 -genkey -noout -out "${script_dir}"/../config/user-service/jwt.key.pem
openssl ec -in "${script_dir}"/../config/user-service/jwt.key.pem -pubout -outform PEM -out "${script_dir}"/../config/user-service/jwt.pub.pem 
