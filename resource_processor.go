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
)

type TemplateResourceProcessor struct {
	TemplateResource

	client        Client
	store         *KVStore
	stageFile     *os.File
	funcMap       template.FuncMap
	keepStageFile bool
	lastIndex     uint64
	syncOnly      bool
	noop          bool
}

func MakeAllTemplateResourceProcessor(
	config *Config, client Client,
) (
	[]*TemplateResourceProcessor,
	error,
) {
	logger.Debug("Loading template resources from confdir " + config.ConfDir)

	if fileNotExists(config.ConfDir) {
		logger.Warning(fmt.Sprintf("Cannot load template resources: confdir '%s' does not exist", config.ConfDir))
		return nil, fmt.Errorf("confdir '%s' does not exist", config.ConfDir)
	}

	paths, err := findFilesRecursive(config.GetConfigDir(), "*toml")
	if err != nil {
		logger.Warning("findFilesRecursive(%q, %q): %v", config.GetConfigDir(), "*toml", err)
		return nil, err
	}

	if len(paths) == 0 {
		logger.Warning("Found no templates")
		return nil, fmt.Errorf("Found no templates")
	}

	var lastError error
	var templates = make([]*TemplateResourceProcessor, 0)

	for _, p := range paths {
		logger.Debugf("Found template: %s", p)

		t, err := NewTemplateResourceProcessor(p, config, client)
		if err != nil {
			logger.Error(err)
			lastError = err
			continue
		}

		templates = append(templates, t)
	}
	if lastError != nil {
		return templates, lastError
	}

	return templates, nil
}

// NewTemplateResourceProcessor creates a NewTemplateResourceProcessor.
func NewTemplateResourceProcessor(
	path string, config *Config, client Client,
) (
	*TemplateResourceProcessor,
	error,
) {
	logger.Debug("Loading template resource from " + path)

	res, err := LoadTemplateResourceFile(path)
	if err != nil {
		logger.Warning(err)
		return nil, fmt.Errorf("Cannot process template resource %s - %v", path, err)
	}

	tr := TemplateResourceProcessor{
		TemplateResource: *res,
	}

	tr.client = client
	tr.store = NewKVStore()
	tr.keepStageFile = config.KeepStageFile
	tr.syncOnly = config.SyncOnly
	tr.noop = config.Noop

	if config.ConfDir != "" {
		if s := tr.Dest; !filepath.IsAbs(s) {
			config.makeTemplateDir()
			tr.Dest = filepath.Join(config.GetTemplateDir(), s)
			tr.Dest = filepath.Clean(tr.Dest)
		}
	}

	if config.Prefix != "" {
		tr.Prefix = config.Prefix
	}

	if !strings.HasPrefix(tr.Prefix, "/") {
		tr.Prefix = "/" + tr.Prefix
	}

	if len(config.PGPPrivateKey) > 0 {
		tr.PGPPrivateKey = append([]byte{}, config.PGPPrivateKey...)
	}

	if tr.Src == "" {
		return nil, errors.New("libconfd: empty src template")
	}

	if tr.Uid == -1 {
		tr.Uid = os.Geteuid()
	}

	if tr.Gid == -1 {
		tr.Gid = os.Getegid()
	}

	tr.funcMap = NewTemplateFunc(tr.store, tr.PGPPrivateKey).FuncMap
	tr.Src = filepath.Join(config.GetTemplateDir(), tr.Src)

	return &tr, nil
}

