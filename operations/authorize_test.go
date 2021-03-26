package operations_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/hamlet"
	"github.com/robocorp/rcc/mocks"
	"github.com/robocorp/rcc/operations"
)

func TestCanCalculateDigestFromText(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	wont_be.Nil(operations.Digest("foo"))
	must_be.Equal(operations.Digest(""), "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=")
	must_be.Equal(operations.Digest("{}"), "RBNvo1WzZ4oRRq0W9+hknpT7T8If536DEMBg9hyq/4o=")
}

func TestCanCalculateDigestHmacFromValues(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	claims := operations.NewClaims("unknown", fmt.Sprintf(operations.WorkspaceApi, ""), 10)
	wont_be.Nil(operations.HmacSignature(claims, "", "", ""))
	must_be.Equal(operations.HmacSignature(claims, "", "", ""), "nwTWU9By+5rAI4yWnltg71QNu7rrv+2o6TMnEZ1XarI=")

	bodyHash := "RBNvo1WzZ4oRRq0W9+hknpT7T8If536DEMBg9hyq/4o="
	must_be.Equal(operations.HmacSignature(claims, "", "", bodyHash), "9CsmhKGbEK2fmIQoHzPJxUwwBrumprAAZUUVzQ3nPN0=")

	nonce := "1590471922"
	must_be.Equal(operations.HmacSignature(claims, "", nonce, bodyHash), "pEVK8rYfOZGflXQpQb08T5fFnT95fXVrhpLignEd3Mc=")

	secret := "hello"
	must_be.Equal(operations.HmacSignature(claims, secret, nonce, bodyHash), "FOuljSztpLu7mvJUUDCSQiCUhTaJtMPvlaotsMlSAx4=")

	claims = operations.NewClaims("unknown", fmt.Sprintf(operations.WorkspaceApi, "2020"), 10)
	must_be.Equal(operations.HmacSignature(claims, secret, nonce, bodyHash), "sTCPARcFoXTZm65kmLeWAgnhAIeYYhMQURhaqLf27vw=")
}

func TestBodyIsCorrectlyConverted(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	reader := strings.NewReader("{\n}")
	wont_be.Nil(reader)
	body, err := ioutil.ReadAll(reader)
	must_be.Nil(err)
	wont_be.Nil(body)
	must_be.Equal("{\n}", string(body))
}

func TestCanCreateBearerToken(t *testing.T) {
	must_be, _ := hamlet.Specifications(t)

	must_be.Equal(operations.BearerToken("foo"), "Bearer foo")
	must_be.Equal(operations.BearerToken("barbie"), "Bearer barbie")
}

func TestCanCreateRobocorpCloudHmac(t *testing.T) {
	must_be, _ := hamlet.Specifications(t)

	must_be.Equal(operations.RobocorpCloudHmac("11", "token"), "robocloud-hmac 11 token")
	must_be.Equal(operations.RobocorpCloudHmac("1234", "abcd"), "robocloud-hmac 1234 abcd")
}

func TestCanCreateNewClaims(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut := operations.NewClaims("Mega", "https://some.com", 232)
	wont_be.Nil(sut)
	must_be.Equal(len(sut.Capabilities), 0)
	sut.Capabilities.Add("secret", true, true, false)
	sut.Capabilities.Add("artifact", false, true, true)
	sut.Capabilities.Add("livedata", false, true, true)
	sut.Capabilities.Add("workitemdata", false, true, true)
	sut.Capabilities.Add("workspace", true, false, false)
	sut.Capabilities.Add("workspaceTree", true, true, false)
	sut.Capabilities.Add("package", true, true, true)
	must_be.Equal(len(sut.Capabilities), 7)
	output, err := sut.AsJson()
	must_be.Nil(err)
	wont_be.Nil(output)
	must_be.True(strings.Contains(output, "workspaceTree"))
	must_be.True(strings.Contains(output, "true"))
	must_be.True(strings.Contains(output, "false"))
	must_be.True(strings.Contains(output, "list"))
	must_be.True(strings.Contains(output, "read"))
	must_be.True(strings.Contains(output, "write"))
}

func TestCanCreateRobotClaims(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	setup := operations.NewClaims("Robot", "https://some.com", 60)
	setup.Capabilities.Add("package", true, true, true)
	expected, err := setup.AsJson()
	must_be.Nil(err)

	sut := operations.RobotClaims(60, "99")
	wont_be.Nil(sut)
	result, err := sut.AsJson()
	must_be.Nil(err)
	must_be.Equal(result, expected)
	must_be.True(strings.Contains(sut.Url, "/workspaces/99/"))
}

func TestCanCreateRunClaims(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	setup := operations.NewClaims("Run", "https://some.com", 88)
	setup.Capabilities.Add("secret", true, true, true)
	setup.Capabilities.Add("artifact", false, false, true)
	setup.Capabilities.Add("livedata", false, true, true)
	setup.Capabilities.Add("workitemdata", false, true, true)
	expected, err := setup.AsJson()
	must_be.Nil(err)

	sut := operations.RunClaims(88, "777")
	wont_be.Nil(sut)
	result, err := sut.AsJson()
	must_be.Nil(err)
	must_be.Equal(result, expected)
	must_be.True(strings.Contains(sut.Url, "/workspaces/777/"))
}

func TestCanConvertEmptyClaimsToJson(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	setup := operations.NewClaims("Empty", "https://some.com", 0)
	actual, err := setup.AsJson()
	must_be.Nil(err)
	wont_be.Nil(actual)
	must_be.Equal("{}", actual)
}

func TestCanGetVerificationClaims(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	sut := operations.VerificationClaims()
	wont_be.Nil(sut)
	actual, err := sut.AsJson()
	must_be.Nil(err)
	must_be.Equal("{}", actual)
	must_be.Equal("GET", sut.Method)
}

func TestCanCreateWorkspaceTreeClaims(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	setup := operations.NewClaims("User", "https://some.com", 49)
	setup.Capabilities.Add("workspace", true, false, false)
	setup.Capabilities.Add("workspaceTree", true, true, false)
	expected, err := setup.AsJson()
	must_be.Nil(err)

	sut := operations.WorkspaceTreeClaims(49)
	wont_be.Nil(sut)
	result, err := sut.AsJson()
	must_be.Nil(err)
	must_be.Equal(result, expected)
	must_be.True(strings.Contains(sut.Url, "/user/"))
}

func TestCanCallAuthorizeCommand(t *testing.T) {
	must_be, wont_be := hamlet.Specifications(t)

	operations.UpdateCredentials("authz", "https://end", "42", "answer")
	account := operations.AccountByName("authz")
	wont_be.Nil(account)
	first := cloud.Response{Status: 200, Body: []byte("{\"token\":\"foo\",\"expiresIn\":1}")}
	client := mocks.NewClient(&first)
	claims := operations.RunClaims(1, "777")
	token, err := operations.AuthorizeCommand(client, account, claims)
	must_be.Nil(err)
	wont_be.Nil(token)
	must_be.Equal(token["token"], "foo")
	must_be.Equal(token["expiresIn"], 1.0)
	must_be.Equal(token["endpoint"], "https://this.is/mock")
}
