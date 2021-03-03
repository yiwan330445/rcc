package operations

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
)

const (
	getMethod              = `GET`
	postMethod             = `POST`
	deleteMethod           = `DELETE`
	contentType            = `content-type`
	contentLength          = `content-length`
	authorization          = `authorization`
	nonceHeader            = `authorization-timestamp`
	applicationJson        = `application/json`
	applicationOctetStream = `application/octet-stream`
	WorkspaceApi           = `/token-vendor-v1/workspaces/%s/tokenrequest`
	UserApi                = `/token-vendor-v1/user/tokenrequest`
	UserDetails            = `/token-vendor-v1/user/details`
	DeleteCredentials      = `/token-vendor-v1/credential`
	newline                = "\n"
)

type Capability map[string]bool
type Capabilities map[string]Capability

type Claims struct {
	ExpiresIn    int          `json:"expiresIn,omitempty"`
	Capabilities Capabilities `json:"capabilities,omitempty"`
	Method       string       `json:"-"`
	Url          string       `json:"-"`
	Name         string       `json:"-"`
}

type Token map[string]interface{}

func (it Token) AsJson() (string, error) {
	body, err := json.MarshalIndent(it, "", "  ")
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (it Token) FromJson(content []byte) error {
	err := json.Unmarshal(content, &it)
	if err != nil {
		return err
	}
	return nil
}

type UserInfo struct {
	User Token `json:"user"`
	Link Token `json:"request"`
}

func NewClaims(name, url string, expires int) *Claims {
	result := Claims{
		ExpiresIn:    expires,
		Capabilities: make(Capabilities),
		Url:          url,
		Name:         name,
		Method:       postMethod,
	}
	return &result
}

func (it *Claims) AsGet() *Claims {
	it.Method = getMethod
	return it
}

func (it *Claims) AsDelete() *Claims {
	it.Method = deleteMethod
	return it
}

func (it *Claims) IsGet() bool {
	return it.Method == getMethod
}

func (it *Claims) AsJson() (string, error) {
	body, err := json.MarshalIndent(it, "", "  ")
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (it Capabilities) Add(name string, list, read, write bool) {
	capability := make(Capability)
	capability["list"] = list
	capability["read"] = read
	capability["write"] = write
	it[name] = capability
}

func ActivityClaims(seconds int, workspace string) *Claims {
	result := NewClaims("Activity", fmt.Sprintf(WorkspaceApi, workspace), seconds)
	result.Capabilities.Add("activity", true, true, true)
	return result
}

func AssistantClaims(seconds int, workspace string) *Claims {
	result := NewClaims("Assistant", fmt.Sprintf(WorkspaceApi, workspace), seconds)
	result.Capabilities.Add("assistant", true, true, true)
	return result
}

func RobotClaims(seconds int, workspace string) *Claims {
	result := NewClaims("Robot", fmt.Sprintf(WorkspaceApi, workspace), seconds)
	result.Capabilities.Add("package", true, true, true)
	return result
}

func RunClaims(seconds int, workspace string) *Claims {
	result := NewClaims("Run", fmt.Sprintf(WorkspaceApi, workspace), seconds)
	result.Capabilities.Add("secret", true, true, false)
	result.Capabilities.Add("artifact", false, false, true)
	result.Capabilities.Add("livedata", false, true, true)
	result.Capabilities.Add("workitemdata", false, true, true)
	return result
}

func WorkspaceTreeClaims(seconds int) *Claims {
	result := NewClaims("User", UserApi, seconds)
	result.Capabilities.Add("workspace", true, false, false)
	result.Capabilities.Add("workspaceTree", true, true, false)
	return result
}

func DeleteClaims() *Claims {
	return NewClaims("Delete", DeleteCredentials, 1).AsDelete()
}

func VerificationClaims() *Claims {
	return NewClaims("Verification", UserDetails, 0).AsGet()
}

func BearerToken(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}

func WorkspaceToken(token string) string {
	return fmt.Sprintf("RC_WST %s", token)
}

func RobocorpCloudHmac(identifier, token string) string {
	return fmt.Sprintf("robocloud-hmac %s %s", identifier, token)
}

func Digest(incoming string) string {
	hasher := sha256.New()
	hasher.Write([]byte(incoming))
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func HmacSignature(claims *Claims, secret, nonce, bodyHash string) string {
	payload := strings.Join([]string{claims.Method, claims.Url, applicationJson, nonce, bodyHash}, newline)
	hasher := hmac.New(sha256.New, []byte(secret))
	hasher.Write([]byte(payload))
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func AuthorizeClaims(accountName string, claims *Claims) (Token, error) {
	account := AccountByName(accountName)
	if account == nil {
		return nil, fmt.Errorf("Could not find account by name: %s", accountName)
	}
	client, err := cloud.NewClient(account.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("Could not create client for endpoint: %s reason: %w", account.Endpoint, err)
	}
	data, err := AuthorizeCommand(client, account, claims)
	if err != nil {
		return nil, fmt.Errorf("Could not authorize: %w", err)
	}
	return data, nil
}

func AuthorizeCommand(client cloud.Client, account *account, claims *Claims) (Token, error) {
	common.Timeline("authorize %s", claims.Name)
	found, ok := account.Cached(claims.Name, claims.Url)
	if ok {
		cached := make(Token)
		cached["endpoint"] = client.Endpoint()
		cached["requested"] = claims
		cached["status"] = "200"
		cached["when"] = common.When
		cached["token"] = found
		return cached, nil
	}
	body, err := claims.AsJson()
	if err != nil {
		return nil, err
	}
	bodyHash := Digest(body)
	size := len([]byte(body))
	nonce := fmt.Sprintf("%d", common.When)
	signed := HmacSignature(claims, account.Secret, nonce, bodyHash)
	request := client.NewRequest(claims.Url)
	request.Headers[contentType] = applicationJson
	request.Headers[authorization] = RobocorpCloudHmac(account.Identifier, signed)
	request.Headers[nonceHeader] = nonce
	request.Headers[contentLength] = fmt.Sprintf("%d", size)
	request.Body = strings.NewReader(body)
	response := client.Post(request)
	if response.Status != 200 {
		return nil, fmt.Errorf("%d: %s", response.Status, response.Body)
	}
	token := make(Token)
	err = json.Unmarshal(response.Body, &token)
	if err != nil {
		return nil, err
	}
	token["endpoint"] = client.Endpoint()
	token["requested"] = claims
	token["status"] = response.Status
	token["when"] = common.When
	account.WasVerified(common.When)
	trueToken, ok := token["token"].(string)
	if ok {
		deadline := common.When + int64(3*(claims.ExpiresIn/4))
		account.CacheToken(claims.Name, claims.Url, trueToken, deadline)
	}
	return token, nil
}

func DeleteAccount(client cloud.Client, account *account) error {
	claims := DeleteClaims()
	bodyHash := Digest("{}")
	nonce := fmt.Sprintf("%d", common.When)
	signed := HmacSignature(claims, account.Secret, nonce, bodyHash)
	request := client.NewRequest(claims.Url)
	request.Headers[contentType] = applicationJson
	request.Headers[authorization] = RobocorpCloudHmac(account.Identifier, signed)
	request.Headers[nonceHeader] = nonce
	response := client.Delete(request)
	if response.Status < 200 || 299 < response.Status {
		return fmt.Errorf("%d: %s", response.Status, response.Body)
	}
	return nil
}

func UserinfoCommand(client cloud.Client, account *account) (*UserInfo, error) {
	claims := VerificationClaims()
	bodyHash := Digest("{}")
	nonce := fmt.Sprintf("%d", common.When)
	signed := HmacSignature(claims, account.Secret, nonce, bodyHash)
	request := client.NewRequest(claims.Url)
	request.Headers[contentType] = applicationJson
	request.Headers[authorization] = RobocorpCloudHmac(account.Identifier, signed)
	request.Headers[nonceHeader] = nonce
	response := client.Get(request)
	if response.Status != 200 {
		return nil, fmt.Errorf("%d: %s", response.Status, response.Body)
	}
	var result UserInfo
	err := json.Unmarshal(response.Body, &result)
	if err != nil {
		return nil, err
	}
	link := make(Token)
	link["endpoint"] = client.Endpoint()
	link["requested"] = claims
	link["status"] = response.Status
	link["when"] = common.When
	result.Link = link
	account.WasVerified(common.When)
	account.UpdateDetails(result.User)
	return &result, nil
}
