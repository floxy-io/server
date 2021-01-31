package home

import (
	"context"
	"fmt"
	"github.com/danielsussa/floxy/internal/infra/compiler"
	"github.com/danielsussa/floxy/internal/infra/recaptcha"
	"github.com/danielsussa/floxy/internal/infra/repo"
	"github.com/danielsussa/floxy/internal/sshserver"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"log"
	"strings"
	"time"
)

var (
	e *echo.Echo
)

func Shutdown(ctx context.Context)error{
	return e.Shutdown(ctx)
}

var AssetsPath string

func Start(){
	if AssetsPath == ""{
		AssetsPath = "internal/home/assets"
	}
	log.Println("using path: ", AssetsPath)
	go func(){
		e = echo.New()
		e.Static("/", AssetsPath)
		e.Static("/burn", AssetsPath)
		e.Static("/share/:fingerprint", AssetsPath)
		e.GET("/api/download/:fingerprint/:kind", downloadBinary)
		e.GET("/api/floxy/:fingerprint", getHostByFingerprint)
		e.GET("/internal/hosts", getAllHosts)
		e.POST("/api/floxy/burn", burn)
		e.Logger.Fatal(e.Start(":8080"))
	}()
}

func downloadBinary(c echo.Context) error{
	file := fmt.Sprintf("internal/home/cooked_bin/%s/%s", c.Param("fingerprint"), c.Param("kind"))
	log.Println(file)
	return c.File(file)
}

type burnResponse struct {
	Status      string `json:"status"`
	Fingerprint string `json:"fingerprint"`
}

func getAllHosts(c echo.Context)error{
	sshHosts, err := repo.GetAll()
	if err != nil {
		log.Println(err)
		return c.String(400, err.Error())
	}
	return c.JSON(200, sshHosts)
}

type getHostResponse struct {
	Fingerprint    string `json:"fingerPrint"`
	RemotePassword *string `json:"remotePassword"`
	LinkExpiration int `json:"linkExpiration"`
}

func getHostByFingerprint(c echo.Context)error{
	sshHosts, err := repo.GetByFingerprint(c.Param("fingerprint"))
	if err != nil {
		log.Println(err)
		return c.String(503, "cannot access this page")
	}
	expLink := int(10.0 - time.Now().Sub(sshHosts.CreatedAt).Minutes())
	if expLink < 0 {
		return c.String(503, "cannot access this page")
	}
	return c.JSON(200, getHostResponse{
		Fingerprint:    sshHosts.Fingerprint,
		RemotePassword: sshHosts.RemotePassword,
		LinkExpiration: expLink,
	})
}

type burnRequest struct {
	Token          string
	Expiration     int
	RemotePassword bool
}

func burn(c echo.Context) error{
	var request burnRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(200, burnResponse{Status: "ssh_err"})
	}

	log.Println("receive request: ", request)

	// recaptcha
	{
		res, err := recaptcha.Get(request.Token)
		if err != nil || !res.Success{
			return c.JSON(200, burnResponse{Status: "non_approve"})
		}
		if res.Score < 0.5 {
			return c.JSON(200, burnResponse{Status: "non_approve"})
		}
	}

	fingerPrint := uuid.New().String()
	serverHost, err := sshserver.AllocateNewHost()
	if err != nil{
		log.Println(err)
		return c.JSON(200, burnResponse{Status: "ssh_err"})
	}

	var remotePass *string
	if request.RemotePassword {
		k := strings.Split(uuid.New().String(), "-")[0]
		remotePass = &k
	}

	binaryRes, err := compiler.Make(compiler.MakeRequest{
		PKey:           serverHost.PrivateKey,
		FingerPrint:    fingerPrint,
		RemotePassword: remotePass,
	})

	if err != nil{
		log.Println(err)
		return c.JSON(200, burnResponse{Status: "bin_err"})
	}

	err = repo.AddNewFloxy(repo.Floxy{
		PublicKey:      serverHost.PublicKey,
		Fingerprint:    fingerPrint,
		RemotePassword: remotePass,
		Expiration:     time.Now().Add(time.Hour * time.Duration(request.Expiration)),
		Port:           serverHost.Port,
	})
	if err != nil{
		log.Println(err)
		return c.JSON(200, burnResponse{Status: "bin_err"})
	}

	return c.JSON(200, burnResponse{Status: "approved", Fingerprint: binaryRes.FingerPrint})
}
