// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package cfgstruct

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	// AnySource 配置可以来自flag设置或者文件配置
	AnySource = "any"
	// FlagSource 配置可以只来自flag设置，不会写到文件配置
	FlagSource = "flag"

	DefaultsDev     = "dev"
	DefaultsTest    = "test"
	DefaultsRelease = "release"
	Defaults        = DefaultsDev
)

var (
	allSources = []string{
		AnySource,
		FlagSource,
	}
)

// BindOpt is an option for the Bind method.
type BindOpt struct {
	isDev   *bool
	isTest  *bool
	isSetup *bool
	varFn   func(vars map[string]confVar)
}

// ConfigFile 设置配置的环境变量，配置的环境遍历可以从这里读取
func ConfigFile(file string) BindOpt {
	return ConfigVar("CONFIG_FILE", os.ExpandEnv(file))
}

// ConfigVar 设置默认配置选项
func ConfigVar(name, val string) BindOpt {
	name = strings.ToUpper(name)
	return BindOpt{varFn: func(vars map[string]confVar) {
		vars[name] = confVar{val: val, nested: false}
	}}
}

// SetupMode issues the bind in a mode where it does not ignore fields with the
// `setup:"true"` tag.
func SetupMode() BindOpt {
	setup := true
	return BindOpt{isSetup: &setup}
}

// UseDevDefaults forces the bind call to use development defaults unless
// something else is provided as a subsequent option.
// Without a specific defaults setting, Bind will default to determining which
// defaults to use based on version.Build.Release.
func UseDevDefaults() BindOpt {
	dev := true
	test := false
	return BindOpt{isDev: &dev, isTest: &test}
}

// UseReleaseDefaults forces the bind call to use release defaults unless
// something else is provided as a subsequent option.
// Without a specific defaults setting, Bind will default to determining which
// defaults to use based on version.Build.Release.
func UseReleaseDefaults() BindOpt {
	dev := false
	test := false
	return BindOpt{isDev: &dev, isTest: &test}
}

// UseTestDefaults forces the bind call to use test defaults unless
// something else is provided as a subsequent option.
// Without a specific defaults setting, Bind will default to determining which
// defaults to use based on version.Build.Release.
func UseTestDefaults() BindOpt {
	dev := false
	test := true
	return BindOpt{isDev: &dev, isTest: &test}
}

type confVar struct {
	val    string
	nested bool
}

// Bind sets flags on a FlagSet that match the configuration struct
// 'config'. This works by traversing the config struct using the 'reflect'
// package.
func Bind(flags FlagSet, config interface{}, opts ...BindOpt) {
	bind(flags, config, opts...)
}

func bind(flags FlagSet, config interface{}, opts ...BindOpt) {
	ptrType := reflect.TypeOf(config)
	if ptrType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("invalid config type: %#v. Expecting pointer to struct.", config))
	}
	isDev := true
	isTest := false
	setupCommand := false
	vars := map[string]confVar{}
	for _, opt := range opts {
		if opt.varFn != nil {
			opt.varFn(vars)
		}
		if opt.isDev != nil {
			isDev = *opt.isDev
		}
		if opt.isTest != nil {
			isTest = *opt.isTest
		}
		if opt.isSetup != nil {
			setupCommand = *opt.isSetup
		}
	}
	bindConfig(flags, "", reflect.ValueOf(config).Elem(), vars, setupCommand, false, isDev, isTest)
}