// process is a convenience function that wraps calls to the three main tasks
// required to keep local configuration files in sync. First we gather vars
// from the store, then we stage a candidate configuration file, and finally sync
// things up.
// It returns an error if any.
func (p *TemplateResourceProcessor) Process(opts ...Options) error {
	opt := newOptions(opts...)

	if len(opt.funcMap) > 0 {
		for k, fn := range opt.funcMap {
			p.funcMap[k] = fn
		}
	}
	if len(opt.funcMapUpdater) > 0 {
		for _, fn := range opt.funcMapUpdater {
			fn(p.funcMap)
		}
	}

	if err := p.setFileMode(opt); err != nil {
		logger.Error(err)
		return err
	}
	if err := p.setVars(opt); err != nil {
		logger.Error(err)
		return err
	}
	if err := p.createStageFile(opt); err != nil {
		logger.Error(err)
		return err
	}
	if err := p.sync(opt); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

// setFileMode sets the FileMode.
func (p *TemplateResourceProcessor) setFileMode(opt *options) error {
	if p.Mode == "" {
		if fi, err := os.Stat(p.Dest); err == nil {
			p.FileMode = fi.Mode()
		} else {
			p.FileMode = 0644
		}
	} else {
		mode, err := strconv.ParseUint(p.Mode, 0, 32)
		if err != nil {
			return err
		}
		p.FileMode = os.FileMode(mode)
	}
	return nil
}

// setVars sets the Vars for template resource.
func (p *TemplateResourceProcessor) setVars(opt *options) error {
	var err error

	logger.Debug("Retrieving keys from store")
	logger.Debug("Key prefix set to " + p.Prefix)

	values, err := p.client.GetValues(p.getAbsKeys())
	if err != nil {
		return err
	}

	logger.Debug("Got the following map from store: %v", values)

	p.store.Purge()
	for k, v := range values {
		p.store.Set(path.Join("/", strings.TrimPrefix(k, p.Prefix)), v)
	}

	return nil
}

// createStageFile stages the src configuration file by processing the src
// template and setting the desired owner, group, and mode. It also sets the
// StageFile for the template resource.
// It returns an error if any.
func (p *TemplateResourceProcessor) createStageFile(opt *options) error {
	if fileNotExists(p.Src) {
		err := errors.New("Missing template: " + p.Src)
		logger.Error(err)
		return err
	}

	tmpl, err := template.New(filepath.Base(p.Src)).Funcs(template.FuncMap(p.funcMap)).ParseFiles(p.Src)
	if err != nil {
		err := fmt.Errorf("Unable to process template %s, %s", p.Src, err)
		logger.Error(err)
		return err
	}

	// create TempFile in Dest directory to avoid cross-filesystem issues
	temp, err := ioutil.TempFile(filepath.Dir(p.Dest), "."+filepath.Base(p.Dest))
	if err != nil {
		logger.Error(err)
		return err
	}

	if err = tmpl.Execute(temp, nil); err != nil {
		temp.Close()
		os.Remove(temp.Name())
		logger.Error(err)
		return err
	}
	defer temp.Close()

	// Set the owner, group, and mode on the stage file now to make it easier to
	// compare against the destination configuration file later.
	os.Chmod(temp.Name(), p.FileMode)
	os.Chown(temp.Name(), p.Uid, p.Gid)

	p.stageFile = temp
	return nil
}

// sync compares the staged and dest config files and attempts to sync them
// if they differ. sync will run a config check command if set before
// overwriting the target config file. Finally, sync will run a reload command
// if set to have the application or service pick up the changes.
// It returns an error if any.
func (p *TemplateResourceProcessor) sync(opt *options) error {
	staged := p.stageFile.Name()

	if p.keepStageFile {
		logger.Info("Keeping staged file: " + staged)
	} else {
		defer os.Remove(staged)
	}

	logger.Debug("Comparing candidate config to " + p.Dest)

	isSame, err := p.checkSameConfig(staged, p.Dest)
	if err != nil {
		logger.Warning(err)
	}

	if p.noop {
		logger.Warning("Noop mode enabled. " + p.Dest + " will not be modified")
		return nil
	}
	if isSame {
		logger.Debug("Target config " + p.Dest + " in sync")
		return nil
	}

	logger.Info("Target config " + p.Dest + " out of sync")
	if !p.syncOnly && strings.TrimSpace(p.CheckCmd) != "" {
		// TODO: support hook
		if err := p.doCheckCmd(); err != nil {
			return fmt.Errorf("Config check failed: %v", err)
		}
	}

	logger.Debug("Overwriting target config " + p.Dest)

	err = os.Rename(staged, p.Dest)
	if err != nil {
		logger.Debug("Rename failed - target is likely a mount. Trying to write instead")

		if !strings.Contains(err.Error(), "device or resource busy") {
			return err
		}

		// try to open the file and write to it

		var contents []byte
		var rerr error
		contents, rerr = ioutil.ReadFile(staged)
		if rerr != nil {
			return rerr
		}

		err := ioutil.WriteFile(p.Dest, contents, p.FileMode)
		// make sure owner and group match the temp file, in case the file was created with WriteFile
		os.Chown(p.Dest, p.Uid, p.Gid)
		if err != nil {
			return err
		}
	}

	if !p.syncOnly && strings.TrimSpace(p.ReloadCmd) != "" {
		// TODO: support hook
		if err := p.doReloadCmd(); err != nil {
			return err
		}
	}

	logger.Info("Target config " + p.Dest + " has been updated")
	return nil
}

// check executes the check command to validate the staged config file. The
// command is modified so that any references to src template are substituted
// with a string representing the full path of the staged file. This allows the
// check to be run on the staged file before overwriting the destination config
// file.
// It returns nil if the check command returns 0 and there are no other errors.
func (p *TemplateResourceProcessor) doCheckCmd() error {
	var cmdBuffer bytes.Buffer
	data := make(map[string]string)
	data["src"] = p.stageFile.Name()
	tmpl, err := template.New("checkcmd").Parse(p.CheckCmd)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(&cmdBuffer, data); err != nil {
		return err
	}
	return p.runCommand(cmdBuffer.String())
}

// reload executes the reload command.
// It returns nil if the reload command returns 0.
func (p *TemplateResourceProcessor) doReloadCmd() error {
	return p.runCommand(p.ReloadCmd)
}

// runCommand is a shared function used by check and reload
// to run the given command and log its output.
// It returns nil if the given cmd returns 0.
// The command can be run on unix and windows.
func (_ *TemplateResourceProcessor) runCommand(cmd string) error {
	cmd = strings.TrimSpace(cmd)

	logger.Debug("TemplateResourceProcessor.runCommand: " + cmd)

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/C", cmd)
	} else {
		c = exec.Command("/bin/sh", "-c", cmd)
	}

	output, err := c.CombinedOutput()
	if err != nil {
		logger.Errorf("%q", string(output))
		return err
	}

	logger.Debugf("%q", string(output))
	return nil
}

// checkSameConfig reports whether src and dest config files are equal.
// Two config files are equal when they have the same file contents and
// Unix permissions. The owner, group, and mode must match.
// It return false in other cases.
func (_ *TemplateResourceProcessor) checkSameConfig(src, dest string) (bool, error) {
	d, err := readFileStat(dest)
	if err != nil {
		return false, err
	}
	s, err := readFileStat(src)
	if err != nil {
		return false, err
	}

	if d.Uid != s.Uid {
		return false, fmt.Errorf("%s has UID %d should be %d", dest, d.Uid, s.Uid)
	}
	if d.Gid != s.Gid {
		return false, fmt.Errorf("%s has GID %d should be %d", dest, d.Gid, s.Gid)
	}
	if d.Mode != s.Mode {
		return false, fmt.Errorf("%s has mode %s should be %s", dest, os.FileMode(d.Mode), os.FileMode(s.Mode))
	}
	if d.Md5 != s.Md5 {
		return false, fmt.Errorf("%s has md5sum %s should be %s", dest, d.Md5, s.Md5)
	}

	return true, nil
}
