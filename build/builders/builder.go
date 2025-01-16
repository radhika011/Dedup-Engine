package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	osFlag := flag.String("target", runtime.GOOS, "Target OS, default windows")
	archFlag := flag.String("arch", runtime.GOARCH, "Target Architecture, default amd64")
	flag.Parse()

	var (
		installerFilePath   string
		uninstallerFilePath string
		ext                 string
		exec_ext            string
	)

	install_name := "installer"
	uninstall_name := "uninstaller"
	lin_ext := ".sh"
	win_ext := ".bat"
	env_background_file := ".env.background"
	env_common_file := ".env.common"

	cur, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	rootPath := filepath.Dir(filepath.Dir(cur))
	os.Chdir(rootPath)
	fmt.Println("root", rootPath)

	buildPath := filepath.Join(rootPath, "build")
	pkgPath := filepath.Join(rootPath, "client-pkg")

	os.RemoveAll(pkgPath)
	os.Mkdir("client-pkg", 0777)

	if *osFlag == "windows" {
		ext = win_ext
		exec_ext = ".exe"
	} else if *osFlag == "linux" {
		ext = lin_ext
		exec_ext = ""
	}

	installerFilePath = filepath.Join(buildPath, "scripts", install_name+ext)
	uninstallerFilePath = filepath.Join(buildPath, "scripts", uninstall_name+ext)

	installer, err := os.OpenFile(installerFilePath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	uninstaller, err := os.OpenFile(uninstallerFilePath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	pkgInstallerPath := filepath.Join(pkgPath, install_name+ext)
	pkgInstaller, err := os.OpenFile(pkgInstallerPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	_, err = io.Copy(pkgInstaller, installer)
	if err != nil {
		fmt.Println(err)
	}

	pkgUninstallerPath := filepath.Join(pkgPath, uninstall_name+ext)
	pkgUninstaller, err := os.OpenFile(pkgUninstallerPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	_, err = io.Copy(pkgUninstaller, uninstaller)
	if err != nil {
		fmt.Println(err)
	}

	cli_name := "bkupClient" + exec_ext

	wailsDir := filepath.Join(rootPath, "client-wails")
	os.Chdir(wailsDir)
	cmd := exec.Command("wails", "build", "-clean", "-platform", *osFlag, "-o", cli_name)
	cmd.Run()
	os.Rename(filepath.Join(wailsDir, "build", "bin", cli_name), filepath.Join(pkgPath, cli_name))

	backgroundDir := filepath.Join(rootPath, "client-background")
	os.Chdir(backgroundDir)
	cmd = exec.Command("go", "build", "-o", filepath.Join(pkgPath, "client-background"+exec_ext))
	cmd.Env = append(os.Environ(),
		"GOOS="+*osFlag,
		"GOARCH="+*archFlag,
	)

	cmd.Run()
	os.Rename(filepath.Join(backgroundDir, "client-background"+exec_ext), filepath.Join(pkgPath, "client-background"+exec_ext))

	envCommonFilePath := filepath.Join(buildPath, "prodenv", env_common_file)
	envBackgroundFilePath := filepath.Join(buildPath, "prodenv", env_background_file)

	envCommonFile, err := os.OpenFile(envCommonFilePath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	envBackgroundFile, err := os.OpenFile(envBackgroundFilePath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	pkgEnvCommonPath := filepath.Join(pkgPath, ".env.common")
	pkgEnvCommonFile, err := os.OpenFile(pkgEnvCommonPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	_, err = io.Copy(pkgEnvCommonFile, envCommonFile)
	if err != nil {
		fmt.Println(err)
	}

	pkgEnvBackgroundPath := filepath.Join(pkgPath, ".env.background")
	pkgEnvBackgroundFile, err := os.OpenFile(pkgEnvBackgroundPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	_, err = io.Copy(pkgEnvBackgroundFile, envBackgroundFile)
	if err != nil {
		fmt.Println(err)
	}

	serviceDir := filepath.Join(buildPath, "service-setup")
	os.Chdir(serviceDir)
	cmd = exec.Command("go", "build", "-o", filepath.Join(pkgPath, "servicemgr"+exec_ext))
	cmd.Env = append(os.Environ(),
		"GOOS="+*osFlag,
		"GOARCH="+*archFlag,
	)

	cmd.Run()
	os.Rename(filepath.Join(serviceDir, "servicemgr"+exec_ext), filepath.Join(pkgPath, "servicemgr"+exec_ext))

}
