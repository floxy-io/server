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
		e.Static("/burnApi", AssetsPath)
		e.Static("/burn", AssetsPath)
		e.Static("/about", AssetsPath)
		e.Static("/form", AssetsPath)
		e.Static("/share/:fingerprint", AssetsPath)
		e.GET("/api/download/:fingerprint/:binary/:kind", downloadBinary)
		e.GET("/api/floxy/:fingerprint", getFloxy)
		e.GET("/api/floxy/:fingerprint/status", getFloxy)
		e.GET("/internal/hosts", getAllHosts)
		e.GET("/internal/exclude", excludeNotActive)
		e.POST("/api/floxy/burn", burnApi)
		e.Logger.Fatal(e.Start(":8080"))
	}()
}

func downloadBinary(c echo.Context) error{
	file := fmt.Sprintf("internal/home/cooked_bin/%s/%s/floxy", c.Param("fingerprint"),c.Param("binary"))
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

func excludeNotActive(c echo.Context)error{
	sshHosts, err := repo.GetAll()
	if err != nil {
		log.Println(err)
		return c.String(400, err.Error())
	}

	for _, floxy := range sshHosts {
		if !floxy.IsActive() {
			err = repo.Remove(floxy.Fingerprint)
			if err != nil {
				log.Println(err)
				return c.String(400, err.Error())
			}
		}
		if floxy.ExpiredLink() {
			// remove download file
			_ = compiler.RemoveLink(floxy.Fingerprint)
		}
	}
	return c.JSON(200, sshHosts)
}

type getHostResponse struct {
	Fingerprint    string `json:"fingerPrint"`
	RemotePassword *string `json:"remotePassword"`
	LinkExpiration int     `json:"linkExpiration"`
	Status         string  `json:"status"`
	Binaries       []hostBinary `json:"binaries"`
}

type hostBinary struct {
	Fingerprint string `json:"fingerPrint"`
	Kind        string `json:"kind"`
	Os          string `json:"os"`
	Platform    string `json:"platform"`

}

func getFloxy(c echo.Context)error{
	sshHosts, err := repo.GetByFingerprint(c.Param("fingerprint"))
	if err != nil {
		log.Println(err)
		return c.String(503, "cannot access this page")
	}

	if sshHosts.Status == "burning" {
		return c.JSON(200, getHostResponse{
			Status:         sshHosts.Status,
		})
	}

	expLink := int(10.0 - time.Now().Sub(sshHosts.CreatedAt).Minutes())
	status := sshHosts.Status
	if expLink < 0 {
		status = "expired"
	}
	binaries, err  := repo.GetFloxyBinaries(c.Param("fingerprint"))
	if err != nil {
		log.Println(err)
		return c.String(503, "cannot access this page")
	}

	hostBinaries := make([]hostBinary,0)
	for _, bin := range binaries {
		hostBinaries = append(hostBinaries, hostBinary{
			Fingerprint: bin.Fingerprint,
			Kind:        bin.Kind,
			Os:          bin.Os,
			Platform:    bin.Platform,
		})
	}

	return c.JSON(200, getHostResponse{
		Fingerprint:    sshHosts.Fingerprint,
		RemotePassword: sshHosts.RemotePassword,
		LinkExpiration: expLink,
		Status:         status,
		Binaries:       hostBinaries,
	})
}

type burnRequest struct {
	Token          string
	Expiration     int
	RemotePassword bool
	Distro         []compiler.DistroRequest
}

func burnApi(c echo.Context) error{
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

	err := repo.AddNewFloxy(fingerPrint)
	if err != nil{
		log.Println(err)
		return c.JSON(200, burnResponse{Status: "ssh_err"})
	}

	go schedulerBurn(request, fingerPrint)

	return c.JSON(200, burnResponse{Status: "burning", Fingerprint: fingerPrint})
}

func schedulerBurn(request burnRequest, fingerPrint string){
	serverHost, err := sshserver.AllocateNewHost()
	if err != nil{
		log.Println(err)
		_ = repo.SetFailed(fingerPrint)
		return
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
		Distro:         request.Distro,
	})

	if err != nil{
		log.Println(err)
		_ = repo.SetFailed(fingerPrint)
		return
	}

	for _, floxyBin := range binaryRes.ChildBinary {
		err = repo.AddFloxyBinary(repo.FloxyBinary{
			Parent:      fingerPrint,
			Fingerprint: floxyBin.Fingerprint,
			Kind:        floxyBin.Kind,
			Os:          floxyBin.Os,
			Platform:    floxyBin.Platform,
		})
		if err != nil{
			log.Println(err)
			_ = repo.SetFailed(fingerPrint)
			return
		}
	}

	err = repo.UpdateFloxy(repo.Floxy{
		PublicKey:      serverHost.PublicKey,
		Fingerprint:    fingerPrint,
		RemotePassword: remotePass,
		Expiration:     time.Now().Add(time.Hour * time.Duration(request.Expiration)),
		Port:           serverHost.Port,
	})
	if err != nil{
		_ = repo.SetFailed(fingerPrint)
	}
	err = repo.SetActive(fingerPrint)
}