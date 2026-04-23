#!/bin/bash
export JAVA_HOME=$(/usr/libexec/java_home -v 21.0)
export PATH=$JAVA_HOME/bin:$PATH

echo "JAVA_HOME " $JAVA_HOME  

source .env

# convert standard Base64 to Base64URL first, then embed in JWK
JWT_KEY_URL=$(echo -n "$JWT_SECRET" | tr '+/' '-_' | tr -d '=')
export JWT_JWK=$(echo -n "{\"kty\":\"oct\",\"alg\":\"HS256\",\"k\":\"$JWT_KEY_URL\",\"use\":\"sig\"}" | base64 | tr '+/' '-_' | tr -d '=')

env | grep JWT

java -version

#java -Dquarkus.profile=prod -jar ./target/quarkus-app/quarkus-run.jar 
#java -Dquarkus.http.port=8180 -Dkafka.bootstrap.servers=OUTSIDE://localhost:55004 \
#	-jar target/quarkus-app/quarkus-run.jar
java -Dquarkus.http.port=8180 -jar target/quarkus-app/quarkus-run.jar
