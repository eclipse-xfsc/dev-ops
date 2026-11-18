package org.eclipse.xfsc.keycloak;

import io.nats.client.Connection;
import io.nats.client.Nats;
import org.jboss.logging.Logger;
import org.keycloak.Config;
import org.keycloak.events.EventListenerProvider;
import org.keycloak.events.EventListenerProviderFactory;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.KeycloakSessionFactory;
import java.nio.charset.StandardCharsets;
import java.time.Duration;

public final class XfscNatsEventListenerFactory implements EventListenerProviderFactory {
    private static final Logger LOG = Logger.getLogger(XfscNatsEventListenerFactory.class);
    private Connection connection;
    @Override public void init(Config.Scope config) {
        String url = System.getenv().getOrDefault("XFSC_NATS_URL", "nats://localhost:4222");
        try { connection = Nats.connect(url); }
        catch (Exception e) { LOG.error("Unable to connect to NATS; events will not be delivered", e); }
    }
    @Override public void postInit(KeycloakSessionFactory factory) { }
    @Override public EventListenerProvider create(KeycloakSession session) { return new XfscNatsEventListener(session, this); }
    @Override public String getId() { return "xfsc-nats"; }
    // Global providers receive events without a per-realm Event Listener setting.
    @Override public boolean isGlobal() { return true; }
    void publish(String subject, String payload) {
        try {
            if (connection == null || connection.getStatus() != Connection.Status.CONNECTED) {
                LOG.errorf("NATS unavailable, event not delivered: %s", subject);
                return;
            }
            connection.publish(subject, payload.getBytes(StandardCharsets.UTF_8));
            connection.flush(Duration.ofSeconds(2));
        } catch (Exception e) { LOG.errorf(e, "Failed to publish %s", subject); }
    }
    @Override public void close() { if (connection != null) { try { connection.close(); } catch (InterruptedException e) { Thread.currentThread().interrupt(); } } }
}
