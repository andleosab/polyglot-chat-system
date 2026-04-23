############################
# Stage 1: build
############################
FROM registry.access.redhat.com/ubi9/openjdk-21:1.21 AS build

# use writable working dir
WORKDIR /code

# copy maven wrapper first (cache layer)
COPY mvnw ./
COPY .mvn .mvn
COPY pom.xml ./

# ensure mvnw executable
USER root
RUN chmod +x mvnw

# download dependencies (cached unless pom.xml changes)
RUN ./mvnw -B org.apache.maven.plugins:maven-dependency-plugin:3.8.1:go-offline

# switch to runtime UID used by UBI images
USER 185

# copy source last so code changes don't invalidate dependency cache
COPY --chown=185 src ./src

# build the application
RUN ./mvnw -B package


############################
# Stage 2: runtime
############################
FROM registry.access.redhat.com/ubi9/openjdk-21:1.21

ENV LANGUAGE='en_US:en'

# copy artifacts from build stage
COPY --from=build --chown=185 /code/target/quarkus-app/lib/ /deployments/lib/
COPY --from=build --chown=185 /code/target/quarkus-app/*.jar /deployments/
COPY --from=build --chown=185 /code/target/quarkus-app/app/ /deployments/app/
COPY --from=build --chown=185 /code/target/quarkus-app/quarkus/ /deployments/quarkus/

EXPOSE 8080

USER 185

ENV JAVA_OPTS_APPEND="-Dquarkus.http.host=0.0.0.0 -Djava.util.logging.manager=org.jboss.logmanager.LogManager"
ENV JAVA_APP_JAR="/deployments/quarkus-run.jar"

ENTRYPOINT ["/opt/jboss/container/java/run/run-java.sh"]