package ddevapp

import (
	"fmt"
	"io/ioutil"

	"os"
	"path/filepath"

	"github.com/drud/ddev/pkg/fileutil"
	"github.com/drud/ddev/pkg/output"
	"github.com/drud/ddev/pkg/util"
)

const typo3AdditionalConfigTemplate = `<?php
/** ` + DdevFileSignature + `: Automatically generated TYPO3 AdditionalConfiguration.php file.
 ddev manages this file and may delete or overwrite the file unless this comment is removed.
 */
 
$GLOBALS['TYPO3_CONF_VARS']['SYS']['trustedHostsPattern'] = '.*';

$GLOBALS['TYPO3_CONF_VARS']['DB']['Connections']['Default'] = array_merge($GLOBALS['TYPO3_CONF_VARS']['DB']['Connections']['Default'], [
                    'dbname' => 'db',
                    'host' => 'db',
                    'password' => 'db',
                    'port' => '3306',
                    'user' => 'db',
]);`

// createTypo3SettingsFile creates the app's LocalConfiguration.php and
// AdditionalConfiguration.php, adding things like database host, name, and
// password. Returns the fullpath to settings file and error
func createTypo3SettingsFile(app *DdevApp) (string, error) {

	if !fileutil.FileExists(app.SiteSettingsPath) {
		util.Warning("TYPO3 does not seem to have been set up yet, missing LocalConfiguration.php (%s)", app.SiteLocalSettingsPath)
	}

	settingsFilePath, err := app.DetermineSettingsPathLocation()
	if err != nil {
		return "", fmt.Errorf("Failed to get TYPO3 AdditionalConfiguration.php file path: %v", err.Error())
	}
	output.UserOut.Printf("Generating %s file for database connection.", filepath.Base(settingsFilePath))

	err = writeTypo3SettingsFile(app)
	if err != nil {
		return settingsFilePath, fmt.Errorf("Failed to write TYPO3 AdditionalConfiguration.php file: %v", err.Error())
	}

	return settingsFilePath, nil
}

// writeTypo3SettingsFile produces AdditionalConfiguration.php file
// It's assumed that the LocalConfiguration.php already exists, and we're
// overriding the db config values in it. The typo3conf/ directory will
// be created if it does not yet exist.
func writeTypo3SettingsFile(app *DdevApp) error {

	filePath := app.SiteLocalSettingsPath

	// Ensure target directory is writable.
	dir := filepath.Dir(filePath)
	var perms os.FileMode = 0755
	if err := os.Chmod(dir, perms); err != nil {
		if !os.IsNotExist(err) {
			// The directory exists, but chmod failed.
			return err
		}

		// The directory doesn't exist, create it with the appropriate permissions.
		if err := os.Mkdir(dir, perms); err != nil {
			return err
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	contents := []byte(typo3AdditionalConfigTemplate)
	err = ioutil.WriteFile(filePath, contents, 0644)
	if err != nil {
		return err
	}
	util.CheckClose(file)
	return nil
}

// getTypo3UploadDir just returns a static upload files (public files) dir.
// This can be made more sophisticated in the future, for example by adding
// the directory to the ddev config.yaml.
func getTypo3UploadDir(app *DdevApp) string {
	// @todo: Check to see if this gets overridden in LocalConfiguration.php
	return "uploads"
}

// Typo3Hooks adds a TYPO3-specific hooks example for post-import-db
const Typo3Hooks = `
#  post-start:
#    - exec: "composer install -d /var/www/html"`

// getTypo3Hooks for appending as byte array
func getTypo3Hooks() []byte {
	// We don't have anything new to add yet.
	return []byte(Typo3Hooks)
}

// setTypo3SiteSettingsPaths sets the paths to settings.php/settings.local.php
// for templating.
func setTypo3SiteSettingsPaths(app *DdevApp) {
	settingsFileBasePath := filepath.Join(app.AppRoot, app.Docroot)
	var settingsFilePath, localSettingsFilePath string
	settingsFilePath = filepath.Join(settingsFileBasePath, "typo3conf", "LocalConfiguration.php")
	localSettingsFilePath = filepath.Join(settingsFileBasePath, "typo3conf", "AdditionalConfiguration.php")
	app.SiteSettingsPath = settingsFilePath
	app.SiteLocalSettingsPath = localSettingsFilePath
}

// isTypoApp returns true if the app is of type typo3
func isTypo3App(app *DdevApp) bool {
	if _, err := os.Stat(filepath.Join(app.AppRoot, app.Docroot, "typo3")); err == nil {
		return true
	}
	return false
}
