package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Request struct {
	Command   string `json:"command"`
	Wallpaper string `json:"wallpaper,omitempty"`
}

type Response struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func main() {
	controllerAddr := flag.String("controller", "127.0.0.1:9000", "controller address")
	name := flag.String("name", "", "optional device name")
	flag.Parse()

	deviceName := strings.TrimSpace(*name)
	if deviceName == "" {
		host, err := os.Hostname()
		if err != nil {
			deviceName = "unknown-device"
		} else {
			deviceName = host
		}
	}

	for {
		if err := runAgent(*controllerAddr, deviceName); err != nil {
			log.Printf("connection lost: %v", err)
		}

		time.Sleep(3 * time.Second)
	}
}

func runAgent(controllerAddr, deviceName string) error {
	conn, err := net.Dial("tcp", controllerAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("connected to controller %s as %s", controllerAddr, deviceName)

	if err := json.NewEncoder(conn).Encode(map[string]string{"name": deviceName}); err != nil {
		return err
	}

	reader := bufio.NewReader(conn)
	decoder := json.NewDecoder(reader)
	encoder := json.NewEncoder(conn)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			return err
		}

		resp := Response{OK: true, Message: "command executed successfully"}
		if err := executeCommand(req); err != nil {
			resp.OK = false
			resp.Message = err.Error()
		}

		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}
}

func executeCommand(req Request) error {
	switch strings.ToLower(strings.TrimSpace(req.Command)) {
	case "lock":
		return lockDevice()
	case "shutdown":
		return shutdownDevice()
	case "wallpaper":
		if strings.TrimSpace(req.Wallpaper) == "" {
			return errors.New("wallpaper path is required for wallpaper command")
		}
		return changeWallpaper(req.Wallpaper)
	default:
		return fmt.Errorf("unsupported command: %s", req.Command)
	}
}

func lockDevice() error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32.exe", "user32.dll,LockWorkStation").Run()
	case "linux":
		return exec.Command("loginctl", "lock-session").Run()
	case "darwin":
		return exec.Command(
			"/System/Library/CoreServices/Menu Extras/User.menu/Contents/Resources/CGSession",
			"-suspend",
		).Run()
	default:
		return fmt.Errorf("lock is not implemented for %s", runtime.GOOS)
	}
}

func shutdownDevice() error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("shutdown", "/s", "/t", "0").Run()
	case "linux":
		return exec.Command("shutdown", "now").Run()
	case "darwin":
		return exec.Command("sudo", "shutdown", "-h", "now").Run()
	default:
		return fmt.Errorf("shutdown is not implemented for %s", runtime.GOOS)
	}
}

func changeWallpaper(path string) error {
	switch runtime.GOOS {
	case "windows":
		script := fmt.Sprintf(
			`Add-Type @"
using System.Runtime.InteropServices;
public class Wallpaper {
  [DllImport("user32.dll", SetLastError = true)]
  public static extern bool SystemParametersInfo(int uAction, int uParam, string lpvParam, int fuWinIni);
}
"@; [Wallpaper]::SystemParametersInfo(20, 0, "%s", 3)`,
			escapeForPowerShell(path),
		)
		return exec.Command("powershell", "-NoProfile", "-Command", script).Run()
	default:
		return fmt.Errorf("wallpaper change is only implemented for windows in this demo")
	}
}

func escapeForPowerShell(value string) string {
	return strings.ReplaceAll(value, `"`, "`\"")
}
