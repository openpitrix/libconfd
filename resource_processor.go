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
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type TemplateResourceProcessor struct {
	TemplateResource

	stageFile     *os.File
	funcMap       template.FuncMap
	lastIndex     uint64
	keepStageFile bool
	noop          bool
	store         *KVStore
	storeClient   Client
	syncOnly      bool
}

func MakeAllTemplateResourceProcessor(
	config Config, client Client,
) (
	[]*TemplateResourceProcessor,
	error,
) {
	var lastError error
	templates := make([]*TemplateResourceProcessor, 0)
	logger.Debug("Loading template resources from confdir " + config.ConfDir)
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
		logger.Debugf("Found template: %s", p)
		t, err := NewTemplateResourceProcessor(p, config, client)
		if err != nil {
			lastError = err
			continue
		}
		templates = append(templates, t)
	}
	return templates, lastError
}

// NewTemplateResourceProcessor creates a NewTemplateResourceProcessor.
func NewTemplateResourceProcessor(
	path string, config Config, client Client,
) (
	*TemplateResourceProcessor,
	error,
) {
	logger.Debug("Loading template resource from " + path)

	res, err := LoadTemplateResourceFile(path)
	if err != nil {
		return nil, fmt.Errorf("Cannot process template resource %s - %v", path, err)
	}

	tr := TemplateResourceProcessor{
		TemplateResource: *res,
	}
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

	_, tr.funcMap = NewTemplateFuncMap(tr.store, tr.PGPPrivateKey)
	tr.Src = filepath.Join(config.TemplateDir, tr.Src)

	return &tr, nil
}

// setVars sets the Vars for template resource.
func (t *TemplateResourceProcessor) SetVars() error {
	var err error
	logger.Debug("Retrieving keys from store")
	logger.Debug("Key prefix set to " + t.Prefix)
	result, err := t.storeClient.GetValues(utilAppendPrefix(t.Prefix, t.Keys))
	if err != nil {
		return err
	}
	logger.Debug("Got the following map from store: %v", result)
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
func (t *TemplateResourceProcessor) CreateStageFile() error {
	logger.Debug("Using source template " + t.Src)
	if !utilFileExist(t.Src) {
		return errors.New("Missing template: " + t.Src)
	}

	logger.Debug("Compiling source template " + t.Src)
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
	t.stageFile = temp
	return nil
}

// sync compares the staged and dest config files and attempts to sync them
// if they differ. sync will run a config check command if set before
// overwriting the target config file. Finally, sync will run a reload command
// if set to have the application or service pick up the changes.
// It returns an error if any.
func (t *TemplateResourceProcessor) Sync() error {
	staged := t.stageFile.Name()
	if t.keepStageFile {
		logger.Info("Keeping staged file: " + staged)
	} else {
		defer os.Remove(staged)
	}

	logger.Debug("Comparing candidate config to " + t.Dest)
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
			if err := t.Check(); err != nil {
				return fmt.Errorf("Config check failed: %v", err)
			}
		}
		logger.Debug("Overwriting target config " + t.Dest)
		err := os.Rename(staged, t.Dest)
		if err != nil {
			if notDeviceOrResourceBusyError(err) {
				return err
			}

			logger.Debug("Rename failed - target is likely a mount. Trying to write instead")
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
			if err := t.Reload(); err != nil {
				return err
			}
		}
		logger.Info("Target config " + t.Dest + " has been updated")
	} else {
		logger.Debug("Target config " + t.Dest + " in sync")
	}
	return nil
}

// check executes the check command to validate the staged config file. The
// command is modified so that any references to src template are substituted
// with a string representing the full path of the staged file. This allows the
// check to be run on the staged file before overwriting the destination config
// file.
// It returns nil if the check command returns 0 and there are no other errors.
func (t *TemplateResourceProcessor) Check() error {
	var cmdBuffer bytes.Buffer
	data := make(map[string]string)
	data["src"] = t.stageFile.Name()
	tmpl, err := template.New("checkcmd").Parse(t.CheckCmd)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&cmdBuffer, data); err != nil {
		return err
	}
	return utilRunCommand(cmdBuffer.String())
}

// reload executes the reload command.
// It returns nil if the reload command returns 0.
func (t *TemplateResourceProcessor) Reload() error {
	return utilRunCommand(t.ReloadCmd)
}

// process is a convenience function that wraps calls to the three main tasks
// required to keep local configuration files in sync. First we gather vars
// from the store, then we stage a candidate configuration file, and finally sync
// things up.
// It returns an error if any.
func (t *TemplateResourceProcessor) Process() error {
	if err := t.SetFileMode(); err != nil {
		return err
	}
	if err := t.SetVars(); err != nil {
		return err
	}
	if err := t.CreateStageFile(); err != nil {
		return err
	}
	if err := t.Sync(); err != nil {
		return err
	}
	return nil
}

// setFileMode sets the FileMode.
func (t *TemplateResourceProcessor) SetFileMode() error {
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
