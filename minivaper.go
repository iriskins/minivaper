package minivaper

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/ini.v1"
)

var cm *ConfigManager

func init() {
	cm = New()
}

var SupportedTypes = []string{"ini"}

var FileExtensionsOfTypes = map[string][]string{
	"ini": []string{"ini"},
	"yaml": []string{"yml", "yaml"},
}

type ConfigManager struct { configName	string
	configType	string
	config	map[string]string
	configPaths	[]string
}

func New() *ConfigManager {
	cm = new(ConfigManager)
	cm.configName = "sso_config"
	cm.configType = "ini"
	cm.config = make(map[string]string)
	return cm
}

func ReadConfig() error {return cm.ReadConfig()}

func (m *ConfigManager) ReadConfig() error {
	if !slices.Contains(SupportedTypes, m.configType) {
		return fmt.Errorf("not supported config file type")
	}

	switch m.configType {
	case "ini":
		err := m.readIniConfig()
		if err != nil {
			return err
		}
	}

	return nil
}

// Return a full path to configure file
func (m *ConfigManager) finder() (string, error) {
	var configPaths []string
	for _, cp := range m.configPaths {
		f, err := os.Stat(cp)
		if err == nil && f.IsDir() {
			acp, err := filepath.Abs(cp)
			if err == nil {
				configPaths = append(configPaths, acp)
			}
		}
	}

	if len(configPaths) == 0 {
		return "", fmt.Errorf("cannot find any config directories: %v", m.configPaths)
	}

	configName := m.configName
	if ext := filepath.Ext(configName); ext != "" {
		configName = strings.TrimSuffix(configName, ext)
	}
	
	var configPathsWithFileNames []string

	for _, fp := range configPaths {
		configPathsWithFileNames = append(configPathsWithFileNames, filepath.Join(fp, configName))
	}

	extensions := m.getConfigExt()
	for _, ext := range extensions {
		for _, cpwfn := range configPathsWithFileNames {
			fullpath := fmt.Sprintf("%s.%s", cpwfn, ext)
			if _, err := os.Stat(fullpath); err == nil {
				return fullpath, nil
			}
		}
	}

    return "", fmt.Errorf("cannot find config file %s with extensions %v in %s", configName, extensions, configPaths)
}

func (m *ConfigManager) readIniConfig() error {
	c, err := m.finder()
	if err != nil {
		return err
	}

	f, err := ini.Load(c)
	if err != nil {
		return err
	}
	
	for _, s := range f.Sections() {
		if s.Name() == "DEFAULT" {
			for _, k := range s.Keys() {
				m.config[k.Name()] = k.Value()
			}
		}
		for _, k := range s.Keys() {
			m.config[fmt.Sprintf("%v.%v", s.Name(), k.Name())] = k.Value()
		}
	}
	
	return nil
}

func (m *ConfigManager) getConfigExt() []string {
	return FileExtensionsOfTypes[m.configType]
}

func Get(key string) any {return cm.Get(key)}

func (m *ConfigManager) Get(key string) any {
	return m.config[key]
}

func SetConfigPath(basepath string) { cm.SetConfigPath(basepath) }

func (m *ConfigManager) SetConfigPath(basepath string) {
	m.configPaths = append(
		m.configPaths,
		basepath,
	)
}

func SetConfigFileName(name string) {cm.SetConfigName(name)}

func (m *ConfigManager) SetConfigName(name string) {
	m.configName = name
}

func SetConfigFileType(cftype string) {cm.SetConfigExt(cftype)}

func (m *ConfigManager) SetConfigExt(cftype string) {
	m.configType=cftype
}
