package home

import (
	"context"
	"fmt"
	"github.com/danielsussa/floxy/internal/infra/compiler"
	"github.com/danielsussa/floxy/internal/infra/recaptcha"
	"github.com/danielsussa/floxy/internal/sshserver"
	"github.com/google/uuid"
	"github.com/labstack/echo"
)

var (
	e *echo.Echo
)

func Shutdown(ctx context.Context)error{
	return e.Shutdown(ctx)
}

func Start(){
	go func(){
		e = echo.New()
		e.Static("/", "internal/home/assets")
		e.GET("/download/:fingerprint/:kind", downloadBinary)
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

	serverHost, err := sshserver.AllocateNewHost()
	if err != nil{
		return c.JSON(200, burnResponse{Success: false})
	}

	binaryRes, err := compiler.Make(compiler.MakeRequest{
		PKey:        serverHost.PKey,
		FingerPrint: uuid.New(),
		Port:        serverHost.Port,
	})

	if err != nil{
		return c.JSON(200, burnResponse{Success: false})
	}

	return c.JSON(200, burnResponse{Success: true, Fingerprint: binaryRes.FingerPrint})
}
