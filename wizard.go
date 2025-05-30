package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go/v4/kv"
	"github.com/google/uuid"
)

const (
	red    = "\033[31m"
	green  = "\033[32m"
	reset  = "\033[0m"
	orange = "\033[38;2;255;165;0m"
	blue   = "\033[94m"
	bold   = "\033[1m"
	title  = bold + blue + "●" + reset
	ask    = bold + "-" + reset
	info   = bold + "+" + reset
)

type DeployType int

const (
	DTWorker DeployType = iota
	DTPage
)

var DeployTypeNames = map[DeployType]string{
	DTWorker: "worker",
	DTPage:   "page",
}

func (dt DeployType) String() string {
	return DeployTypeNames[dt]
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading worker.js: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}

func generateRandomString(charSet string, length int, isDomain bool) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomBytes := make([]byte, length)

	for i := range randomBytes {
		for {
			char := charSet[r.Intn(len(charSet))]
			if isDomain && (i == 0 || i == length-1) && char == byte('-') {
				continue
			}
			randomBytes[i] = char
			break
		}
	}

	return string(randomBytes)
}

func generateRandomDomain(subDomainLength int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789-"
	return generateRandomString(charset, subDomainLength, true)
}

func generateTrPassword(passwordLength int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+[]{}|;:',.<>?"
	return generateRandomString(charset, passwordLength, false)
}

func generateSubURIPath(uriLength int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@$&*_-+;:,."
	return generateRandomString(charset, uriLength, false)
}

func promptUser(prompt string) string {
	fmt.Printf("%s %s", ask, prompt)
	var response string
	fmt.Scanln(&response)

	return strings.TrimSpace(response)
}

func failMessage(message string) {
	errMark := bold + red + "✗" + reset
	fmt.Printf("%s %s\n", errMark, message)
}

func successMessage(message string) {
	succMark := bold + green + "✓" + reset
	fmt.Printf("%s %s\n", succMark, message)
}

func openURL(isAndroid bool, url string) error {
	var cmd string
	var args = []string{url}

	switch runtime.GOOS {
	case "darwin": // MacOS
		cmd = "open"
	case "windows": // Windows
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default: // Linux, BSD, Android, etc.
		if isAndroid {
			termuxBin := os.Getenv("PATH")
			cmd = filepath.Join(termuxBin, "termux-open-url")
		} else {
			cmd = "xdg-open"
		}
	}

	err := exec.Command(cmd, args...).Start()
	if err != nil {
		return err
	}

	return nil
}

func checkBPBPanel(isAndroid bool, url string) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(5000) * time.Millisecond,
				}

				return d.DialContext(ctx, "udp", "8.8.8.8:53")
			},
		},
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	transport := &http.Transport{
		DisableKeepAlives: true,
		DialContext:       dialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	for range ticker.C {
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf(".")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf(".")
			resp.Body.Close()
			continue
		}

		resp.Body.Close()
		message := fmt.Sprintf("BPB panel is ready -> %s", url)
		successMessage(message)
		fmt.Print("\n")
		prompt := fmt.Sprintf("Would you like to open %sBPB panel%s in browser? (y/n): ", blue, reset)

		if response := promptUser(prompt); strings.ToLower(response) == "n" {
			return nil
		}

		if err = openURL(isAndroid, url); err != nil {
			return err
		}

		return nil
	}

	return nil
}

