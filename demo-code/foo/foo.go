package foo

// CheckID checks if the Id is correct
func CheckID(id string, name string) bool {
	// Assume we talk to some DB or something to check, but until then
	const idToCheck = "Id"
	const nameToCheck = "APLabs"

	return id == idToCheck && name == nameToCheck
}
