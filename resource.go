// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
)

// TemplateResourceConfig holds the parsed template resource.
type TemplateResourceConfig struct {
	TemplateResource TemplateResource `toml:"template"`
}

// TemplateResource is the representation of a parsed template resource.
type TemplateResource struct {
	CheckCmd      string `toml:"check_cmd"`
	Dest          string
	FileMode      os.FileMode
	Gid           int
	Keys          []string
	Mode          string
	Prefix        string
	ReloadCmd     string `toml:"reload_cmd"`
	Src           string
	StageFile     *os.File
	Uid           int
	funcMap       template.FuncMap
	lastIndex     uint64
	keepStageFile bool
	noop          bool
	store         *KVStore
	storeClient   StoreClient
	syncOnly      bool
	PGPPrivateKey []byte
}

func MakeTemplateResourceList(config Config, client StoreClient) ([]*TemplateResource, error) {
	var lastError error
	templates := make([]*TemplateResource, 0)
	if logger.V(1) {
		logger.Info("Loading template resources from confdir " + config.ConfDir)
	}
	if !utilFileExist(config.ConfDir) {
		logger.Warning(fmt.Sprintf("Cannot load template resources: confdir '%s' does not exist", config.ConfDir))
		return nil, nil
	}
	paths, err := utilRecursiveFindFiles(config.ConfigDir, "*toml")
	if err != nil {
		return nil, err
	}

	if len(paths) < 1 {
		logger.Warning("Found no templates")
	}

	for _, p := range paths {
		if logger.V(1) {
			logger.Infof("Found template: %s", p)
		}
		t, err := NewTemplateResource(p, config, client)
		if err != nil {
			lastError = err
			continue
		}
		templates = append(templates, t)
	}
	return templates, lastError
}

// NewTemplateResource creates a TemplateResource.
func NewTemplateResource(path string, config Config, client StoreClient) (*TemplateResource, error) {
	// Set the default uid and gid so we can determine if it was
	// unset from configuration.
	tc := &TemplateResourceConfig{TemplateResource{Uid: -1, Gid: -1}}

	if logger.V(1) {
		logger.Info("Loading template resource from " + path)
	}

	_, err := toml.DecodeFile(path, &tc)
	if err != nil {
		return nil, fmt.Errorf("Cannot process template resource %s - %v", path, err)
	}

	tr := tc.TemplateResource
	tr.keepStageFile = config.KeepStageFile
	tr.noop = config.Noop
	tr.storeClient = client
	tr.store = NewKVStore()
	tr.syncOnly = config.SyncOnly

	if config.Prefix != "" {
		tr.Prefix = config.Prefix
	}

	if !strings.HasPrefix(tr.Prefix, "/") {
		tr.Prefix = "/" + tr.Prefix
	}

	if len(config.PGPPrivateKey) > 0 {
		tr.PGPPrivateKey = config.PGPPrivateKey
	}

	if tr.Src == "" {
		return nil, ErrEmptySrc
	}

	if tr.Uid == -1 {
		tr.Uid = os.Geteuid()
	}

	if tr.Gid == -1 {
		tr.Gid = os.Getegid()
	}

	_, tr.funcMap = MakeTemplateFuncMap(tr.store, tr.PGPPrivateKey)
	tr.Src = filepath.Join(config.TemplateDir, tr.Src)

	return &tr, nil
}

// setVars sets the Vars for template resource.
func (t *TemplateResource) setVars() error {
	var err error
	if logger.V(1) {
		logger.Info("Retrieving keys from store")
		logger.Info("Key prefix set to " + t.Prefix)
	}
	result, err := t.storeClient.GetValues(utilAppendPrefix(t.Prefix, t.Keys))
	if err != nil {
		return err
	}
	if logger.V(1) {
		logger.Info("Got the following map from store: %v", result)
	}
	t.store.Purge()

	for k, v := range result {
		t.store.Set(path.Join("/", strings.TrimPrefix(k, t.Prefix)), v)
	}
	return nil
}

// createStageFile stages the src configuration file by processing the src
// template and setting the desired owner, group, and mode. It also sets the
// StageFile for the template resource.
// It returns an error if any.
func (t *TemplateResource) createStageFile() error {
	if logger.V(1) {
		logger.Info("Using source template " + t.Src)
	}
	if !utilFileExist(t.Src) {
		return errors.New("Missing template: " + t.Src)
	}

	if logger.V(1) {
		logger.Info("Compiling source template " + t.Src)
	}
	tmpl, err := template.New(filepath.Base(t.Src)).Funcs(template.FuncMap(t.funcMap)).ParseFiles(t.Src)
	if err != nil {
		return fmt.Errorf("Unable to process template %s, %s", t.Src, err)
	}

	// create TempFile in Dest directory to avoid cross-filesystem issues
	temp, err := ioutil.TempFile(filepath.Dir(t.Dest), "."+filepath.Base(t.Dest))
	if err != nil {
		return err
	}

	if err = tmpl.Execute(temp, nil); err != nil {
		temp.Close()
		os.Remove(temp.Name())
		return err
	}
	defer temp.Close()

	// Set the owner, group, and mode on the stage file now to make it easier to
	// compare against the destination configuration file later.
	os.Chmod(temp.Name(), t.FileMode)
	os.Chown(temp.Name(), t.Uid, t.Gid)
	t.StageFile = temp
	return nil
}

