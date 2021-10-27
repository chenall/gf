package gcfg

import (
	"fmt"
)

const NoExistsConfigFile = "NoExistsFile"

// FindConfig 判断配置文件是否存在，如果存在返回配置，否则返回 nil 可以指定多个备用配置文件.
func FindConfig(name string, file ...string) *Config {
	key := name
	if name == "" {
		for _, n := range file {
			if n != "" {
				key = n
				break
			}
		}
	}
	if key == "" {
		return nil
	}
	cfg := instances.GetOrSetFuncLock(key, func() interface{} {
		ConfigFile := name + ".toml"
		c := New(ConfigFile)
		if !c.Available() {
			c.SetFileName(NoExistsConfigFile)
		}
		return c
	}).(*Config)

	if cfg.GetFileName() != NoExistsConfigFile {
		return cfg
	}

	for _, name := range append(file, name) {
		if name == "" {
			continue
		}
		for _, fileType := range supportedFileTypes {
			if file := fmt.Sprintf(`%s.%s`, name, fileType); cfg.Available(file) {
				cfg.SetFileName(file)
				return cfg
			}
		}
	}
	return nil
}
