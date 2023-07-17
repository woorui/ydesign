// Package auth provides authentication.
package auth

import (
	"strings"

	"github.com/woorui/ydesign/core/frame"
	"github.com/woorui/ydesign/core/metadata"
)

var (
	auths = make(map[string]Authentication)
)

// Authentication for server
type Authentication interface {
	// Authenticate authentication client's credential
	Authenticate(payload string) (metadata.MD, bool)
	// Name authentication name
	Name() string
}

// Register register authentication
func Register(authentication Authentication) {
	auths[authentication.Name()] = authentication
}

// GetAuth get authentication by name
func GetAuth(name string) (Authentication, bool) {
	auth, ok := auths[name]
	return auth, ok
}

// Credential client credential.
type credential struct {
	name    string
	payload string
}

type Credential interface {
	Name() string
	Payload() string
}

// NewCredential create client credential
func NewCredential(payload string) Credential {
	idx := strings.Index(payload, ":")
	if idx != -1 {
		authName := payload[:idx]
		idx++
		authPayload := payload[idx:]
		return &credential{
			name:    authName,
			payload: authPayload,
		}
	}
	return &credential{name: "none"}
}

// Payload client credential payload
func (c *credential) Payload() string {
	return c.payload
}

// Name client credential name
func (c *credential) Name() string {
	return c.name
}

// Authenticate finds an authentication way in `auths` and authenticates the Object.
//
// If `auths` is nil or empty, It returns true, It think that authentication is not required.
func Authenticate(auths map[string]Authentication, obj *frame.AuthenticationFrame) (metadata.MD, bool) {
	if auths == nil || len(auths) <= 0 {
		return nil, true
	}

	if obj == nil {
		return nil, false
	}

	auth, ok := auths[obj.AuthName]
	if !ok {
		return nil, false
	}

	return auth.Authenticate(obj.AuthPayload)
}
