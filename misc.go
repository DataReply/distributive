package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// command runs a shell command, and collapses its error code to 0 or 1.
// It outputs stderr and stdout if the command has error code != 0.
func command(parameters []string) (exitCode int, exitMessage string) {
	toExec := parameters[0]
	params := strings.Split(toExec, " ")
	out, err := exec.Command(params[0], params[1:]...).CombinedOutput()
	if strings.Contains(err.Error(), "not found in $PATH") {
		return 1, "Executable not found: " + params[0]
	}
	if err == nil {
		return 0, ""
	}
	// Create output message
	exitMessage += "Command exited with non-zero exit code:"
	exitMessage += "\n\tCommand: " + toExec
	exitMessage += "\n\tExit code: " + fmt.Sprint(exitCode)
	exitMessage += "\n\tOutput: " + fmt.Sprint(out)
	return 1, exitMessage
}

// running checks if a process is running using `ps aux`, and searching for the
// process name, excluding this process (in case the process name is in the JSON
// file name)
func running(parameters []string) (exitCode int, exitMessage string) {
	// getRunningCommands returns the entries in the "COMMAND" column of `ps aux`
	getRunningCommands := func() (commands []string) {
		cmd := exec.Command("ps", "aux")
		return commandColumnNoHeader(10, cmd)
	}
	proc := parameters[0]
	// remove this process from consideration
	commands := getRunningCommands()
	var filtered []string
	for _, cmd := range commands {
		if !strings.Contains(cmd, "distributive") {
			filtered = append(filtered, cmd)
		}
	}
	if strIn(proc, filtered) {
		return 0, ""
	}
	return genericError("Process not running", proc, filtered)
}

// temp parses the output of lm_sensors and determines if Core 0 (all cores) are
// over a certain threshold as specified in the JSON.
func temp(parameters []string) (exitCode int, exitMessage string) {
	// getCoreTemp returns an integer temperature for a certain core
	getCoreTemp := func(core int) (temp int) {
		out, err := exec.Command("sensors").Output()
		if err != nil {
			log.Fatal("Error while executing `sensors`:\n\t" + err.Error())
		}
		// get all-core line up to paren
		lineRegex := regexp.MustCompile("Core " + fmt.Sprint(core) + ":?(.*)\\(")
		line := lineRegex.Find(out)
		// get temp from that line
		tempRegex := regexp.MustCompile("\\d+\\.\\d*")
		tempString := string(tempRegex.Find(line))
		tempFloat, err := strconv.ParseFloat(tempString, 64)
		if err != nil {
			msg := "Error while parsing output from `sensors`:\n\t"
			log.Fatal(msg + err.Error())
		}
		return int(tempFloat)

	}
	max := parseMyInt(parameters[0])
	temp := getCoreTemp(0)
	if temp < max {
		return 0, ""
	}
	msg := "Core temp exceeds defined maximum"
	return genericError(msg, fmt.Sprint(max), []string{fmt.Sprint(temp)})
}

// module checks to see if a kernel module is installed
func module(parameters []string) (exitCode int, exitMessage string) {
	// kernelModules returns a list of all modules that are currently loaded
	kernelModules := func() (modules []string) {
		cmd := exec.Command("/sbin/lsmod")
		return commandColumnNoHeader(0, cmd)
	}
	name := parameters[0]
	modules := kernelModules()
	if strIn(name, modules) {
		return 0, ""
	}
	return genericError("Module is not loaded", name, modules)
}

// kernelParameter checks to see if a kernel parameter was set
func kernelParameter(parameters []string) (exitCode int, exitMessage string) {
	// parameterValue returns the value of a kernel parameter
	parameterSet := func(name string) bool {
		_, err := exec.Command("/sbin/sysctl", "-q", "-n", name).Output()
		// failed on incorrect module name
		if err != nil && strings.Contains(err.Error(), "255") {
			return false
		} else if err != nil {
			log.Fatal("Error while executing /sbin/systctl:\n\tError: " + err.Error())
		}
		return true
	}
	name := parameters[0]
	if parameterSet(name) {
		return 0, ""
	}
	return 1, "Kernel parameter not set: " + name
}

// phpConfig checks the value of a PHP configuration variable
func phpConfig(parameters []string) (exitCode int, exitMessage string) {
	// getPHPVariable returns the value of a PHP configuration value as a string
	// or just "" if it doesn't exist
	getPHPVariable := func(name string) (val string) {
		quote := func(str string) string {
			return "\"" + str + "\""
		}
		// php -r 'echo get_cfg_var("default_mimetype");
		echo := fmt.Sprintf("echo get_cfg_var(%s);", quote(name))
		cmd := exec.Command("php", "-r", echo)
		out, err := cmd.CombinedOutput()
		if err != nil {
			msg := "Couldn't execute command:"
			msg += "\n\tPath: " + cmd.Path
			msg += "\n\tCommand: php -r " + echo
			msg += "\n\tOutput: " + string(out)
			msg += "\n\tError: " + err.Error()
			log.Fatal(msg)
		}
		return string(out)
	}
	name := parameters[0]
	value := parameters[1]
	actualValue := getPHPVariable(name)
	if actualValue == value {
		return 0, ""
	} else if actualValue == "" {
		msg := "PHP configuration variable not set"
		return genericError(msg, value, []string{actualValue})
	}
	msg := "PHP variable did not match expected value"
	return genericError(msg, value, []string{actualValue})
}
