package session

import (
  "net/http"
  "net/url"

  "github.com/CanadianCommander/MicroWeb/pkg/logger"
)

const (
  //DefaultTTL is the default time to live for a session in seconds
  DefaultTTL = 360
)

//Save saves a session cookie in to the response to an http request under the given cookie name.
func Save(cookieName string, ses *Session, res http.ResponseWriter) error {
  return save(cookieName, ses, DefaultTTL, false, res)
}

func save(cookieName string, ses *Session, ttl int, secure bool, res http.ResponseWriter) error {
  cookieData, err :=  ses.GetBuffer();
  if err != nil{
    logger.LogError("failed to save cookie [%s] with error: %s", cookieName, err.Error())
    return err
  }

  cookie := &http.Cookie{}
  cookie.Name = cookieName
  cookie.Value = url.QueryEscape(string(cookieData))
  cookie.MaxAge = ttl
  cookie.Secure = secure

  http.SetCookie(res, cookie)
  return nil
}

//Load loades a session from the data found in the http request under the given cookie name
func Load(cookieName string, ses *Session, req *http.Request) error {
  cookie, err := req.Cookie(cookieName)
  if err != nil {
    return err
  }

  cookieData, err := url.QueryUnescape(cookie.Value)
  if err != nil {
    logger.LogError("error un encoding cookie data: %s", err.Error())
    return err
  }

  err = ses.FromBuffer([]byte(cookieData))
  if err != nil {
    logger.LogWarning("session decode failed for [%s] from IP [%s] with error: %s", cookieName, req.RemoteAddr, err.Error())
    return err
  }

  return nil
}
