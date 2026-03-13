package process

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"rest_demo/pkg/cfgstruct"
	"sort"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zeebo/errs"
	"gopkg.in/yaml.v3"
)

var DefaultCfgFilename = "config.yaml"

// SaveConfigOption is a function that updates the options for SaveConfig.
type SaveConfigOption func(*SaveConfigOptions)

// SaveConfigOptions controls the behavior of SaveConfig.
type SaveConfigOptions struct {
	Overrides        map[string]interface{}
	RemoveDeprecated bool
}

// SaveConfigWithOverrides sets the overrides to the provided map.
func SaveConfigWithOverrides(overrides map[string]interface{}) SaveConfigOption {
	return func(opts *SaveConfigOptions) {
		opts.Overrides = overrides
	}
}

// SaveConfigWithOverride adds a single override to SaveConfig.
func SaveConfigWithOverride(name string, value interface{}) SaveConfigOption {
	return func(opts *SaveConfigOptions) {
		if opts.Overrides == nil {
			opts.Overrides = make(map[string]interface{})
		}
		opts.Overrides[name] = value
	}
}

// SaveConfigRemovingDeprecated tells SaveConfig to not store deprecated flags.
func SaveConfigRemovingDeprecated() SaveConfigOption {
	return func(opts *SaveConfigOptions) {
		opts.RemoveDeprecated = true
	}
}

// SaveConfig will save only the user-specific flags with default values to
// outfile with specific values specified in 'overrides' overridden.
func SaveConfig(cmd *cobra.Command, outfile string, opts ...SaveConfigOption) error {
	// 1. 设置要替换的参数选贤
	var options SaveConfigOptions
	for _, opt := range opts {
		opt(&options)
	}

	// 2. 读取所有的设置准备保存
	flags := cmd.Flags()
	vip, err := Viper(cmd)
	if err != nil {
		return errs.Wrap(err)
	}
	if err := vip.MergeConfigMap(options.Overrides); err != nil {
		return errs.Wrap(err)
	}
	settings := vip.AllSettings()

	// 3. 对比准备好要保存到config 的配置
	type configValue struct {
		value   interface{}
		comment string
		set     bool
	}
	flat := make(map[string]configValue)
	flatKeys := make([]string, 0)

	// 递归处理方法.
	var filterAndFlatten func(string, map[string]interface{})
	filterAndFlatten = func(base string, settings map[string]interface{}) {
		for key, value := range settings {
			if value, ok := value.(map[string]interface{}); ok {
				filterAndFlatten(base+key+".", value)
				continue
			}
			fullKey := base + key

			// help排除
			if fullKey == "help" {
				continue
			}

			//手机标志信息
			var (
				changed    bool
				setup      bool
				hidden     bool
				user       bool
				deprecated bool
				source     string
				comment    string
				typ        string

				_, overrideExists = options.Overrides[fullKey]
			)
			if f := flags.Lookup(fullKey); f != nil {
				//flag设置的值如果不是默认值，则认为修改，以设置值为准
				changed = f.Changed || f.Value.String() != f.DefValue
				setup = readBoolAnnotation(f, "setup")
				hidden = readBoolAnnotation(f, "hidden")
				user = readBoolAnnotation(f, "user")
				deprecated = readBoolAnnotation(f, "deprecated")
				source = readSourceAnnotation(f)
				comment = f.Usage
				typ = f.Value.Type()
			} else if f := flag.Lookup(fullKey); f != nil {
				// then stdlib flags
				changed = f.Value.String() != f.DefValue
				comment = f.Usage
			} else {
				// 默认都是改变的，以便写入配置文件
				changed = true
			}

			// 如果符合下面一些条件，则不将该配置写入文件
			if setup ||
				hidden ||
				options.RemoveDeprecated && deprecated ||
				source == cfgstruct.FlagSource {
				continue
			}

			// viper需要对float的特殊处理
			if typ == "float64" {
				value = cast.ToFloat64(value)
			}

			flatKeys = append(flatKeys, fullKey)
			flat[fullKey] = configValue{
				value:   value,
				comment: comment,
				set:     user || changed || overrideExists,
			}
		}
	}
	filterAndFlatten("", settings)
	sort.Strings(flatKeys)

	// 4. 写入文件
	var nl = []byte("\n")
	var lines [][]byte
	for _, key := range flatKeys {
		config := flat[key]

		if config.comment != "" {
			lines = append(lines, []byte("# "+config.comment))
		}

		data, err := yaml.Marshal(map[string]interface{}{key: config.value})
		if err != nil {
			return errs.Wrap(err)
		}
		dataLines := bytes.Split(bytes.TrimSpace(data), nl)

		if config.set {
			lines = append(lines, dataLines...)
		} else {
			// 没修改的注释掉
			for _, line := range dataLines {
				lines = append(lines, append([]byte("# "), line...))
			}
		}

		// 空行分割
		lines = append(lines, nil)
	}

	return errs.Wrap(AtomicWriteFile(outfile, bytes.Join(lines, nl), 0600))
}

// readSourceAnnotation 获取一个flag的 annotation 的source设置，如果没有，返回 AnySource
func readSourceAnnotation(flag *pflag.Flag) string {
	annotation := flag.Annotations["source"]
	if len(annotation) == 0 {
		return cfgstruct.AnySource
	}
	return annotation[0]
}

// readBoolAnnotation 检查一个flag的annotation的key值bool设置.
func readBoolAnnotation(flag *pflag.Flag, key string) bool {
	annotation := flag.Annotations[key]
	return len(annotation) > 0 && annotation[0] == "true"
}

// AtomicWriteFile is a helper to atomically write the data to the outfile.
func AtomicWriteFile(outfile string, data []byte, _ os.FileMode) (err error) {
	// directory and, on windows, using MoveFileEx with MOVEFILE_WRITE_THROUGH.
	dir := filepath.Dir(outfile)
	ds, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(dir, 0766); err != nil {
			return err
		}
	} else {
		if !ds.IsDir() {
			return fmt.Errorf("path %s is not dir", dir)
		}
	}

	fh, err := os.CreateTemp(filepath.Dir(outfile), filepath.Base(outfile))
	if err != nil {
		return errs.Wrap(err)
	}
	needsClose, needsRemove := true, true
	defer func() {
		if needsClose {
			err = errs.Combine(err, errs.Wrap(fh.Close()))
		}
		if needsRemove {
			err = errs.Combine(err, errs.Wrap(os.Remove(fh.Name())))
		}
	}()
	if _, err := fh.Write(data); err != nil {
		return errs.Wrap(err)
	}
	needsClose = false
	if err := fh.Close(); err != nil {
		return errs.Wrap(err)
	}
	if err := os.Rename(fh.Name(), outfile); err != nil {
		return errs.Wrap(err)
	}
	needsRemove = false
	return nil
}
