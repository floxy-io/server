package compiler

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type MakeRequest struct {
	PKey           string
	FingerPrint    string
	RemotePassword *string
	Distro         []DistroRequest
}

type DistroRequest struct {
	Os       string
	Platform string
	Kind     string
}

type MakeResponse struct {
	FingerPrint string
	ChildBinary []ChildBinary
}

type ChildBinary struct {
	Fingerprint string
	Kind        string
	Os          string
	Platform    string
}

var mutex sync.Mutex

func RemoveLink(fingerprint string)error{
	path := filepath.Join("internal", "home", "cooked_bin", fingerprint)
	os.Remove(path)
	return nil
}

func Make(req MakeRequest)(MakeResponse, error){
	mutex.Lock()
	defer mutex.Unlock()

	folder := filepath.Join("internal", "home", "cooked_bin", req.FingerPrint)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		err = os.Mkdir(folder, 0700)
		if err != nil {
			return MakeResponse{}, err
		}
	}

	childs := make([]ChildBinary,0)

	for _, distro := range req.Distro {
		cFingerprint, err := compile(compileRequest{
			FingerPrint: req.FingerPrint,
			PKey:        req.PKey,
			Password:    req.RemotePassword,
			Folder:      folder,
			Distro:      distro,
		})
		if err != nil {
			return MakeResponse{}, err
		}
		childs = append(childs, ChildBinary{
			Fingerprint: cFingerprint,
			Kind:        distro.Kind,
			Os:          distro.Os,
			Platform:    distro.Platform,
		})
	}

	if os.Getenv("LOG_KEY") == "true"{
		log.Println("\nkey: ", req.PKey, "\nfingerprint: ", req.FingerPrint)
	}

	return MakeResponse{FingerPrint: req.FingerPrint, ChildBinary: childs}, nil
}

var CustomGoPath string
var CustomPath string

type compileRequest struct {
	FingerPrint string
	PKey        string
	Password    *string
	Folder      string
	Distro      DistroRequest
}

func (d DistroRequest)binaryName()string{
	name := "floxy"
	if strings.Contains(d.Os, "windows"){
		name += ".exe"
	}
	return name
}

func (d DistroRequest) envs()[]string{
	f1 := []string{fmt.Sprintf("GOOS=%s", d.Os), fmt.Sprintf("GOARCH=%s", d.Platform)}
	//f2 := []string{"GOOS", d.Os, "GOARCH", d.Platform}
	return f1
}

func compile(req compileRequest)(string, error){
	ldFlags := fmt.Sprintf("-X main.FingerPrint=%s -X main.PrivateKey=%s -X main.SshHost=%s -X main.Kind=%s", req.FingerPrint, req.PKey, os.Getenv("FLOXY_SSH_HOST"), req.Distro.Kind)
	if req.Password != nil {
		ldFlags += fmt.Sprintf(" -X main.RemotePassword=%s", *req.Password)
	}
	cFingerprint := uuid.New().String()

	err := os.Mkdir(filepath.Join(req.Folder,cFingerprint), 0700)
	if err != nil {
		return "", err
	}


	executable := filepath.Join(req.Folder,cFingerprint, req.Distro.binaryName())

	cmdArgs := append([]string{"build"}, "-ldflags", ldFlags)
	cmdArgs = append(cmdArgs, "-o", executable, "internal/cook/cook.go")

	build := exec.Command("go", cmdArgs...)
	build.Env = append(build.Env,os.Environ()...)
	build.Env = append(build.Env, req.Distro.envs()...)

	_, err = build.CombinedOutput()
	if err != nil {
		return "", err
	}
	return cFingerprint, nil
}

func tarFile(folder string)error{
	err := os.Mkdir(filepath.Join(folder, "compress"), 0700)
	if err != nil {
		return err
	}

	tarfile, err := os.Create(filepath.Join(folder,"compress", "floxy.tar"))
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	return filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.Contains(info.Name(), "compress"){
				return filepath.SkipDir
			}
			return nil
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		if err := tarball.WriteHeader(header); err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(tarball, file)
		return err
	})
}

func zipFile(folder string)error{
	err := os.Mkdir(filepath.Join(folder, "compress"), 0700)
	if err != nil {
		return err
	}


	zipFile, err := os.Create(filepath.Join(folder,"compress", "floxy.zip"))
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.Contains(info.Name(), "compress"){
				return filepath.SkipDir
			}
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := zipWriter.Create(info.Name())
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)

		return err
	})
}