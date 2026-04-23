FROM eclipse-temurin:21-jre-jammy

ARG JAR_FILE=target/*.jar

# Create dirs & set permissions for non-root user
RUN mkdir -p /opt/app/logs \
    && chown -R 1001 /opt/app \
    && chmod -R 775 /opt/app
	
WORKDIR /opt/app
COPY --chown=1001 ${JAR_FILE} service.jar

USER 1001
ENTRYPOINT ["java","-jar","service.jar"]
