
package org.eclipse.xfsc.keycloak;

import org.keycloak.events.Event;
import org.keycloak.events.EventListenerProvider;
import org.keycloak.events.EventType;
import org.keycloak.events.admin.AdminEvent;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.RealmModel;
import org.keycloak.models.UserModel;
import org.keycloak.models.AbstractKeycloakTransaction;

import java.time.Instant;
import java.util.UUID;

public final class XfscNatsEventListener implements EventListenerProvider {

    private final KeycloakSession session;
    private final XfscNatsEventListenerFactory factory;

    XfscNatsEventListener(
            KeycloakSession session,
            XfscNatsEventListenerFactory factory) {
        this.session = session;
        this.factory = factory;
    }

    @Override
    public void onEvent(Event event) {
        if (event.getError() != null || event.getUserId() == null) {
            return;
        }

        EventType kind = event.getType();

        if (kind != EventType.REGISTER
                && kind != EventType.DELETE_ACCOUNT) {
            return;
        }

        String type = kind == EventType.REGISTER
                ? "org.eclipse.xfsc.identity.user.created"
                : "org.eclipse.xfsc.identity.user.deleted";

        String subject = kind == EventType.REGISTER
                ? "tenant.management.identity.user.created"
                : "tenant.management.identity.user.deleted";

        String id = UUID.randomUUID().toString();

        String email = null;
        String firstName = null;
        String lastName = null;

        if (kind == EventType.REGISTER) {
            RealmModel realm = session.realms()
                    .getRealm(event.getRealmId());

            if (realm == null) {
                System.err.println(
                        "XFSC: Realm not found: " + event.getRealmId()
                );
            } else {
                UserModel user = session.users()
                        .getUserById(realm, event.getUserId());

                if (user == null) {
                    System.err.println(
                            "XFSC: User not found: " + event.getUserId()
                    );
                } else {
                    email = user.getEmail();
                    firstName = user.getFirstName();
                    lastName = user.getLastName();

                    System.out.println(
                            "XFSC: REGISTER user found"
                                    + ", emailPresent="
                                    + (email != null && !email.isBlank())
                                    + ", firstNamePresent="
                                    + (firstName != null && !firstName.isBlank())
                                    + ", lastNamePresent="
                                    + (lastName != null && !lastName.isBlank())
                    );
                }
            }
        }

        // Copy user attributes before transaction completion.
        // Do not query the user after commit.
        String data = "{\"realmId\":\""
                + json(event.getRealmId())
                + "\",\"userId\":\""
                + json(event.getUserId())
                + "\""
                + (email == null
                        ? ""
                        : ",\"email\":\"" + json(email) + "\"")
                + (firstName == null
                        ? ""
                        : ",\"firstName\":\"" + json(firstName) + "\"")
                + (lastName == null
                        ? ""
                        : ",\"lastName\":\"" + json(lastName) + "\"")
                + "}";

        String payload = "{\"specversion\":\"1.0\""
                + ",\"id\":\"" + id + "\""
                + ",\"source\":\"urn:xfsc:keycloak\""
                + ",\"type\":\"" + type + "\""
                + ",\"subject\":\"" + json(event.getUserId()) + "\""
                + ",\"time\":\"" + Instant.now() + "\""
                + ",\"datacontenttype\":\"application/json\""
                + ",\"data\":" + data
                + "}";

        session.getTransactionManager().enlistAfterCompletion(
                new AbstractKeycloakTransaction() {
                    @Override
                    protected void commitImpl() {
                        factory.publish(subject, payload);
                    }

                    @Override
                    protected void rollbackImpl() {
                        // Do not publish events on rollback.
                    }
                }
        );
    }

    private static String json(String value) {
        return (value == null ? "" : value)
                .replace("\\", "\\\\")
                .replace("\"", "\\\"")
                .replace("\n", "\\n")
                .replace("\r", "\\r");
    }

    @Override
    public void onEvent(
            AdminEvent event,
            boolean includeRepresentation) {
    }

    @Override
    public void close() {
    }
}
