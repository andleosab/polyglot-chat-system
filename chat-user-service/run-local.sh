#!/bin/bash
export JAVA_HOME=$(/usr/libexec/java_home -v 21.0)
export PATH=$JAVA_HOME/bin:$PATH

echo "JAVA_HOME " $JAVA_HOME  

source .env

env | grep DB

# read -n 1 -s -r -p "Press any key to continue..."

export SPRING_DATASOURCE_HIKARI_MAXIMUMPOOLSIZE=30

export HTTP_WIRE_LOGGING_LEVEL=TRACE

export HIBERNATE_BASIC_BINDER_LOGGING_LEVEL=TRACE

export SERVER_PORT=8080

# java -Xdebug -Xnoagent -Djava.compiler=NONE -Xrunjdwp:transport=dt_socket,address=8000,server=y,suspend=n \
# 	-jar  ./target/*.jar $1 --spring.profiles.active=default,DV,local
	
java  -DHIBERNATE_BASIC_BINDER_LOGGING_LEVEL=TRACE -jar ./target/chat-user-service.jar  --spring.profiles.active=default,local
