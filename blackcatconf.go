// ┌──────────────────────────────────┐
// │ Marius 'f0wL' Genheimer, 2021    │
// └──────────────────────────────────┘

package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
)

// check errors as they occur and panic :o
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// scanFile searches a byte array for a byte pattern; if found it returns the
//postition of the pattern. If it found nothing it will return -1
func scanFile(data []byte, search []byte) (int, error) {
	return bytes.Index(data, search), nil
}

// calcSHA256 reads the sample file and calculates its SHA-256 hashsum
func calcSHA256(file string) string {
	f, readErr := os.Open(file)
	check(readErr)
	defer f.Close()

	h := sha256.New()
	_, hashErr := io.Copy(h, f)
	check(hashErr)
	return hex.EncodeToString(h.Sum(nil))
}

// calcMD5 reads the sample file and calculates its MD5 hashsum
func calcMD5(file string) string {

	f, readErr := os.Open(file)
	check(readErr)
	defer f.Close()

	h := md5.New()
	_, hashErr := io.Copy(h, f)
	check(hashErr)
	return hex.EncodeToString(h.Sum(nil))
}

// getFileInfo returns the size on disk of the specified file
func getFileInfo(file string) int64 {
	f, readErr := os.Open(file)
	check(readErr)
	defer f.Close()

	fileInfo, fileErr := f.Stat()
	check(fileErr)

	return fileInfo.Size()
}

// Flag variables for commandline arguments
var jsonFlag bool

// BlackCatConfig is used to store the extracted configuration
type BlackCatConfig struct {
	ID                     string      `json:"config_id"`
	PubKey                 string      `json:"public_key"`
	Extension              string      `json:"extension"`
	RansomnoteName         string      `json:"note_file_name"`
	FullRansomnote         string      `json:"note_full_text"`
	ShortRansomnote        string      `json:"note_short_text"`
	DefaultFileMode        interface{} `json:"default_file_mode"` // can either be a string or an int array (SmartPattern)
	DefaultFileCipher      string      `json:"default_file_cipher"`
	CompromisedCredentials [][]string  `json:"credentials"`
	KillServices           []string    `json:"kill_services"`
	KillProcesses          []string    `json:"kill_processes"`
	ExcludeDirectories     []string    `json:"exclude_directory_names"`
	ExcludeFilesByName     []string    `json:"exclude_file_names"`
	ExcludeFilesByExt      []string    `json:"exclude_file_extensions"`
	ExcludeFilePathWC      []string    `json:"exclude_file_path_wildcard"`
	NetworkDiscovery       bool        `json:"enable_network_discovery"`
	SelfPropagation        bool        `json:"enable_self_propagation"`
	SetWallpaper           bool        `json:"enable_set_wallpaper"`
	ESXIVMKill             bool        `json:"enable_esxi_vm_kill"`
	ESXIVMSnapshotKill     bool        `json:"enable_esxi_vm_snapshot_kill"`
	StrictIncludePaths     []string    `json:"strict_include_paths"`
	ESXIVMKillExclude      []string    `json:"esxi_vm_kill_exclude"`
}

