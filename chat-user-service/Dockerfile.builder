# ---- build stage ----
# Use Eclipse Temurin JDK 21 on Ubuntu Jammy as the build image
# This stage compiles the application
FROM eclipse-temurin:21-jdk-jammy AS builder

# Set working directory inside container
# All following commands run relative to /app
WORKDIR /app

# Copy Maven wrapper configuration directory
# Contains wrapper JAR and config that defines which Maven version to use
COPY .mvn .mvn

# Copy Maven wrapper script and project descriptor
# pom.xml defines dependencies and build configuration
# mvnw ensures consistent Maven version without requiring Maven preinstalled
COPY mvnw pom.xml ./

# Pre-download all dependencies and Maven plugins
# -B : batch mode (non-interactive, good for CI/docker)
# -q : quiet output (less log noise)
# -e : show full stacktrace if build fails
# -DskipTests : skip tests to speed up image build
# dependency:go-offline : fetch dependencies now so later steps can run without network
#
# This layer is cached unless pom.xml or .mvn changes
RUN ./mvnw -B -q -e -DskipTests dependency:go-offline

# Copy application source code
# Placed AFTER dependency download to maximize Docker cache reuse
# Source code changes frequently, dependencies usually do not
COPY src src

# Compile and package the application (usually creates a JAR in target/)
# Uses previously downloaded dependencies from cached layer
# Much faster rebuild when only source code changes
# add -q to show errors only
RUN ./mvnw -B -DskipTests package

# ---- runtime stage ----
# Use a smaller runtime-only image (JRE instead of full JDK)
# JRE contains only what is needed to RUN the app, not build it
# Results in smaller final image size and fewer security vulnerabilities	
FROM eclipse-temurin:21-jre-jammy

# Create dirs & set permissions for non-root user
# Ensure directory is owned by UID 1001 (non-root user)
# Without this, the application may fail to write logs or temp files
RUN mkdir -p /opt/app/logs \
    && chown -R 1001 /opt/app \
    && chmod -R 775 /opt/app

# Set working directory inside container for this stage
# The application will start from this directory
# Relative paths (logs/, config/, temp files) resolve from here
WORKDIR /opt/app

# Copy the built JAR from the previous stage ("builder")
# --from=builder references the earlier stage in the same Dockerfile
# /app/target/*.jar matches the packaged artifact created by mvn package
# The jar is renamed to app.jar for consistency and simplicity
COPY --from=builder --chown=1001 /app/target/*.jar service.jar

# Switch container to run as non-root user (UID 1001)
USER 1001

# Define the container startup command
# Runs the Java application when container starts
# JSON array form avoids shell interpretation issues
ENTRYPOINT ["java","-Xms256m","-Xmx512m","-jar","service.jar"]
