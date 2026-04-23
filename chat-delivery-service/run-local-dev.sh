#!/bin/bash
export JAVA_HOME=$(/usr/libexec/java_home -v 21.0)
export PATH=$JAVA_HOME/bin:$PATH

echo "JAVA_HOME " $JAVA_HOME  

java -version

./mvnw quarkus:dev -Dquarkus.profile=dev

#java -Dquarkus.profile=prod -jar ./target/quarkus-app/quarkus-run.jar 
#java -jar target/quarkus-app/quarkus-run.jar