func bindConfig(flags FlagSet, prefix string, val reflect.Value, vars map[string]confVar, setupCommand, setupStruct bool, isDev, isTest bool) {
	if val.Kind() != reflect.Struct {
		panic(fmt.Sprintf("invalid config type: %#v. Expecting struct.", val.Interface()))
	}
	typ := val.Type()
	resolvedVars := make(map[string]string, len(vars))
	{
		structPath := strings.ReplaceAll(prefix, ".", string(filepath.Separator))
		for k, v := range vars {
			if !v.nested {
				resolvedVars[k] = v.val
				continue
			}
			resolvedVars[k] = filepath.Join(v.val, structPath)
		}
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		flagName := hyphenate(snakeCase(field.Name))

		if field.Tag.Get("noprefix") != "true" {
			flagName = prefix + flagName
		}

		onlyForSetup := (field.Tag.Get("setup") == "true") || setupStruct
		// ignore setup params for non setup commands
		if !setupCommand && onlyForSetup {
			continue
		}

		if !fieldVal.CanAddr() {
			panic(fmt.Sprintf("cannot addr field %s in %s", field.Name, typ))
		}

		fieldRef := fieldVal.Addr()
		if !fieldRef.CanInterface() {
			panic(fmt.Sprintf("cannot get interface of field %s in %s", field.Name, typ))
		}

		fieldAddr := fieldRef.Interface()
		if fieldValue, ok := fieldAddr.(pflag.Value); ok {
			help := field.Tag.Get("help")
			def := getDefault(field.Tag, isTest, isDev, flagName)

			if field.Tag.Get("internal") == "true" {
				if def != "" {
					panic(fmt.Sprintf("unapplicable default value set for internal flag: %s", flagName))
				}
				continue
			}

			err := fieldValue.Set(def)
			if err != nil {
				panic(fmt.Sprintf("invalid default value for %s: %#v, %v", flagName, def, err))
			}
			flags.Var(fieldValue, flagName, help)

			markHidden := false
			if onlyForSetup {
				SetBoolAnnotation(flags, flagName, "setup", true)
			}
			if field.Tag.Get("user") == "true" {
				SetBoolAnnotation(flags, flagName, "user", true)
			}
			if field.Tag.Get("hidden") == "true" {
				markHidden = true
				SetBoolAnnotation(flags, flagName, "hidden", true)
			}
			if field.Tag.Get("deprecated") == "true" {
				markHidden = true
				SetBoolAnnotation(flags, flagName, "deprecated", true)
			}
			if source := field.Tag.Get("source"); source != "" {
				setSourceAnnotation(flags, flagName, source)
			}
			if markHidden {
				err := flags.MarkHidden(flagName)
				if err != nil {
					panic(fmt.Sprintf("mark hidden failed %s: %v", flagName, err))
				}
			}
			continue
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			if field.Anonymous {
				bindConfig(flags, prefix, fieldVal, vars, setupCommand, onlyForSetup, isDev, isTest)
			} else {
				bindConfig(flags, flagName+".", fieldVal, vars, setupCommand, onlyForSetup, isDev, isTest)
			}
		case reflect.Array:
			digits := len(fmt.Sprint(fieldVal.Len()))
			for j := 0; j < fieldVal.Len(); j++ {
				padding := strings.Repeat("0", digits-len(fmt.Sprint(j)))
				bindConfig(flags, fmt.Sprintf("%s.%s%d.", flagName, padding, j), fieldVal.Index(j), vars, setupCommand, onlyForSetup, isDev, isTest)
			}
		default:
			help := field.Tag.Get("help")
			def := getDefault(field.Tag, isTest, isDev, flagName)

			if field.Tag.Get("internal") == "true" {
				if def != "" {
					panic(fmt.Sprintf("unapplicable default value set for internal flag: %s", flagName))
				}
				continue
			}

			def = expand(resolvedVars, def)

			fieldaddr := fieldVal.Addr().Interface()
			check := func(err error) {
				if err != nil {
					panic(fmt.Sprintf("invalid default value for %s: %#v", flagName, def))
				}
			}
			switch field.Type {
			case reflect.TypeOf(int(0)):
				val, err := strconv.ParseInt(def, 0, strconv.IntSize)
				check(err)
				flags.IntVar(fieldaddr.(*int), flagName, int(val), help)
			case reflect.TypeOf(int64(0)):
				val, err := strconv.ParseInt(def, 0, 64)
				check(err)
				flags.Int64Var(fieldaddr.(*int64), flagName, val, help)
			case reflect.TypeOf(uint(0)):
				val, err := strconv.ParseUint(def, 0, strconv.IntSize)
				check(err)
				flags.UintVar(fieldaddr.(*uint), flagName, uint(val), help)
			case reflect.TypeOf(uint64(0)):
				val, err := strconv.ParseUint(def, 0, 64)
				check(err)
				flags.Uint64Var(fieldaddr.(*uint64), flagName, val, help)
			case reflect.TypeOf(time.Duration(0)):
				val, err := time.ParseDuration(def)
				check(err)
				flags.DurationVar(fieldaddr.(*time.Duration), flagName, val, help)
			case reflect.TypeOf(float64(0)):
				val, err := strconv.ParseFloat(def, 64)
				check(err)
				flags.Float64Var(fieldaddr.(*float64), flagName, val, help)
			case reflect.TypeOf(string("")):
				if field.Tag.Get("path") == "true" {
					// NB: conventionally unix path separators are used in default values
					def = filepath.FromSlash(def)
				}
				flags.StringVar(fieldaddr.(*string), flagName, def, help)
			case reflect.TypeOf(bool(false)):
				val, err := strconv.ParseBool(def)
				check(err)
				flags.BoolVar(fieldaddr.(*bool), flagName, val, help)
			case reflect.TypeOf([]string(nil)):
				// allow either a single string, or comma separated values for defaults
				defaultValues := []string{}
				if def != "" {
					defaultValues = strings.Split(def, ",")
				}
				flags.StringSliceVar(fieldaddr.(*[]string), flagName, defaultValues, help)
			default:
				panic(fmt.Sprintf("invalid field type: %s", field.Type.String()))
			}
			if onlyForSetup {
				SetBoolAnnotation(flags, flagName, "setup", true)
			}
			if field.Tag.Get("user") == "true" {
				SetBoolAnnotation(flags, flagName, "user", true)
			}

			markHidden := false
			if field.Tag.Get("hidden") == "true" {
				markHidden = true
				SetBoolAnnotation(flags, flagName, "hidden", true)
			}
			if field.Tag.Get("deprecated") == "true" {
				markHidden = true
				SetBoolAnnotation(flags, flagName, "deprecated", true)
			}
			if source := field.Tag.Get("source"); source != "" {
				setSourceAnnotation(flags, flagName, source)
			}
			if markHidden {
				err := flags.MarkHidden(flagName)
				if err != nil {
					panic(fmt.Sprintf("mark hidden failed %s: %v", flagName, err))
				}
			}
		}
	}
}