func configureBPB(isAndroid bool) {
	token := <-obtainedToken
	ctx := context.Background()
	cfClient = NewClient(token)
	var err error

	cfAccount, err = getAccount(ctx)
	if err != nil {
		failMessage("Failed to get Cloudflare account.")
		log.Fatalln(err)
	}

	srcPath, err := os.MkdirTemp("", ".bpb-wizard")
	workerURL := "https://github.com/bia-pain-bache/BPB-Worker-Panel/releases/latest/download/worker.js"
	if err != nil {
		failMessage("Failed to create temp directory.")
		log.Fatalln(err)
	}

	fmt.Printf("\n%s Get settings...\n", title)
	fmt.Printf("\n%s You can use %sWorkers%s or %sPages%s to deploy.\n", info, green, reset, green, reset)
	fmt.Printf("%s %sWarning%s: If you choose %sPages%s, sometimes it takes up to 5 minutes until you can access panel, so please keep calm!\n", info, red, reset, green, reset)
	var deployType DeployType

	for {
		response := promptUser("Please enter 1 for Workers or 2 for Pages deployment: ")
		switch response {
		case "1":
			deployType = DTWorker
		case "2":
			deployType = DTPage
		default:
			failMessage("Wrong selection, Please choose 1 or 2 only!")
			continue
		}

		break
	}

	var projectName string
	for {
		projectName = generateRandomDomain(32)
		fmt.Printf("\n%s The random generated name (%sSubdomain%s) is: %s%s%s\n", info, green, reset, orange, projectName, reset)
		if response := promptUser("Please enter a custom name or press ENTER to use generated one: "); response != "" {
			if strings.Contains(strings.ToLower(response), "bpb") {
				message := fmt.Sprintf("Name cannot contain %sbpb%s! Please try another name.", red, reset)
				failMessage(message)
				continue
			}

			projectName = response
		}

		var isAvailable bool
		fmt.Printf("\n%s Checking domain availablity...\n", title)

		if deployType == DTWorker {
			isAvailable = isWorkerAvailable(ctx, projectName)
		} else {
			isAvailable = isPageAvailable(ctx, projectName)
		}

		if !isAvailable {
			prompt := fmt.Sprintf("This already exists! This will %sRESET%s all panel settings, would you like to override it? (y/n): ", red, reset)
			if response := promptUser(prompt); strings.ToLower(response) == "n" {
				continue
			}
		}

		successMessage("Available!")
		break
	}

	uid := uuid.NewString()
	fmt.Printf("\n%s The random generated %sUUID%s is: %s%s%s\n", info, green, reset, orange, uid, reset)
	if response := promptUser("Please enter a custom uid or press ENTER to use generated one: "); response != "" {
		uid = response
	}

	trPass := generateTrPassword(12)
	fmt.Printf("\n%s The random generated %sTrojan password%s is: %s%s%s\n", info, green, reset, orange, trPass, reset)
	if response := promptUser("Please enter a custom Trojan password or press ENTER to use generated one: "); response != "" {
		trPass = response
	}

	proxyIP := "bpb.yousef.isegaro.com"
	fmt.Printf("\n%s The default %sProxy IP%s is: %s%s%s\n", info, green, reset, orange, proxyIP, reset)
	if response := promptUser("Please enter custom Proxy IP/Domains or press ENTER to use default: "); response != "" {
		proxyIP = response
	}

	fallback := "speed.cloudflare.com"
	fmt.Printf("\n%s The default %sFallback domain%s is: %s%s%s\n", info, green, reset, orange, fallback, reset)
	if response := promptUser("Please enter a custom Fallback domain or press ENTER to use default: "); response != "" {
		fallback = response
	}

	subPath := generateSubURIPath(16)
	fmt.Printf("\n%s The random generated %sSubscription path%s is: %s%s%s\n", info, green, reset, orange, subPath, reset)
	if response := promptUser("Please enter a custom Subscription path or press ENTER to use generated one: "); response != "" {
		subPath = response
	}

	var customDomain string
	fmt.Printf("\n%s You can set %sCustom domain%s ONLY if you registered domain on this cloudflare account.\n", info, green, reset)
	if response := promptUser("Please enter a custom domain (if you have any) or press ENTER to ignore: "); response != "" {
		customDomain = response
	}

	fmt.Printf("\n%s Downloading %sworker.js%s...\n", title, green, reset)
	workerPath := filepath.Join(srcPath, "worker.js")

	for {
		if err = downloadFile(workerURL, workerPath); err != nil {
			failMessage("Failed to download worker.js")
			log.Printf("%v\n\n", err)
			if response := promptUser("Would you like to try again? (y/n): "); strings.ToLower(response) == "n" {
				return
			}
			continue
		}
		successMessage("Worker downloaded successfully!")
		break
	}

	fmt.Printf("\n%s Creating KV namespace...\n", title)
	var kvNamespace *kv.Namespace

	for {
		now := time.Now().Format("2006-01-02_15-04-05")
		kvName := fmt.Sprintf("panel-kv-%s", now)
		kvNamespace, err = createKVNamespace(ctx, kvName)
		if err != nil {
			failMessage("Failed to create KV.")
			log.Printf("%v\n\n", err)
			if response := promptUser("Would you like to try again? (y/n): "); strings.ToLower(response) == "n" {
				return
			}
			continue
		}

		successMessage("KV created successfully!")
		break
	}

	var panel string
	cachePath := filepath.Join(srcPath, "tld.cache")

	switch deployType {
	case DTWorker:
		panel, err = deployBPBWorkers(ctx, projectName, uid, trPass, proxyIP, fallback, subPath, workerPath, kvNamespace, customDomain, cachePath)
	case DTPage:
		panel, err = deployBPBPages(ctx, projectName, uid, trPass, proxyIP, fallback, subPath, workerPath, kvNamespace, customDomain)
	}

	if err != nil {
		failMessage("Failed to get panel URL.")
		log.Fatalln(err)
	}

	if err := checkBPBPanel(isAndroid, panel); err != nil {
		failMessage("Failed to checkout BPB panel.")
		log.Fatalln(err)
	}
}