// sync compares the staged and dest config files and attempts to sync them
// if they differ. sync will run a config check command if set before
// overwriting the target config file. Finally, sync will run a reload command
// if set to have the application or service pick up the changes.
// It returns an error if any.
func (t *TemplateResource) sync() error {
	staged := t.StageFile.Name()
	if t.keepStageFile {
		logger.Info("Keeping staged file: " + staged)
	} else {
		defer os.Remove(staged)
	}

	if logger.V(1) {
		logger.Info("Comparing candidate config to " + t.Dest)
	}
	ok, err := utilSameConfig(staged, t.Dest)
	if err != nil {
		logger.Error(err)
	}
	if t.noop {
		logger.Warning("Noop mode enabled. " + t.Dest + " will not be modified")
		return nil
	}
	if !ok {
		logger.Info("Target config " + t.Dest + " out of sync")
		if !t.syncOnly && t.CheckCmd != "" {
			if err := t.check(); err != nil {
				return fmt.Errorf("Config check failed: %v", err)
			}
		}
		if logger.V(1) {
			logger.Info("Overwriting target config " + t.Dest)
		}
		err := os.Rename(staged, t.Dest)
		if err != nil {
			if notDeviceOrResourceBusyError(err) {
				return err
			}

			if logger.V(1) {
				logger.Info("Rename failed - target is likely a mount. Trying to write instead")
			}
			// try to open the file and write to it
			var contents []byte
			var rerr error
			contents, rerr = ioutil.ReadFile(staged)
			if rerr != nil {
				return rerr
			}
			err := ioutil.WriteFile(t.Dest, contents, t.FileMode)
			// make sure owner and group match the temp file, in case the file was created with WriteFile
			os.Chown(t.Dest, t.Uid, t.Gid)
			if err != nil {
				return err
			}
		}
		if !t.syncOnly && t.ReloadCmd != "" {
			if err := t.reload(); err != nil {
				return err
			}
		}
		logger.Info("Target config " + t.Dest + " has been updated")
	} else {
		if logger.V(1) {
			logger.Info("Target config " + t.Dest + " in sync")
		}
	}
	return nil
}

// check executes the check command to validate the staged config file. The
// command is modified so that any references to src template are substituted
// with a string representing the full path of the staged file. This allows the
// check to be run on the staged file before overwriting the destination config
// file.
// It returns nil if the check command returns 0 and there are no other errors.
func (t *TemplateResource) check() error {
	var cmdBuffer bytes.Buffer
	data := make(map[string]string)
	data["src"] = t.StageFile.Name()
	tmpl, err := template.New("checkcmd").Parse(t.CheckCmd)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&cmdBuffer, data); err != nil {
		return err
	}
	return runCommand(cmdBuffer.String())
}

// reload executes the reload command.
// It returns nil if the reload command returns 0.
func (t *TemplateResource) reload() error {
	return runCommand(t.ReloadCmd)
}

// runCommand is a shared function used by check and reload
// to run the given command and log its output.
// It returns nil if the given cmd returns 0.
// The command can be run on unix and windows.
func runCommand(cmd string) error {
	if logger.V(1) {
		logger.Info("Running " + cmd)
	}
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/C", cmd)
	} else {
		c = exec.Command("/bin/sh", "-c", cmd)
	}

	output, err := c.CombinedOutput()
	if err != nil {
		logger.Error(fmt.Sprintf("%q", string(output)))
		return err
	}
	if logger.V(1) {
		logger.Infof("%q", string(output))
	}
	return nil
}

// process is a convenience function that wraps calls to the three main tasks
// required to keep local configuration files in sync. First we gather vars
// from the store, then we stage a candidate configuration file, and finally sync
// things up.
// It returns an error if any.
func (t *TemplateResource) process() error {
	if err := t.setFileMode(); err != nil {
		return err
	}
	if err := t.setVars(); err != nil {
		return err
	}
	if err := t.createStageFile(); err != nil {
		return err
	}
	if err := t.sync(); err != nil {
		return err
	}
	return nil
}

// setFileMode sets the FileMode.
func (t *TemplateResource) setFileMode() error {
	if t.Mode == "" {
		if !utilFileExist(t.Dest) {
			t.FileMode = 0644
		} else {
			fi, err := os.Stat(t.Dest)
			if err != nil {
				return err
			}
			t.FileMode = fi.Mode()
		}
	} else {
		mode, err := strconv.ParseUint(t.Mode, 0, 32)
		if err != nil {
			return err
		}
		t.FileMode = os.FileMode(mode)
	}
	return nil
}
