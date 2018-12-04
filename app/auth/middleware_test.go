package auth

import (
  "encoding/json"
  "io/ioutil"
  "net/http"
  "net/http/httptest"
  "net/url"
  "strconv"
  "testing"

  assert_lib "github.com/stretchr/testify/assert"
)

type authResp struct {
  UserID int64  `json:"userID"`
  Error  string `json:"error"`
}

func callAuthThroughMiddleware(sessionID string, authBackendFn func(w http.ResponseWriter, r *http.Request), wrongURL bool) (bool, *http.Response) {
  // setup auth backend
  backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    authBackendFn(w, r)
  }))
  defer backend.Close()

  // get the URL (and alter it for some tests)
  backendURL, _ := url.Parse(backend.URL)
  backendURLStr := backendURL.String()
  if wrongURL {
    backendURLStr = backendURLStr + "9"
  }

  // dummy server using the middleware
  middleware := UserIDMiddleware(backendURLStr + "/a_path")
  enteredService := false // used to log if the service has been reached
  handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    enteredService = true // has passed into the service
    userID := r.Context().Value(ctxUserID).(int64)
    body := "user_id:" + strconv.FormatInt(userID, 10)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(body))
  })
  mainSrv := httptest.NewServer(middleware(handler))
  defer mainSrv.Close()

  // calling web server
  mainRequest, _ := http.NewRequest("GET", mainSrv.URL, nil)
  if sessionID != "" {
    mainRequest.AddCookie(&http.Cookie{Name: "PHPSESSID", Value: sessionID})
  }
  client := &http.Client{}
  resp, _ := client.Do(mainRequest)

  return enteredService, resp
}

func TestValid(t *testing.T) {
  assert := assert_lib.New(t)

  didService, resp := callAuthThroughMiddleware("123", func(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.ParseInt(r.URL.Query()["sessionid"][0], 10, 64)
    dataJSON, _ := json.Marshal(&authResp{id, ""})
    w.Write(dataJSON)
  }, false)
  defer resp.Body.Close()
  bodyBytes, _ := ioutil.ReadAll(resp.Body)
  assert.Equal(200, resp.StatusCode)
  assert.True(didService)
  assert.Contains(string(bodyBytes), "user_id:123")
}

func TestMissingSession(t *testing.T) {
  assert := assert_lib.New(t)

  didService, resp := callAuthThroughMiddleware("", func(w http.ResponseWriter, r *http.Request) {}, false)
  defer resp.Body.Close()
  bodyBytes, _ := ioutil.ReadAll(resp.Body)
  assert.Equal(401, resp.StatusCode)
  assert.False(didService)
  assert.Contains(string(bodyBytes), "expected auth cookie")
}

func TestNotResponding(t *testing.T) {
  assert := assert_lib.New(t)

  didService, resp := callAuthThroughMiddleware("123", func(w http.ResponseWriter, r *http.Request) {}, true)
  assert.Equal(502, resp.StatusCode)
  assert.False(didService)
}

func TestInvalidResponseFormat1(t *testing.T) {
  type invalidAuthResp struct {
    Message string `json:"message"`
  }
  assert := assert_lib.New(t)

  didService, resp := callAuthThroughMiddleware("1", func(w http.ResponseWriter, r *http.Request) {
    dataJSON, _ := json.Marshal([]invalidAuthResp{invalidAuthResp{"duh?"}}) // unexpected format
    w.Write(dataJSON)
  }, false)
  defer resp.Body.Close()
  bodyBytes, _ := ioutil.ReadAll(resp.Body)
  assert.Equal(502, resp.StatusCode)
  assert.False(didService)
  // the lib does not unmarshal as it cannot fit an array into the struct
  assert.Contains(string(bodyBytes), "Unable to parse")
}
func TestInvalidResponseFormat2(t *testing.T) {
  type invalidAuthResp struct {
    Message string `json:"message"`
  }
  assert := assert_lib.New(t)

  didService, resp := callAuthThroughMiddleware("1", func(w http.ResponseWriter, r *http.Request) {
    dataJSON, _ := json.Marshal(invalidAuthResp{"duh?"}) // unexpected format
    w.Write(dataJSON)
  }, false)
  defer resp.Body.Close()
  bodyBytes, _ := ioutil.ReadAll(resp.Body)
  assert.Equal(502, resp.StatusCode)
  assert.False(didService)
  // the lib still unmarshals it but with empty field as none is matching
  assert.Contains(string(bodyBytes), "Invalid response ")
}

func TestInvalidJSON(t *testing.T) {
  assert := assert_lib.New(t)

  didService, resp := callAuthThroughMiddleware("1", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("this is invalid json"))
  }, false)
  defer resp.Body.Close()
  bodyBytes, _ := ioutil.ReadAll(resp.Body)
  assert.Equal(502, resp.StatusCode)
  assert.False(didService)
  // the lib does not unmarshal as it cannot fit an array into the struct
  assert.Contains(string(bodyBytes), "Unable to parse")
}

func TestAuthError(t *testing.T) {
  assert := assert_lib.New(t)

  didService, resp := callAuthThroughMiddleware("1", func(w http.ResponseWriter, r *http.Request) {
    dataJSON, _ := json.Marshal(&authResp{-1, "invalid token error"})
    w.Write(dataJSON)
  }, false)
  defer resp.Body.Close()
  bodyBytes, _ := ioutil.ReadAll(resp.Body)
  assert.Equal(401, resp.StatusCode)
  assert.False(didService)
  assert.Contains(string(bodyBytes), "Unable to validate the session ID") // middleware
  assert.Contains(string(bodyBytes), "invalid token error")               // returned by the server
}

func TestAuthErrorPositiveID(t *testing.T) {
  assert := assert_lib.New(t)

  didService, resp := callAuthThroughMiddleware("1", func(w http.ResponseWriter, r *http.Request) {
    dataJSON, _ := json.Marshal(&authResp{99, "invalid token error"})
    w.Write(dataJSON)
  }, false)
  defer resp.Body.Close()
  bodyBytes, _ := ioutil.ReadAll(resp.Body)
  assert.Equal(401, resp.StatusCode)
  assert.False(didService)
  assert.Contains(string(bodyBytes), "Unable to validate the session ID") // middleware
  assert.Contains(string(bodyBytes), "invalid token error")               // returned by the server
}

func TestInvalidID(t *testing.T) {
  assert := assert_lib.New(t)

  didService, resp := callAuthThroughMiddleware("1", func(w http.ResponseWriter, r *http.Request) {
    dataJSON, _ := json.Marshal(&authResp{-1, ""}) // unexpected resp from the auth server
    w.Write(dataJSON)
  }, false)
  defer resp.Body.Close()
  bodyBytes, _ := ioutil.ReadAll(resp.Body)
  assert.Equal(502, resp.StatusCode)
  assert.False(didService)
  assert.Contains(string(bodyBytes), "Invalid response")
}
