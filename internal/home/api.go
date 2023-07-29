package home

import (
	"context"
	"fmt"
	"github.com/danielsussa/floxy/internal/env"
	"github.com/danielsussa/floxy/internal/infra/recaptcha"
	"github.com/danielsussa/floxy/internal/sshserver"
	"github.com/labstack/echo"
	"log"
	"net/http"
)

var (
	e *echo.Echo
)

func Shutdown(ctx context.Context) error {
	return e.Shutdown(ctx)
}

var AssetsPath string

func Start() {
	if AssetsPath == "" {
		AssetsPath = "internal/home/assets"
	}
	log.Println("using path: ", AssetsPath)
	go func() {
		e = echo.New()
		e.Static("/", AssetsPath)
		e.Static("/c2s", AssetsPath)
		e.Static("/share/:fingerprint", AssetsPath)
		e.POST("/api/c2s", remoteLocalProxy)
		e.GET("/api/users/:user", getUserProxy)
		e.POST("/api/w2s", webRemoteProxy)
		e.Logger.Fatal(e.Start(":8080"))
	}()
}

type remoteLocalProxyRequest struct {
	Captcha    string    `json:"captcha"`
	PublicKeys *[]string `json:"public_keys"`
}

type remoteWebProxyResponse struct {
	Kind          string  `json:"kind"`
	Id            string  `json:"id"`
	Password      *string `json:"password,omitempty"`
	ServerCommand string  `json:"server_command"`
	Domain        string  `json:"domain"`
}

func newRemoteWebProxyResponse(user *sshserver.ProxyUserMap) remoteWebProxyResponse {
	sshDns := env.GetOrDefault(env.ServerSshDns, "localhost")

	var sshPortFmt string
	sshPort := env.GetOrDefault(env.ServerSshPort, "2222")
	if sshPort != "22" {
		sshPortFmt = fmt.Sprintf(" -p %s", env.GetOrDefault(env.ServerSshPort, "2222"))
	}

	serverCmd := fmt.Sprintf("ssh -N -R %d:localhost:$PORT %s%s",
		user.Port, sshDns, sshPortFmt)

	serverDns := env.GetOrDefault(env.ServerBaseDns, "localhost")

	return remoteWebProxyResponse{
		Id:            user.Id,
		Password:      user.Password,
		ServerCommand: serverCmd,
		Domain:        fmt.Sprintf("https://%s.%s", *user.SubDns, serverDns),
		Kind:          "w2s",
	}
}

type remoteLocalProxyResponse struct {
	Id            string  `json:"id"`
	Password      *string `json:"password,omitempty"`
	LocalCommand  string  `json:"local_command,omitempty"`
	ServerCommand string  `json:"server_command"`
	Kind          string  `json:"kind"`
}

func newRemoteLocalProxyResponse(user *sshserver.ProxyUserMap) remoteLocalProxyResponse {
	sshDns := env.GetOrDefault(env.ServerSshDns, "localhost")

	var sshPortFmt string
	sshPort := env.GetOrDefault(env.ServerSshPort, "2222")
	if sshPort != "22" {
		sshPortFmt = fmt.Sprintf(" -p %s", env.GetOrDefault(env.ServerSshPort, "2222"))
	}

	localCmd := fmt.Sprintf("ssh -N -L $PORT:localhost:%d %s%s", user.Port, sshDns, sshPortFmt)
	serverCmd := fmt.Sprintf("ssh -N -R %d:localhost:$PORT %s%s",
		user.Port, sshDns, sshPortFmt)

	return remoteLocalProxyResponse{
		Id:            user.Id,
		Password:      user.Password,
		LocalCommand:  localCmd,
		ServerCommand: serverCmd,
		Kind:          "c2s",
	}
}

func webRemoteProxy(c echo.Context) error {
	var req remoteLocalProxyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err,
		})
	}

	res, err := recaptcha.Get(req.Captcha)
	if err != nil || !res.Success {
		return c.JSON(403, map[string]interface{}{
			"error": "captcha not approved",
		})
	}
	if res.Score < 0.5 {
		return c.JSON(403, map[string]interface{}{
			"error": "low score",
		})
	}

	user, err := sshserver.AddNewUser(sshserver.AddNewUserRequest{
		PublicKeys: req.PublicKeys,
		GenDomain:  true,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err,
		})
	}

	return c.JSON(http.StatusCreated, newRemoteWebProxyResponse(user))
}

func getUserProxy(c echo.Context) error {
	userId := c.Param("user")

	user, err := sshserver.GetUserById(userId)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": "user not found",
		})
	}
	if user.SubDns != nil {
		return c.JSON(http.StatusOK, newRemoteWebProxyResponse(user))
	}
	return c.JSON(http.StatusOK, newRemoteLocalProxyResponse(user))
}

func remoteLocalProxy(c echo.Context) error {
	var req remoteLocalProxyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err,
		})
	}

	res, err := recaptcha.Get(req.Captcha)
	if err != nil || !res.Success {
		return c.JSON(403, map[string]interface{}{
			"error": "captcha not approved",
		})
	}
	if res.Score < 0.5 {
		return c.JSON(403, map[string]interface{}{
			"error": "low score",
		})
	}

	user, err := sshserver.AddNewUser(sshserver.AddNewUserRequest{
		PublicKeys: req.PublicKeys,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err,
		})
	}

	return c.JSON(http.StatusCreated, newRemoteLocalProxyResponse(user))
}