func main() {

	fmt.Printf("                                                              _\n")
	fmt.Printf("                                                              \\`*-. \n")
	fmt.Printf("                                                               )  _`-.  \n")
	fmt.Printf("                                                              .  : `. .   \n")
	fmt.Printf("                                                              : _   '  \\  \n")
	fmt.Printf("      ___  __         __   _____     __  _____          ___   ; *` _.   `*-._  \n")
	fmt.Printf("     / _ )/ /__ _____/ /__/ ___/__ _/ /_/ ___/__  ___  / _/   `-.-'          `-.  \n")
	fmt.Printf("    / _  / / _ `/ __/  '_/ /__/ _ `/ __/ /__/ _ \\/ _ \\/ _/      ;       `       `.  \n")
	fmt.Printf("   /____/_/\\_,_/\\__/_/\\_\\___/\\_,_/\\__/\\___/\\___/_//_/_/         :.       .        \\  \n")
	fmt.Printf("                                                                 . \\  .   :   .-'   .\n")
	fmt.Printf("   Static Configuration Extractor for BlackCat Ransomware        '  `+.;  ;  '      : \n")
	fmt.Printf("   Marius 'f0wL' Genheimer | https://dissectingmalwa.re          :  '  |    ;       ;-. \n")
	fmt.Printf("                                                                 ; '   : :`-:     _.`* ; \n")
	fmt.Printf("                                                         [bug] .*' /  .*' ; .*`- +'  `*' \n")
	fmt.Printf("                                                              `*-*   `*-*  `*-*'\n\n")

	// parse passed flags
	flag.BoolVar(&jsonFlag, "j", false, "Write extracted config to a JSON file")
	flag.Parse()

	if flag.NArg() == 0 {
		color.Red("✗ No path to sample provided.\n\n")
		os.Exit(1)
	}

	// calculate hash sums of the sample
	md5sum := calcMD5(flag.Args()[0])
	sha256sum := calcSHA256(flag.Args()[0])

	// basic file info
	w1 := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w1, "File size (bytes): \t", getFileInfo(flag.Args()[0]))
	fmt.Fprintln(w1, "Sample MD5: \t", md5sum)
	fmt.Fprintln(w1, "Sample SHA-256: \t", sha256sum)
	w1.Flush()

	// read the contents of the BlackCat sample
	sample, readErr := ioutil.ReadFile(flag.Args()[0])
	check(readErr)

	// offset: start of the config
	offBytes, byteErr := hex.DecodeString("7B22636F6E6669675F696422") //  = {"config_id"
	check(byteErr)
	off, scanErr := scanFile(sample, offBytes)
	check(scanErr)

	if off == -1 {
		color.Red("\n✗ Unable to find config offset.\n\n")
		os.Exit(1)
	}

	// carve out the config
	cfgRaw := sample[off : off+8000]

	// trim the superfluous nullbytes from the end of the decrypted config
	cfg := bytes.Trim(cfgRaw, "\x20")

	// if blackCatConf is run with -j the configuration will be written to disk in a JSON file
	if jsonFlag {

		// assemble file name
		filename := "blackCat_config-" + md5sum + ".json"

		// write the JSON string to a file
		jsonOutput, createErr := os.Create(filename)
		check(createErr)
		defer jsonOutput.Close()
		n3, writeErr := jsonOutput.WriteString(string(cfg))
		check(writeErr)
		color.Green("\n✓ Wrote %d bytes to %v\n\n", n3, filename)
		jsonOutput.Sync()
	}

	// unmarshal json string into the BlackCatConf struct
	var config BlackCatConfig
	jsonErr := json.Unmarshal(cfg, &config)
	check(jsonErr)

	// print configuration contents
	fmt.Fprintln(w1, "Config ID: \t", config.ID)
	fmt.Fprintln(w1, "Public Key: \t", config.PubKey)
	fmt.Fprintln(w1, "File Extension: \t", config.Extension)
	fmt.Fprintln(w1, "Ransomnote Filename: \t", config.RansomnoteName)
	fmt.Fprintln(w1, "Default File Encryption Mode: \t", config.DefaultFileMode)
	fmt.Fprintln(w1, "Default File Encryption Cipher: \t", config.DefaultFileCipher)
	fmt.Fprintln(w1, "Compromised Credentials: \t", config.CompromisedCredentials)
	fmt.Fprintln(w1, "Services to be killed: \t", config.KillServices)
	fmt.Fprintln(w1, "Processes to be killed: \t", config.KillProcesses)
	fmt.Fprintln(w1, "Directories to be excluded: \t", config.ExcludeDirectories)
	fmt.Fprintln(w1, "Files to be excluded: \t", config.ExcludeFilesByName)
	fmt.Fprintln(w1, "Extensions to be excluded: \t", config.ExcludeFilesByExt)
	fmt.Fprintln(w1, "File Path Wildcards: \t", config.ExcludeFilePathWC)
	fmt.Fprintln(w1, "Network Discovery: \t", config.NetworkDiscovery)
	fmt.Fprintln(w1, "Self-Propagation: \t", config.SelfPropagation)
	fmt.Fprintln(w1, "Set Wallpaper: \t", config.SetWallpaper)
	fmt.Fprintln(w1, "ESXI VM Kill: \t", config.ESXIVMKill)
	fmt.Fprintln(w1, "ESXI Snapshot Kill: \t", config.ESXIVMSnapshotKill)
	fmt.Fprintln(w1, "Strict Include Paths: \t", config.StrictIncludePaths)
	fmt.Fprintln(w1, "ESXI VM Kill Exclude: \t", config.ESXIVMKillExclude)
	w1.Flush()

	color.Yellow("\nShort Ransomnote:\n")
	fmt.Printf("\n%v\n", config.ShortRansomnote)

	color.Yellow("\nFull Ransomnote:\n")
	fmt.Printf("\n%v\n\n", config.FullRansomnote)
}
