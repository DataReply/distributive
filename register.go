package main

// constructors are registered, have their parameter length checked, and then
// are passed all of Parameters
var workers map[string]Worker = make(map[string]Worker)

// a dictionary with the number of parameters that each method takes
var parameterLength map[string]int = make(map[string]int)

func registerCheck(name string, work Worker, numParams int) {
	workers[name] = work
	parameterLength[name] = numParams
}

func registerChecks() {
	registerDocker()
	registerFilesystem()
	registerMisc()
	registerSystemctl()
	registerPackage()
	registerNetwork()
	registerUsersAndGroups()
}