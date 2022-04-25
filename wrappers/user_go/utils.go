package user_go

func getFirstAndLastNameWithPrivacy(status NamePrivacyStatus, firstName string, lastName string, userName string) (string, string) {
	if status == NamePrivacyStatusAllHidden || status == NamePrivacyStatusFirstNameHidden ||
		status == NamePrivacyStatusLastNameHidden || (len(firstName) == 0 && len(lastName) == 0) {
		return formatUserName(userName), ""
	}

	return firstName, lastName
}

func formatUserName(userName string) string {
	return userName
}
