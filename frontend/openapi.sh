#!/bin/sh

# https://openapi-generator.tech/docs/generators/typescript-rxjs/

VERSION=7.15.0
OPENAPI_PATH=http://localhost:8080/api/openapi.json

# Download OpenaAPI spec
# Download generator
wget https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/$VERSION/openapi-generator-cli-$VERSION.jar
# Generate
java -jar openapi-generator-cli-$VERSION.jar generate -i $OPENAPI_PATH -g typescript-rxjs -o src/api --enable-post-process-file
# Remove generator and spec
rm openapi-generator-cli-$VERSION.jar
