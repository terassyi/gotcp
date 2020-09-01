package util

import (
	"math/rand"
	"os/exec"
	"strconv"
	"time"
)

func DisableOsTcpStack(port int) error {
	// iptables -t raw -A PREROUTING -p tcp --dport <your port> -j DROP
	cmd := exec.Command("iptables",
		"-t", "raw",
		"-A", "PREROUTING",
		"-p", "tcp",
		"--dport", strconv.Itoa(port),
		"-j", "DROP")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func GetRandomUint32() uint32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint32()
}
