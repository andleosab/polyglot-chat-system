export JAVA_HOME=$(/usr/libexec/java_home -v 21.0)
export PATH=$JAVA_HOME/bin:$PATH

echo "JAVA_HOME " $JAVA_HOME  

java -version

./mvnw clean package -DskipTests
