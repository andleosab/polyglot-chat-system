export JAVA_HOME=$(/usr/libexec/java_home -v 21.0)
export PATH=$JAVA_HOME/bin:$PATH

echo "JAVA_HOME " $JAVA_HOME  

java -version

./mvnw clean package -DskipTests
# ./mvnw spring-boot:build-image -DskipTests -Dspring-boot.build-image.imageName=spring/chat-user-service:1.0.0

