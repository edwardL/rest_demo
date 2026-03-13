package process

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"rest_demo/pkg/cfgstruct"
	"strings"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/zeebo/structs"
)

var (
	commandMtx sync.Mutex
	contexts   = map[*cobra.Command]context.Context{}
	cancels    = map[*cobra.Command]context.CancelFunc{}
	configs    = map[*cobra.Command][]interface{}{}
	vipers     = map[*cobra.Command]*viper.Viper{}
)

func Bind(cmd *cobra.Command, config interface{}, opts ...cfgstruct.BindOpt) {
	commandMtx.Lock()
	defer commandMtx.Unlock()

	cfgstruct.Bind(cmd.Flags(), config, opts...)
	configs[cmd] = append(configs[cmd], config)
}

func Viper(cmd *cobra.Command) (*viper.Viper, error) {
	return ViperWithCustomConfig(cmd, LoadConfig)
}

// ViperWithCustomConfig returns the appropriate *viper.Viper for the command, creating if necessary. Custom
// config load logic can be defined with "loadConfig" parameter.
func ViperWithCustomConfig(cmd *cobra.Command, loadConfig func(cmd *cobra.Command, vip *viper.Viper) error) (*viper.Viper, error) {
	commandMtx.Lock()
	defer commandMtx.Unlock()

	if vip := vipers[cmd]; vip != nil {
		return vip, nil
	}

	vip := viper.New()
	if err := vip.BindPFlags(cmd.Flags()); err != nil {
		return nil, err
	}

	prefix := os.Getenv("GO_ENV_PREFIX")
	if prefix == "" {
		prefix = "GE"
	}

	vip.SetEnvPrefix(prefix)
	vip.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	vip.AutomaticEnv()

	err := loadConfig(cmd, vip)
	if err != nil {
		return nil, err
	}

	vipers[cmd] = vip
	return vip, nil
}

// LoadConfig 从配置的--config 配置文件读取配置到viper
func LoadConfig(cmd *cobra.Command, vip *viper.Viper) error {
	cfgFlag := cmd.Flags().Lookup("config")
	if cfgFlag != nil && cfgFlag.Value.String() != "" {
		path := os.ExpandEnv(cfgFlag.Value.String())
		if FileExists(path) {
			setupCommand := cmd.Annotations["type"] == "setup"
			vip.SetConfigFile(path)
			if err := vip.ReadInConfig(); err != nil && !setupCommand {
				return err
			}
		}
	}
	return nil
}

// Ctx returns the appropriate context.Context for ExecuteWithConfig commands.
func Ctx(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	commandMtx.Lock()
	defer commandMtx.Unlock()

	ctx := contexts[cmd]
	if ctx == nil {
		ctx = context.Background()
		contexts[cmd] = ctx
	}

	cancel := cancels[cmd]
	if cancel == nil {
		ctx, cancel = context.WithCancel(ctx)
		contexts[cmd] = ctx
		cancels[cmd] = cancel

		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-c
			log.Printf("Got a signal from the OS: %q", sig)
			signal.Stop(c)
			cancel()
		}()
	}

	return ctx, cancel
}

func Exec(cmd *cobra.Command) {
	ExecWithCustomConfig(cmd, LoadConfig)
}

// ExecWithCustomConfig runs a Cobra command. Custom configuration can be loaded.
func ExecWithCustomConfig(cmd *cobra.Command, loadConfig func(cmd *cobra.Command, vip *viper.Viper) error) {
	exe, err := os.Executable()
	if err == nil && cmd.Use == "" {
		cmd.Use = exe
	}

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	cleanup(cmd, loadConfig)
	err = cmd.Execute()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func cleanup(cmd *cobra.Command, loadConfig func(cmd *cobra.Command, vip *viper.Viper) error) {
	for _, ccmd := range cmd.Commands() {
		cleanup(ccmd, loadConfig)
	}
	if cmd.Run != nil {
		panic("Please use cobra's RunE instead of Run")
	}
	internalRun := cmd.RunE
	if internalRun == nil {
		return
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		vip, err := ViperWithCustomConfig(cmd, loadConfig)
		if err != nil {
			return err
		}

		commandMtx.Lock()
		configValues := configs[cmd]
		commandMtx.Unlock()

		var (
			brokenKeys  = map[string]struct{}{}
			missingKeys = map[string]struct{}{}
			usedKeys    = map[string]struct{}{}
			allKeys     = map[string]struct{}{}
			allSettings = vip.AllSettings()
		)

		for _, config := range configValues {
			res := structs.Decode(allSettings, config)
			for key := range res.Used {
				usedKeys[key] = struct{}{}
				allKeys[key] = struct{}{}
			}
			for key := range res.Missing {
				missingKeys[key] = struct{}{}
				allKeys[key] = struct{}{}
			}
			for key := range res.Broken {
				brokenKeys[key] = struct{}{}
				allKeys[key] = struct{}{}
			}
		}

		// Propagate keys that are missing to flags, and remove any used keys
		// from the missing set.
		for key := range missingKeys {
			if f := cmd.Flags().Lookup(key); f != nil {
				val := vip.GetString(key)
				err := f.Value.Set(val)
				f.Changed = val != f.DefValue
				if err != nil {
					brokenKeys[key] = struct{}{}
				} else {
					usedKeys[key] = struct{}{}
				}
			} else if f := flag.Lookup(key); f != nil {
				err := f.Value.Set(vip.GetString(key))
				if err != nil {
					brokenKeys[key] = struct{}{}
				} else {
					usedKeys[key] = struct{}{}
				}
			}
		}
		for key := range missingKeys {
			if _, ok := usedKeys[key]; ok {
				delete(missingKeys, key)
			}
		}

		if vip.ConfigFileUsed() != "" {
			_, err = filepath.Abs(vip.ConfigFileUsed())
			if err != nil {
				_ = vip.ConfigFileUsed()
				//log.Println("unable to resolve path: ", err)
			}
			//log.Println("Configuration loaded: ", path)
		}

		work := func(ctx context.Context) error {
			commandMtx.Lock()
			contexts[cmd] = ctx
			commandMtx.Unlock()
			defer func() {
				commandMtx.Lock()
				delete(contexts, cmd)
				delete(cancels, cmd)
				commandMtx.Unlock()
			}()
			return internalRun(cmd, args)
		}
		return work(ctx)
	}
}

// FileExists checks whether file exists, handle error correctly if it doesn't.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Fatalf("failed to check for file existence: %v", err)
	}
	return true
}