func getDefault(tag reflect.StructTag, isTest, isDev bool, flagname string) string {
	var order []string
	var opposites []string
	if isTest {
		order = []string{"testDefault", "devDefault", "default"}
		opposites = []string{"releaseDefault"}
	} else if isDev {
		order = []string{"devDefault", "default"}
		opposites = []string{"releaseDefault", "testDefault"}
	} else {
		order = []string{"releaseDefault", "default"}
		opposites = []string{"devDefault", "testDefault"}
	}

	for _, name := range order {
		if val, ok := tag.Lookup(name); ok {
			return val
		}
	}

	for _, name := range opposites {
		if _, ok := tag.Lookup(name); ok {
			panic(fmt.Sprintf("%q missing but %q defined for %v", order[0], name, flagname))
		}
	}

	return ""
}

func setSourceAnnotation(flagset interface{}, name, source string) {
	switch source {
	case AnySource:
	case FlagSource:
	default:
		panic(fmt.Sprintf("invalid source annotation %q for %s: must be one of %q", source, name, allSources))
	}

	setStringAnnotation(flagset, name, "source", source)
}

func setStringAnnotation(flagset interface{}, name, key, value string) {
	flags, ok := flagset.(*pflag.FlagSet)
	if !ok {
		return
	}

	err := flags.SetAnnotation(name, key, []string{value})
	if err != nil {
		panic(fmt.Sprintf("unable to set %s annotation for %s: %v", key, name, err))
	}
}

// SetBoolAnnotation sets an annotation (if it can) on flagset with a value of []string{"true|false"}.
func SetBoolAnnotation(flagset interface{}, name, key string, value bool) {
	flags, ok := flagset.(*pflag.FlagSet)
	if !ok {
		return
	}

	err := flags.SetAnnotation(name, key, []string{strconv.FormatBool(value)})
	if err != nil {
		panic(fmt.Sprintf("unable to set %s annotation for %s: %v", key, name, err))
	}
}

func expand(vars map[string]string, val string) string {
	return os.Expand(val, func(key string) string { return vars[key] })
}

// FindConfigFileParam returns '--config' param from os.Args (if exists).
func FindConfigFileParam() string {
	return FindFlagEarly("config")
}

// FindDefaultsParam returns '--defaults' param from os.Args (if it exists).
func FindDefaultsParam() string {
	return FindFlagEarly("defaults")
}

// FindFlagEarly retrieves the value of a flag before `flag.Parse` has been called.
func FindFlagEarly(flagName string) string {
	// workaround to have early access to 'dir' param
	for i, arg := range os.Args {
		if strings.HasPrefix(arg, fmt.Sprintf("--%s=", flagName)) {
			return strings.TrimPrefix(arg, fmt.Sprintf("--%s=", flagName))
		} else if arg == fmt.Sprintf("--%s", flagName) && i < len(os.Args)-1 {
			return os.Args[i+1]
		}
	}
	return ""
}

// SetupFlag sets up flags that are needed before `flag.Parse` has been called.
func SetupFlag(cmd *cobra.Command, dest *string, name, value, usage string) {
	if foundValue := FindFlagEarly(name); foundValue != "" {
		value = foundValue
	}
	cmd.PersistentFlags().StringVar(dest, name, value, usage)
	if cmd.PersistentFlags().SetAnnotation(name, "setup", []string{"true"}) != nil {
		log.Println("Failed to set 'setup' annotation, Flag:", name)
	}
}

// DefaultsType returns the type of defaults (release/dev) this binary should use.
func DefaultsType() string {
	// define a flag so that the flag parsing system will be happy.
	defaults := strings.ToLower(FindDefaultsParam())
	if defaults != "" {
		return defaults
	}
	return Defaults
}

// DefaultsFlag sets up the defaults=dev/release flag options, which is needed
// before `flag.Parse` has been called.
func DefaultsFlag(cmd *cobra.Command) BindOpt {
	// define a flag so that the flag parsing system will be happy.
	defaults := DefaultsType()

	// we're actually going to ignore this flag entirely and parse the commandline
	// arguments early instead
	_ = cmd.PersistentFlags().String("defaults", defaults,
		"determines which set of configuration defaults to use. can either be 'dev' or 'test' or 'release'")
	setSourceAnnotation(cmd.PersistentFlags(), "defaults", FlagSource)

	switch defaults {
	case DefaultsDev:
		return UseDevDefaults()
	case DefaultsRelease:
		return UseReleaseDefaults()
	case DefaultsTest:
		return UseTestDefaults()
	default:
		panic(fmt.Sprintf("unsupported defaults value %q", FindDefaultsParam()))
	}
}
