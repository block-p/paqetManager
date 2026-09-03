package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/block-p/paqetManager/bin"
	"github.com/block-p/paqetManager/internal/config"
	"github.com/block-p/paqetManager/internal/paqet"
	"github.com/block-p/paqetManager/internal/uri"
)

const (
	filepath        = "/opt/paqet/paqet"
	configpath      = "/opt/paqet/"
	restartinterval = time.Hour
)

func main() {
	if err := paqet.InstallPaqet(filepath, bin.PaqetBinFS); err != nil {
		log.Printf("Warning: failed to install paqet binary: %v", err)
	}

	if err := paqet.EnsureDir(configpath); err != nil {
		log.Printf("Warning: failed to create config directory: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	switch {
	case len(os.Args) == 1:
		configFile := path.Join(configpath, "config.yaml")
		runAndSupervise(ctx, filepath, configFile)

	case len(os.Args) == 3 && os.Args[1] == "c", len(os.Args) == 3 && os.Args[1] == "client":
		uriData, err := uri.DecodeUri(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		opts := []config.Option{}
		opts = append(opts, config.WithConn(uriData.Conn))
		opts = append(opts, config.WithBlock(uriData.Block))
		opts = append(opts, config.WithMode(uriData.Mode))
		opts = append(opts, config.WithMTU(uriData.Mtu))
		cconfig, err := config.DefaultClientConfig(uriData.Addr, uriData.Key, opts...)
		if err != nil {
			log.Fatal(err)
		}

		clientConfigFile := path.Join(configpath, "config.yaml")
		if err := cconfig.WriteConfig(clientConfigFile); err != nil {
			log.Fatal(err)
		}

		runAndSupervise(ctx, filepath, clientConfigFile)

	case os.Args[1] == "s", os.Args[1] == "server":
		fs := flag.NewFlagSet("server", flag.ExitOnError)
		keyflag := fs.String("key", "", "key for authentication")
		connflag := fs.Int("conn", 0, "connection number")
		modeflag := fs.String("mode", "", "mode for server")
		portflag := fs.String("port", "9999", "port for server")

		err := fs.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("failed to parse args: %v", err)
		}

		opts := []config.Option{}

		if *keyflag != "" {
			opts = append(opts, config.WithKey(*keyflag))
		}
		if *connflag != 0 {
			opts = append(opts, config.WithConn(*connflag))
		}
		if *modeflag != "" {
			opts = append(opts, config.WithMode(*modeflag))
		}

		serverConfig, err := config.GenerateDefaultConfig(*portflag, opts...)
		if err != nil {
			log.Fatal(err)
		}

		serverConfigFile := path.Join(configpath, "config.yaml")
		if err := serverConfig.WriteConfig(serverConfigFile); err != nil {
			log.Fatal(err)
		}

		ur := uri.MakeUri(*serverConfig)
		fmt.Println(ur)
		runAndSupervise(ctx, filepath, serverConfigFile)

	case os.Args[1] == "c", os.Args[1] == "client":
		fs := flag.NewFlagSet("client", flag.ExitOnError)
		var opts []config.Option

		addrflag := fs.String("addr", "", "server address")
		keyflag := fs.String("key", "", "key for authentication")
		connflag := fs.Int("conn", 0, "connection number")
		modeflag := fs.String("mode", "", "mode for server")
		err := fs.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("failed to parse args: %v", err)
		}

		if *addrflag == "" {
			log.Fatal("addr flag is required")
		}

		if *keyflag == "" {
			log.Fatal("key flag is required")
		}

		if *connflag != 0 {
			opts = append(opts, config.WithConn(*connflag))
		}
		if *modeflag != "" {
			opts = append(opts, config.WithMode(*modeflag))
		}

		clientConfig, err := config.DefaultClientConfig(*addrflag, *keyflag, opts...)
		if err != nil {
			log.Fatal(err)
		}

		clientConfigFile := path.Join(configpath, "config.yaml")
		if err := clientConfig.WriteConfig(clientConfigFile); err != nil {
			log.Fatal(err)
		}

		runAndSupervise(ctx, filepath, clientConfigFile)

	default:
		log.Fatal("invalid arguments")
	}

}

func runAndSupervise(ctx context.Context, paqetBinary, configFilePath string) {
	ticker := time.NewTicker(restartinterval)
	defer ticker.Stop()

	for {
		// Step A: Check and update Router MAC in the config file (server.yaml or client.yaml)
		if err := config.RefreshConfigFileMAC(configFilePath); err != nil {
			log.Printf("[Warning] Failed to refresh router MAC in %s: %v", configFilePath, err)
		}

		// Step B: Start paqet process with its own cancellation context
		procCtx, cancelProc := context.WithCancel(ctx)
		cmd := exec.CommandContext(procCtx, paqetBinary, "run", "-c", configFilePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		log.Println("[Supervisor] Starting paqet...")
		if err := cmd.Start(); err != nil {
			log.Printf("[Error] Failed to start paqet: %v", err)
			cancelProc()
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second): // backoff before retrying
				continue
			}
		}

		// Step C: Monitor process completion in background
		procDone := make(chan error, 1)
		go func() {
			procDone <- cmd.Wait()
		}()

		// Step D: Wait for either interval, process crash, or main app exit
		select {
		case <-ctx.Done():
			// App is shutting down
			log.Println("[Supervisor] Shutting down paqet...")
			cancelProc()
			<-procDone
			return
		case err := <-procDone:
			// paqet exited unexpectedly
			log.Printf("[Supervisor] paqet exited: %v. Restarting in 5s...", err)
			cancelProc()
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		case <-ticker.C:
			// Interval elapsed: Restart cycle with refreshed MAC
			log.Println("[Supervisor] Interval elapsed. Restarting paqet with updated MAC...")

			// Send graceful SIGTERM
			if cmd.Process != nil {
				_ = cmd.Process.Signal(syscall.SIGTERM)
			}
			// Wait up to 5 seconds for clean exit, force kill if stuck
			select {
			case <-procDone:
			case <-time.After(5 * time.Second):
				cancelProc() // triggers SIGKILL via CommandContext
				<-procDone
			}
			cancelProc()
		}
	}
}
