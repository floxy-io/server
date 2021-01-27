package home

import (
	"context"
	"fmt"
	"github.com/danielsussa/floxy/internal/infra/compiler"
	"github.com/danielsussa/floxy/internal/infra/recaptcha"
	"github.com/danielsussa/floxy/internal/sshserver"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"log"
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
	fmt.Println("using path: ", AssetsPath)
	go func(){
		e = echo.New()
		e.Static("/", AssetsPath)
		e.GET("/download/:fingerprint/:kind", downloadBinary)
		e.GET("/internal/hosts", getAllHosts)
		e.POST("/burn", burn)
		e.Logger.Fatal(e.Start(":8080"))
	}()
}

func downloadBinary(c echo.Context) error{
	file := fmt.Sprintf("internal/home/cooked_bin/%s/%s", c.Param("fingerprint"), c.Param("kind"))
	fmt.Println(file)
	return c.File(file)
}

type burnResponse struct {
	Success bool `json:"success"`
	Fingerprint string `json:"fingerprint"`
}

func getAllHosts(c echo.Context)error{
	sshHosts, err := sshserver.GetAll()
	if err != nil {
		log.Println(err)
		return c.String(400, err.Error())
	}
	return c.JSON(200, sshHosts)
}

func burn(c echo.Context) error{
	{
		res, err := recaptcha.Get(c.QueryParam("token"))
		if err != nil || !res.Success{
			return c.String(200, "nok")
		}
		if res.Score < 0.5 {
			return c.String(200, "nok")
		}
		if res.Score < 0.9 {
			return c.String(200, "challenge")
		}
	}

	fingerPrint := uuid.New().String()
	serverHost, err := sshserver.AllocateNewHost(fingerPrint)
	if err != nil{
		log.Println(err)
		return c.JSON(200, burnResponse{Success: false})
	}

	binaryRes, err := compiler.Make(compiler.MakeRequest{
		PKey:        serverHost.PrivateKey,
		FingerPrint: fingerPrint,
	})

	if err != nil{
		log.Println(err)
		return c.JSON(200, burnResponse{Success: false})
	}

	return c.JSON(200, burnResponse{Success: true, Fingerprint: binaryRes.FingerPrint})
}
