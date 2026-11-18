package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	messaging "github.com/eclipse-xfsc/nats-message-library"
	"github.com/eclipse-xfsc/nats-message-library/common"
	"github.com/nats-io/nats.go"
)

type config struct{ base, adminRealm, clientID, clientSecret string }
type controller struct {
	http *http.Client
	cfg  config
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func (c *controller) token(ctx context.Context) (string, error) {
	v := url.Values{"grant_type": {"client_credentials"}, "client_id": {c.cfg.clientID}, "client_secret": {c.cfg.clientSecret}}
	endpoint := c.cfg.base + "/realms/" + url.PathEscape(c.cfg.adminRealm) + "/protocol/openid-connect/token"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(v.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("token endpoint: HTTP %d", res.StatusCode)
	}
	var t struct {
		AccessToken string `json:"access_token"`
	}
	if err = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&t); err != nil {
		return "", err
	}
	if t.AccessToken == "" {
		return "", errors.New("empty access token")
	}
	return t.AccessToken, nil
}
func (c *controller) call(ctx context.Context, method, path string, body any, allowed ...int) error {
	token, err := c.token(ctx)
	if err != nil {
		return err
	}
	var reader io.Reader
	if body != nil {
		b, e := json.Marshal(body)
		if e != nil {
			return e
		}
		reader = strings.NewReader(string(b))
	}
	req, err := http.NewRequestWithContext(ctx, method, c.cfg.base+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	for _, code := range allowed {
		if res.StatusCode == code {
			return nil
		}
	}
	b, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
	return fmt.Errorf("admin API HTTP %d: %s", res.StatusCode, string(b))
}
func (c *controller) handle(ctx context.Context, subject string, raw []byte) (common.Reply, string, error) {
	var header common.Request
	if err := json.Unmarshal(raw, &header); err != nil {
		return common.Reply{}, "", err
	}
	reply := common.Reply{TenantId: header.TenantId, RequestId: header.RequestId, GroupId: header.GroupId}
	if strings.TrimSpace(header.TenantId) == "" || strings.TrimSpace(header.RequestId) == "" {
		return reply, "", errors.New("tenant_id and request_id required")
	}
	realm := url.PathEscape(header.TenantId)
	prefix := "/admin/realms/" + realm
	switch subject {
	case messaging.TenantIdentityCreate, messaging.TenantIdentityUpdate:
		var r messaging.TenantIdentityRequest
		if err := json.Unmarshal(raw, &r); err != nil {
			return reply, "", err
		}
		if err := r.Validate(); err != nil {
			return reply, "", err
		}
		obj := map[string]any{}
		if subject == messaging.TenantIdentityCreate {
			obj["realm"] = r.TenantId
		}
		if r.Enabled != nil {
			obj["enabled"] = *r.Enabled
		}
		if r.RegistrationAllowed != nil {
			obj["registrationAllowed"] = *r.RegistrationAllowed
		}
		if r.DisplayName != "" {
			obj["displayName"] = r.DisplayName
		}
		if subject == messaging.TenantIdentityCreate {
			return reply, "", c.call(ctx, "POST", "/admin/realms", obj, 201, 204)
		}
		return reply, "", c.call(ctx, "PUT", prefix, obj, 204)
	case messaging.TenantIdentityDelete:
		var r messaging.TenantIdentityDeleteRequest
		if err := json.Unmarshal(raw, &r); err != nil {
			return reply, "", err
		}
		return reply, "", c.call(ctx, "DELETE", prefix, nil, 204, 404)
	case messaging.TenantIdentityAppCreate, messaging.TenantIdentityAppUpdate:
		var r messaging.TenantIdentityAppRequest
		if err := json.Unmarshal(raw, &r); err != nil {
			return reply, "", err
		}
		if subject == messaging.TenantIdentityAppCreate {
			if err := r.ValidateCreate(); err != nil {
				return reply, r.AppId, err
			}
		} else {
			if err := r.Validate(); err != nil {
				return reply, r.AppId, err
			}
		}
		obj := map[string]any{"clientId": r.AppId, "protocol": "openid-connect"}
		if r.Name != "" {
			obj["name"] = r.Name
		}
		if r.Enabled != nil {
			obj["enabled"] = *r.Enabled
		}
		if r.Client != nil {
			obj["publicClient"] = r.Client.Type == "public"
			obj["clientAuthenticatorType"] = "client-secret"
		}
		if r.Authorization != nil {
			a := r.Authorization
			obj["standardFlowEnabled"] = true
			obj["implicitFlowEnabled"] = false
			obj["directAccessGrantsEnabled"] = false
			obj["serviceAccountsEnabled"] = false
			if a.RedirectURIs != nil {
				obj["redirectUris"] = a.RedirectURIs
			}
			if a.PKCERequired {
				obj["attributes"] = map[string]string{"pkce.code.challenge.method": "S256"}
			}
		}
		path := prefix + "/clients"
		if subject == messaging.TenantIdentityAppCreate {
			return reply, r.AppId, c.call(ctx, "POST", path, obj, 201, 204)
		}
		// Resolve internal UUID using exact clientId, never treat appId as internal UUID.
		token, err := c.token(ctx)
		if err != nil {
			return reply, r.AppId, err
		}
		u := c.cfg.base + path + "?clientId=" + url.QueryEscape(r.AppId)
		req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
		if err != nil {
			return reply, r.AppId, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := c.http.Do(req)
		if err != nil {
			return reply, r.AppId, err
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return reply, r.AppId, fmt.Errorf("client lookup HTTP %d", res.StatusCode)
		}
		var found []struct {
			ID       string `json:"id"`
			ClientID string `json:"clientId"`
		}
		if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&found); err != nil {
			return reply, r.AppId, err
		}
		for _, v := range found {
			if v.ClientID == r.AppId {
				return reply, r.AppId, c.call(ctx, "PUT", path+"/"+url.PathEscape(v.ID), obj, 204)
			}
		}
		return reply, r.AppId, errors.New("app not found")
	case messaging.TenantIdentityAppDelete:
		var r messaging.TenantIdentityAppDeleteRequest
		if err := json.Unmarshal(raw, &r); err != nil {
			return reply, "", err
		}
		if err := r.Validate(); err != nil {
			return reply, "", err
		}
		// Delete by appId resolved to internal ID through same lookup logic.
		return reply, r.AppId, c.deleteApp(ctx, pathForApps(prefix), r.AppId)
	default:
		return reply, "", fmt.Errorf("unsupported subject %s", subject)
	}
}
func pathForApps(prefix string) string { return prefix + "/clients" }
func (c *controller) deleteApp(ctx context.Context, path, appID string) error {
	token, err := c.token(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", c.cfg.base+path+"?clientId="+url.QueryEscape(appID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("client lookup HTTP %d", res.StatusCode)
	}
	var found []struct {
		ID       string `json:"id"`
		ClientID string `json:"clientId"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&found); err != nil {
		return err
	}
	for _, v := range found {
		if v.ClientID == appID {
			return c.call(ctx, "DELETE", path+"/"+url.PathEscape(v.ID), nil, 204, 404)
		}
	}
	return nil
}
func main() {
	cfg := config{base: strings.TrimRight(env("KEYCLOAK_ADMIN_URL", "http://keycloak:8080"), "/"), adminRealm: env("KEYCLOAK_ADMIN_REALM", "master"), clientID: os.Getenv("KEYCLOAK_CLIENT_ID"), clientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET")}
	if cfg.clientID == "" || cfg.clientSecret == "" {
		log.Fatal("KEYCLOAK_CLIENT_ID and KEYCLOAK_CLIENT_SECRET are required")
	}
	nc, err := nats.Connect(env("NATS_URL", "nats://nats:4222"), nats.MaxReconnects(-1), nats.ReconnectWait(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	c := &controller{http: &http.Client{Timeout: 12 * time.Second}, cfg: cfg}
	subjects := []string{messaging.TenantIdentityCreate, messaging.TenantIdentityUpdate, messaging.TenantIdentityDelete, messaging.TenantIdentityAppCreate, messaging.TenantIdentityAppUpdate, messaging.TenantIdentityAppDelete}
	for _, subject := range subjects {
		_, err = nc.QueueSubscribe(subject, "identity-provisioners", func(m *nats.Msg) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			reply, appID, e := c.handle(ctx, m.Subject, m.Data)
			if e != nil {
				reply.Error = &common.Error{Status: 500, Id: "provisioning_failed", Msg: e.Error()}
				log.Printf("%s: %v", m.Subject, e)
			}
			if m.Reply != "" {
				var out any = reply
				if strings.Contains(m.Subject, ".app.") {
					out = messaging.TenantIdentityAppReply{Reply: reply, AppID: appID}
				}
				b, _ := json.Marshal(out)
				if err := nc.Publish(m.Reply, b); err != nil {
					log.Printf("reply publish: %v", err)
				}
			}
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := nc.Flush(); err != nil {
		log.Fatal(err)
	}
	log.Print("identity provisioning consumer ready")
	select {}
}
