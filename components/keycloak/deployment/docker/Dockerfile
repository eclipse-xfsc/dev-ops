FROM maven:3.9-eclipse-temurin-17 AS listener-builder
WORKDIR /src
COPY listener/pom.xml listener/pom.xml
COPY listener/src listener/src
RUN mvn -B -f listener/pom.xml -DskipTests package

FROM quay.io/keycloak/keycloak:26.4.0
WORKDIR /opt/keycloak
COPY themes/th1 themes/xfsc
COPY --from=listener-builder /src/listener/target/xfsc-keycloak-nats-listener-1.0.0.jar providers/xfsc-keycloak-nats-listener.jar
RUN /opt/keycloak/bin/kc.sh build --spi-events-listener--xfsc-nats--enabled=true
