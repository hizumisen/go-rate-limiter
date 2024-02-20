#!/bin/bash

docker stop dynamodb-local
docker rm dynamodb-local
docker run -d --name dynamodb-local -p 8000:8000 --restart always amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb


home_dir=$(pwd)

cd $home_dir
for dir in */
do

    dir=$home_dir/$dir
    cd $dir

    if !([ -f "go.mod" ]); then
        continue
    else
        echo $dir
    fi

    echo "Running on $dir"
    go test -cover ./...
    if [ $? -ne 0 ]; then 
        echo "Tests failed"
        exit 1
    fi

done
